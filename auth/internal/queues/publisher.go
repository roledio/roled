package queues

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	PublishDLQ(ctx context.Context, msg Message) error
}
