package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/ilivestrong/email_warmup_service/internal/config"
	"github.com/ilivestrong/email_warmup_service/internal/processor"
	"github.com/ilivestrong/email_warmup_service/internal/providers"
	"github.com/ilivestrong/email_warmup_service/internal/queue"
	"github.com/ilivestrong/email_warmup_service/internal/quota"
	"github.com/ilivestrong/email_warmup_service/internal/scheduler"
	"github.com/ilivestrong/email_warmup_service/internal/validator"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize queue client
	qClient, err := queue.NewClient(cfg.QueueURL)
	if err != nil {
		log.Fatalf("queue init error: %v", err)
	}
	fmt.Println("email queue connected")

	quotaStore, err := quota.NewStore(cfg.RedisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}

	emailValidator := validator.New(cfg.Validator.DisposableDomains)
	provFactory := providers.NewFactory(cfg.ProviderMap, cfg.SMTP, cfg.GoogleOAuth)

	processor := processor.New(quotaStore, emailValidator, provFactory, qClient, cfg.RetryPolicy)
	for i := 0; i < cfg.WorkerCount; i++ {
		go processor.Start(ctx)
	}

	sched := scheduler.NewScheduler(cfg, quotaStore, provFactory)
	go sched.StartDaily(ctx)

	<-ctx.Done()
	log.Println("shutting down the service")

}
