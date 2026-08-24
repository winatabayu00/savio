package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/savio/savio/backend/internal/platform/database"
	"github.com/savio/savio/backend/internal/worker"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://savio:savio@localhost:5433/savio?sslmode=disable"
	}
	db, err := database.Connect(dsn)
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("worker started (interval %s)", interval)
	for {
		run := context.WithoutCancel(ctx)
		now := time.Now().UTC()
		if err := svc.AutoPostDue(run, now); err != nil {
			log.Printf("worker autopost: %v", err)
		}
		if err := svc.SweepNotifications(run, now); err != nil {
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