package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/savio/savio/backend/internal/ai"
	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/database"
	"github.com/savio/savio/backend/internal/telegram"
	"github.com/savio/savio/backend/internal/transactions"
	"github.com/savio/savio/backend/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	interval := time.Minute
	if v := os.Getenv("WORKER_INTERVAL_SECONDS"); v != "" {
		if d, err := time.ParseDuration(v + "s"); err == nil {
			interval = d
		}
	}

	svc := worker.NewService(db)
	// Telegram poller reuses the shared AI and transaction services instead of
	// duplicating domain behavior (AGENTS #102).
	tg := telegram.NewService(db, ai.NewService(db, cfg), transactions.NewService(db))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go telegramPoll(ctx, tg)

	log.Printf("worker started (interval %s)", interval)
	for {
		now := time.Now().UTC()
		if err := svc.AutoPostDue(ctx, now); err != nil {
			log.Printf("worker autopost: %v", err)
		}
		if err := svc.SweepNotifications(ctx, now); err != nil {
			log.Printf("worker notifications: %v", err)
		}
		select {
		case <-ctx.Done():
			log.Printf("worker stopping")
			return
		case <-time.After(interval):
		}
	}
}

// telegramPoll long-polls Telegram (~25s per request) in its own goroutine so
// the minute-scale job loop never blocks recap ingestion.
func telegramPoll(ctx context.Context, tg *telegram.Service) {
	const tick = 2 * time.Second
	for {
		if err := tg.Poll(ctx); err != nil {
			log.Printf("telegram poll: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(tick):
		}
	}
}
