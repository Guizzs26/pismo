package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Guizzs26/pismo/internal/config"
	db "github.com/Guizzs26/pismo/internal/infra/database"
	"github.com/Guizzs26/pismo/internal/middleware"
	"github.com/Guizzs26/pismo/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	logger.Setup(cfg.AppEnv, cfg.LogLevel)
	slog.Info("logger initialized", "env", cfg.AppEnv, "level", cfg.LogLevel)

	deps, err := initDependencies(cfg)
	if err != nil {
		slog.Error("failed to initialize dependencies", "error", err)
		os.Exit(1)
	}
	defer deps.cleanup()

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      setupRoutes(deps),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutdown signal received", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("shutting down server gracefully", "timeout", "30s")
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("forced shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

type dependencies struct {
	pool    *pgxpool.Pool
	cleanup func()
}

func initDependencies(cfg *config.Config) (*dependencies, error) {
	ctx := context.Background()

	slog.Info("connecting to postgres database", "host", cfg.DB.Host, "name", cfg.DB.Name)
	pool, err := db.NewPostgresPool(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("database connection: %v", err)
	}
	slog.Info("database connection pool established",
		"max_conns", cfg.DB.MaxConns,
		"min_conns", cfg.DB.MinConns,
	)

	return &dependencies{
		pool: pool,
		cleanup: func() {
			slog.Info("closing database connection pool")
			pool.Close()
		},
	}, nil
}

func setupRoutes(deps *dependencies) http.Handler {
	mux := http.NewServeMux()

	// accountRepo := pg.NewAccountRepository(deps.pool)
	// accountService := service.NewAccountService(accountRepo)
	// accountHandler := handler.NewAccountHandler(accountService)

	// mux.HandleFunc("POST /accounts", accountHandler.Create)
	// mux.HandleFunc("GET /accounts/{accountId}", accountHandler.FindByID)

	return middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.RequestID,
		middleware.Logging,
	)
}
