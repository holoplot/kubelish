package publisher

// Protocol is a bit mask of the IP protocols a service is announced on.
type Protocol int

const (
	ProtocolIPv4 = Protocol(1 << iota)
	ProtocolIPv6
)

func (p Protocol) String() string {
	switch p {
	case ProtocolIPv4:
		return "IPv4"
	case ProtocolIPv6:
		return "IPv6"
	case ProtocolIPv4 | ProtocolIPv6:
		return "IPv4+IPv6"
	default:
		return "none"
	}
}

type Publisher interface {
	Publish(serviceName, serviceType, txt string, protocols Protocol, port int) (PublishedService, error)
}

type PublishedService interface {
	Close()
}
