package publisher

type Publisher interface {
	Publish(serviceName, serviceType, txt string, port int) (PublishedService, error)
}

type PublishedService interface {
	Close()
}
