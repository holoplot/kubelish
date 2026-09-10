//go:build linux

package avahipublisher

import (
	"fmt"
	"log/slog"
	"strings"

	dbus "github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
	"github.com/holoplot/kubelish/pkg/publisher"
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

// avahiProtocol maps a protocol mask onto the single protocol value that
// AddService() expects. Avahi's protocol argument is an enum, not a bit
// mask, so announcing on both protocols means leaving it unspecified.
func avahiProtocol(protocols publisher.Protocol) (int32, error) {
	switch protocols {
	case publisher.ProtocolIPv4:
		return avahi.ProtoInet, nil
	case publisher.ProtocolIPv6:
		return avahi.ProtoInet6, nil
	case publisher.ProtocolIPv4 | publisher.ProtocolIPv6:
		return avahi.ProtoUnspec, nil
	default:
		return 0, fmt.Errorf("no protocol to announce on")
	}
}

func (a *AvahiPublisher) Publish(serviceName, serviceType, txt string, protocols publisher.Protocol, port int) (publisher.PublishedService, error) {
	protocol, err := avahiProtocol(protocols)
	if err != nil {
		return nil, err
	}

	eg, err := a.avahiServer.EntryGroupNew()
	if err != nil {
		return nil, fmt.Errorf("EntryGroupNew() failed: %w", err)
	}

	txtBytes := [][]byte{}

	if txt != "" {
		txtBytes = [][]byte{[]byte(txt)}
	}

	localName := strings.Join([]string{serviceName, "on", a.hostnameFqdn}, " ")

	if err := eg.AddService(avahi.InterfaceUnspec, protocol, 0, localName,
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

	slog.Info("Starting Avahi publisher", "hostname", hostname)

	return &AvahiPublisher{
		avahiServer:  avahiServer,
		hostnameFqdn: hostname,
	}, nil
}
