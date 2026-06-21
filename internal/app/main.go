package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-template/internal/config"
	"go-template/internal/jwtauth"
	"go-template/internal/repository"
	"go-template/internal/service"
	thttp "go-template/internal/transport/http"
)

type App struct {
	cfg    *config.Config
	pool   *pgxpool.Pool
	server *thttp.Server
	logger *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	pool, err := pgxpool.New(ctx, cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}

	ttl, err := time.ParseDuration(cfg.JWT.TTL)
	if err != nil {
		return nil, fmt.Errorf("jwt.ttl: %w", err)
	}
	jm := jwtauth.New(cfg.JWT.Secret, ttl)

	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo)

	// example: emailSender := email.New(logger) — pass to services that need it

	server := thttp.NewServer(userSvc, jm, logger)

	return &App{cfg: cfg, pool: pool, server: server, logger: logger}, nil
}

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:         a.cfg.HTTP.Addr,
		Handler:      a.server,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		a.logger.Info("shutting down")
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() {
	a.pool.Close()
}
