package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gofiber/fiber/v3/log"
	"github.com/roledio/roled/internal/configs"
	"github.com/roledio/roled/internal/constants"
	"github.com/roledio/roled/internal/queues"
	"github.com/roledio/roled/internal/services/infra"
)

func main() {
	config, err := configs.LoadDefaultConfig("./configs/default.yml")
	if err != nil {
		log.Fatal("Failed to load configs: ", err)
	}

	app, err := NewApp(config)
	if err != nil {
		log.Fatal("Failed to initialize app: ", err)
	}
	defer app.Close()

	// Create a context that we can cancel. This will be used to signal the worker to
	// stop when we are shutting down the server.
	ctx, cancel := context.WithCancel(context.Background())

	// Ensure the cancel() is called in case of any unexpected exit before receiving
	// the shutdown signal, such as a panic or an error. It is idempotent, so it's safe
	// to call even if the worker has already stopped.
	defer cancel()

	registerQueues(config, app)

	var wg sync.WaitGroup

	// start workers
	startQueueWorkers(ctx, app.redisService, &wg)

	// Channel to listen for interrupt or terminate signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	go func() {
		if err := app.Run(); err != nil {
			log.Error("Server error: ", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Info("Shutting down server...")

	// Signal the worker to stop by canceling the context
	cancel()

	// Attempt graceful shutdown
	if err := app.Shutdown(); err != nil {
		log.Error("Server forced to shutdown: ", err)
	} else {
		log.Info("Server shutdown gracefully")
	}

	// Wait for all workers to finish processing
	wg.Wait()

	log.Info("All workers stopped, exiting.")
}

func registerQueues(defaultConfig *configs.DefaultConfig, app *App) {
	emailQueue := fmt.Sprintf("%s:%s", defaultConfig.Redis.Prefix, constants.QueueEmail)
	queues.Register(emailQueue, app.queuePublishers.EmailPublisher, app.queueHandlers.EmailHandler)
}

func startQueueWorkers(ctx context.Context, redisService infra.RedisService, wg *sync.WaitGroup) {
	configs := []queues.WorkerConfig{
		{
			Stream:       constants.QueueEmail,
			Group:        "email-workers",
			Consumer:     "email-worker",
			RetryEnabled: true,
			MaxRetry:     3,
			DLQStream:    constants.QueueEmailDLQ,
		},
	}
	for _, cfg := range configs {
		wg.Add(1)
		go func(cfg queues.WorkerConfig) {
			defer wg.Done()
			log.Infow("Starting worker", "stream", cfg.Stream, "group", cfg.Group, "consumer", cfg.Consumer)
			queues.StartWorker(ctx, redisService, cfg)
		}(cfg)
	}
}
