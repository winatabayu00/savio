package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/savio/savio/backend/internal/platform/config"
	"github.com/savio/savio/backend/internal/platform/database"
	"github.com/savio/savio/backend/internal/platform/redisclient"
	"github.com/savio/savio/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err.Error())
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err.Error())
		os.Exit(1)
	}

	rdb, err := redisclient.Connect(cfg.RedisURL)
	if err != nil {
		slog.Warn("redis connection failed; continuing without redis", "error", err.Error())
		rdb = nil
	}

	app := server.New(cfg, db, rdb)
	srv := app.Server()

	go func() {
		slog.Info("savio api listening", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("api process exited", "error", err.Error())
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "error", err.Error())
	}
}
