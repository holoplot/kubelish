//go:build linux
// +build linux

package avahipublisher

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
	"github.com/holoplot/kubelish/pkg/publisher"
	"github.com/rs/zerolog/log"
)

type AvahiPublisher struct {
	avahiServer  *avahi.Server
	hostnameFqdn string
}

type PublishedAvahiService struct {
	publisher  *AvahiPublisher
	entryGroup *avahi.EntryGroup
}

func (p *PublishedAvahiService) Close() {
	p.publisher.avahiServer.EntryGroupFree(p.entryGroup)
}

func (a *AvahiPublisher) Publish(serviceName, serviceType, txt string, port int) (publisher.PublishedService, error) {
	eg, err := a.avahiServer.EntryGroupNew()
	if err != nil {
		return nil, fmt.Errorf("EntryGroupNew() failed: %w", err)
	}

	txtBytes := [][]byte{}

	if txt != "" {
		txtBytes = [][]byte{[]byte(txt)}
	}

	if err := eg.AddService(avahi.InterfaceUnspec, avahi.ProtoUnspec, 0, serviceName,
		serviceType, "local", a.hostnameFqdn, uint16(port), txtBytes); err != nil {
		return nil, fmt.Errorf("AddService() failed: %w", err)
	}

	if err := eg.Commit(); err != nil {
		return nil, fmt.Errorf("commit() failed: %w", err)
	}

	return &PublishedAvahiService{publisher: a, entryGroup: eg}, nil
}

func New() (*AvahiPublisher, error) {
	dbusConn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	avahiServer, err := avahi.ServerNew(dbusConn)
	if err != nil {
		return nil, fmt.Errorf("avahi.ServerNew() failed: %w", err)
	}

	hostname, err := avahiServer.GetHostNameFqdn()
	if err != nil {
		return nil, fmt.Errorf("GetHostNameFqdn() failed: %w", err)
	}

	log.Info().Str("hostname", hostname).Msg("Starting Avahi publisher")

	return &AvahiPublisher{
		avahiServer:  avahiServer,
		hostnameFqdn: hostname,
	}, nil
}
