package queues

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/roledio/roled/auth/internal/services/infra"
	"github.com/roledio/roled/auth/pkg/utils/jsonutil"
)

type WorkerConfig struct {
	Stream       string
	Group        string
	Consumer     string
	RetryEnabled bool
	MaxRetry     int
	DLQStream    string
}

func StartWorker(ctx context.Context, redisService infra.RedisService, cfg WorkerConfig) {
	// Override stream names with prefix
	cfg.Stream = redisService.KeyWithPrefix(cfg.Stream)
	cfg.DLQStream = redisService.KeyWithPrefix(cfg.DLQStream)

	// Override consumer name with unique name
	cfg.Consumer = buildConsumerName(cfg.Consumer)

	rdb := redisService.Client()
	initGroup(rdb, cfg.Stream, cfg.Group)

	for {
		// Check if context is done to allow graceful shutdown
		select {
		case <-ctx.Done():
			log.Infow("Worker shutting down", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)
			return
		default:
			// Continue processing
		}

		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    cfg.Group,
			Consumer: cfg.Consumer,
			Streams:  []string{cfg.Stream, ">"},
			Count:    10, // Read up to 10 messages at a time
			Block:    -1, // Non-blocking read (-1), returns immediately if no messages are available
		}).Result()

		if err != nil && err != redis.Nil {
			if isTimeoutErr(err) {
				// Timeout is expected when no messages are available, just continue to the next loop iteration
				log.Debugw("No messages available, retrying...", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)
				continue
			}
			log.Errorw("Failed to read messages", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer, "error", err)
			continue
		}

		// When using a non-blocking read, delay the loop if no messages are returned
		// to avoid a tight loop that consumes CPU when the stream is empty.
		if len(streams) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		for _, s := range streams {
			for _, msg := range s.Messages {
				processMessage(ctx, rdb, cfg, msg)
			}
		}
	}
}

func initGroup(rdb redis.UniversalClient, stream, group string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rdb.XGroupCreateMkStream(ctx, stream, group, "$").Err()

	if err != nil {
		log.Debugw("Failed to create group, it might already exist", "stream", stream, "group", group, "error", err)
	}
}

// buildConsumerName generates a unique consumer name using the base name, hostname, and a random UUID
// to ensure uniqueness across multiple instances of the worker.
func buildConsumerName(base string) string {
	host, _ := os.Hostname()
	return base + "-" + host + "-" + uuid.NewString()
}

func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "i/o timeout")
}

func processMessage(
	ctx context.Context,
	rdb redis.UniversalClient,
	cfg WorkerConfig,
	msg redis.XMessage,
) {
	// Create a new context for this message processing using the original context
	msgCtx := ctx

	ctxFields := msg.Values["context"].(string)
	msgCtx = injectContext(msgCtx, ctxFields)

	payload := msg.Values["payload"].(string)

	retry, err := strconv.Atoi(msg.Values["retry_count"].(string))
	if err != nil {
		log.WithContext(msgCtx).Errorw("Failed to parse retry count", "error", err, "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)
		return
	}

	handler, ok := GetHandler(cfg.Stream)
	if !ok {
		log.WithContext(msgCtx).Infow("No handler for the stream", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)
		ack(msgCtx, rdb, cfg, msg.ID)
		return
	}

	err = handler.Handle(msgCtx, payload)

	if err != nil {
		handleFailure(msgCtx, rdb, cfg, msg, payload, ctxFields, retry, err)
		return
	}

	ack(msgCtx, rdb, cfg, msg.ID)
}

func injectContext(ctx context.Context, raw string) context.Context {
	var m map[string]any
	err := jsonutil.Parse(raw, &m)
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to unmarshal context", "error", err)
		return ctx
	}

	nctx := ctx

	for k, v := range m {
		if k == "" {
			continue
		}
		// Forced to use string context key since the fiberzap's LoggerConfig.ExtraKeys cannot accept types other than string.
		// The request id will not be printed on the log if the context key is not using string.
		//nolint:staticcheck
		nctx = context.WithValue(nctx, k, v)
	}
	return nctx
}

func handleFailure(
	ctx context.Context,
	rdb redis.UniversalClient,
	cfg WorkerConfig,
	msg redis.XMessage,
	payload string,
	ctxFields string,
	retry int,
	err error,
) {

	retry++

	log.WithContext(ctx).Infow("Job failed", "retry", retry, "error", err, "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)

	if !cfg.RetryEnabled { // No retry
		ack(ctx, rdb, cfg, msg.ID)
		return
	}

	publisher, ok := GetPublisher(cfg.Stream)
	if !ok {
		log.WithContext(ctx).Infow("No publisher for the stream to process retry", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)
		ack(ctx, rdb, cfg, msg.ID)
		return
	}

	if retry < cfg.MaxRetry {
		// Requeue
		err = publisher.Publish(ctx, Message{
			Payload:    payload,
			RetryCount: retry,
			Context:    ctxFields,
		})
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to retry publish message to stream", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer, "error", err)
		}
	} else if cfg.DLQStream != "" {
		// DLQ
		err = publisher.PublishDLQ(ctx, Message{
			Payload:    payload,
			RetryCount: retry,
			Context:    ctxFields,
		})
		if err != nil {
			log.WithContext(ctx).Errorw("Failed to publish message to DLQ", "stream", cfg.DLQStream, "error", err)
		}
	}

	ack(ctx, rdb, cfg, msg.ID)
}

func ack(ctx context.Context, rdb redis.UniversalClient, cfg WorkerConfig, id string) {
	err := rdb.XAck(ctx, cfg.Stream, cfg.Group, id).Err()
	if err != nil {
		log.WithContext(ctx).Errorw("Failed to acknowledge message", "stream", cfg.Stream, "group", cfg.Group, "id", id, "error", err)
	}
	log.WithContext(ctx).Debugw("Message acknowledged", "stream", cfg.Stream, "group", cfg.Group, "id", id)
}
