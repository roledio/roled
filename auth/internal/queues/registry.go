package queues

type registry struct {
	publisher Publisher
	handler   Handler
}

var registryMap = map[string]registry{}

func Register(queue string, pulisher Publisher, handler Handler) {
	registryMap[queue] = registry{
		publisher: pulisher,
		handler:   handler,
	}
}

func GetPublisher(queue string) (Publisher, bool) {
	r, ok := registryMap[queue]
	return r.publisher, ok
}

func GetHandler(queue string) (Handler, bool) {
	r, ok := registryMap[queue]
	return r.handler, ok
}
