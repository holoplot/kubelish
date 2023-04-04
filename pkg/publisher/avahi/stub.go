//go:build !linux
// +build !linux

package avahipublisher

func New() (*AvahiPublisher, error) {
	return nil, fmt.Errorf("Avahi is not supported on this platform")
}
