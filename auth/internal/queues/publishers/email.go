package publishers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/roledio/roled/internal/queues"
	"github.com/roledio/roled/internal/services/infra"
)

type emailPublisher struct {
	rdb       redis.UniversalClient
	stream    string
	dlqStream string
}

func NewEmailPublisher(redisService infra.RedisService, stream, dlqStream string) queues.Publisher {
	return &emailPublisher{
		rdb:       redisService.Client(),
		stream:    redisService.KeyWithPrefix(stream),
		dlqStream: redisService.KeyWithPrefix(dlqStream),
	}
}

func (p *emailPublisher) Publish(ctx context.Context, msg queues.Message) error {
	return p.publishToStream(ctx, msg, p.stream)
}

func (p *emailPublisher) PublishDLQ(ctx context.Context, msg queues.Message) error {
	return p.publishToStream(ctx, msg, p.dlqStream)
}

func (p *emailPublisher) publishToStream(ctx context.Context, msg queues.Message, stream string) error {
	nctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.rdb.XAdd(nctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000, // Keep only the latest 10000 messages in this stream
		Approx: true,  // Use approximate trimming for better performance
		Values: msg.ToMap(),
	}).Err()
}
