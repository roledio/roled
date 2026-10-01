package queues_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/roledio/roled/auth/internal/queues"
	qm "github.com/roledio/roled/auth/internal/queues/mocks"
	rm "github.com/roledio/roled/auth/pkg/redis/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Redis's command hook keeps the worker tests deterministic without a network server.
type commandHook struct {
	process func(context.Context, redis.Cmder) error
}

func (h commandHook) DialHook(next redis.DialHook) redis.DialHook       { return next }
func (h commandHook) ProcessHook(_ redis.ProcessHook) redis.ProcessHook { return h.process }
func (h commandHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestWorkerDispatchAndRetry(t *testing.T) {
	for _, name := range []string{"success", "no handler", "invalid context", "invalid retry", "retry disabled", "retry", "retry error", "dead letter", "dead letter error", "no dead letter", "ack error"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cfg := queues.WorkerConfig{Stream: t.Name(), Group: "workers", Consumer: "email", RetryEnabled: true, MaxRetry: 3, DLQStream: "dead"}
			if name == "retry disabled" {
				cfg.RetryEnabled = false
			}
			if name == "no dead letter" {
				cfg.DLQStream = ""
			}
			stream := "test:" + cfg.Stream
			contextJSON, retry := `{"request_id":"request-1","":"ignored"}`, "0"
			if name == "invalid context" {
				contextJSON = "{"
			}
			if name == "invalid retry" {
				retry = "invalid"
			}
			if strings.Contains(name, "dead letter") {
				retry = "2"
			}
			handler := qm.NewMockHandler(t)
			publisher := qm.NewMockPublisher(t)
			if name != "no handler" {
				queues.Register(stream, publisher, handler)
			}
			var handlerErr error
			if strings.Contains(name, "retry") || strings.Contains(name, "dead letter") {
				handlerErr = errors.New("delivery failed")
			}
			if name != "no handler" && name != "invalid retry" {
				handler.EXPECT().Handle(mock.MatchedBy(func(c context.Context) bool {
					if name == "invalid context" {
						return c.Value("request_id") == nil
					}
					return c.Value("request_id") == "request-1" && c.Value("") == nil
				}), "payload").Return(handlerErr).Once()
			}
			if name == "retry" || name == "retry error" || name == "dead letter" || name == "dead letter error" {
				var publishErr error
				if strings.HasSuffix(name, "error") {
					publishErr = errors.New("publisher unavailable")
				}
				msg := queues.Message{Payload: "payload", Context: contextJSON, RetryCount: 1}
				if strings.HasPrefix(name, "dead letter") {
					msg.RetryCount = 3
					publisher.EXPECT().PublishDLQ(mock.Anything, msg).Return(publishErr).Once()
				} else {
					publisher.EXPECT().Publish(mock.Anything, msg).Return(publishErr).Once()
				}
			}
			ackCount, readCount := 0, 0
			client := redis.NewClient(&redis.Options{Addr: "unused:6379"})
			t.Cleanup(func() { require.NoError(t, client.Close()) })
			client.AddHook(commandHook{process: func(_ context.Context, cmd redis.Cmder) error {
				switch cmd.Name() {
				case "xgroup":
					require.Equal(t, []any{"xgroup", "create", stream, cfg.Group, "$", "mkstream"}, cmd.Args())
					cmd.(*redis.StatusCmd).SetVal("OK")
				case "xreadgroup":
					readCount++
					require.Contains(t, cmd.Args(), stream)
					require.Contains(t, cmd.Args(), cfg.Group)
					cancel()
					cmd.(*redis.XStreamSliceCmd).SetVal([]redis.XStream{{Stream: stream, Messages: []redis.XMessage{{ID: "1-0", Values: map[string]any{"context": contextJSON, "payload": "payload", "retry_count": retry}}}}})
				case "xack":
					ackCount++
					require.Equal(t, []any{"xack", stream, cfg.Group, "1-0"}, cmd.Args())
					if name == "ack error" {
						err := errors.New("ack failed")
						cmd.SetErr(err)
						return err
					}
					cmd.(*redis.IntCmd).SetVal(1)
				default:
					t.Fatalf("unexpected Redis command: %s", cmd.Name())
				}
				return nil
			}})
			service := rm.NewMockService(t)
			service.EXPECT().Client().Return(client)
			service.EXPECT().KeyWithPrefix(cfg.Stream).Return(stream)
			service.EXPECT().KeyWithPrefix(cfg.DLQStream).Return(cfg.DLQStream)
			queues.StartWorker(ctx, service, cfg)
			require.Equal(t, 1, readCount)
			if name == "invalid retry" {
				require.Zero(t, ackCount)
			} else {
				require.Equal(t, 1, ackCount)
			}
		})
	}
}

func TestWorkerReadFailuresAndCancellation(t *testing.T) {
	for _, name := range []string{"cancelled before read", "empty", "redis nil", "timeout", "connection failure", "existing group"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "cancelled before read" {
				cancel()
			}
			reads := 0
			client := redis.NewClient(&redis.Options{Addr: "unused:6379"})
			t.Cleanup(func() { require.NoError(t, client.Close()) })
			client.AddHook(commandHook{process: func(_ context.Context, cmd redis.Cmder) error {
				if cmd.Name() == "xgroup" {
					if name == "existing group" {
						err := errors.New("BUSYGROUP already exists")
						cmd.SetErr(err)
						return err
					}
					cmd.(*redis.StatusCmd).SetVal("OK")
					return nil
				}
				require.Equal(t, "xreadgroup", cmd.Name())
				reads++
				cancel()
				var err error
				switch name {
				case "redis nil":
					err = redis.Nil
				case "timeout":
					err = errors.New("i/o timeout")
				case "connection failure":
					err = errors.New("connection refused")
				}
				cmd.SetErr(err)
				return err
			}})
			service := rm.NewMockService(t)
			service.EXPECT().Client().Return(client)
			service.EXPECT().KeyWithPrefix("jobs").Return("test:jobs")
			service.EXPECT().KeyWithPrefix("").Return("")
			queues.StartWorker(ctx, service, queues.WorkerConfig{Stream: "jobs", Group: "group", Consumer: "worker"})
			if name == "cancelled before read" {
				require.Zero(t, reads)
			} else {
				require.Equal(t, 1, reads)
			}
		})
	}
}
