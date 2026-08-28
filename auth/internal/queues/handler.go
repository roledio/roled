package queues

import "context"

type Handler interface {
	Handle(ctx context.Context, payload string) error
}
