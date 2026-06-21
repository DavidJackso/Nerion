package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-template/internal/app"
	"go-template/internal/config"
	"go-template/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	l := logger.New(cfg.Log.Level, cfg.Log.Format)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx, cfg, l)
	if err != nil {
		l.Error("init failed", "err", err)
		os.Exit(1)
	}
	defer a.Close()

	l.Info("listening", "addr", cfg.HTTP.Addr)
	if err := a.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.Error("server error", "err", err)
		os.Exit(1)
	}
}
