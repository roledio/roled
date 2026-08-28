package queues

type Message struct {
	RetryCount int
	Payload    string
	Context    string
}

func (m *Message) ToMap() map[string]any {
	return map[string]any{
		"payload":     m.Payload,
		"retry_count": m.RetryCount,
		"context":     m.Context,
	}
}
