//go:build !linux

package avahipublisher

import (
	"fmt"

	"github.com/holoplot/kubelish/pkg/publisher"
)

type AvahiPublisher struct{}

func (a *AvahiPublisher) Publish(serviceName, serviceType, txt string, protocols publisher.Protocol, port int) (publisher.PublishedService, error) {
	return nil, fmt.Errorf("Avahi is not supported on this platform")
}

func New() (*AvahiPublisher, error) {
	return nil, fmt.Errorf("Avahi is not supported on this platform")
}
