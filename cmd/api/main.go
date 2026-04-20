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

	_ "github.com/Guizzs26/pismo/docs"
	"github.com/Guizzs26/pismo/internal/config"
	"github.com/Guizzs26/pismo/internal/handler"
	db "github.com/Guizzs26/pismo/internal/infra/database"
	"github.com/Guizzs26/pismo/internal/middleware"
	pg "github.com/Guizzs26/pismo/internal/repository/postgres"
	"github.com/Guizzs26/pismo/internal/service"
	"github.com/Guizzs26/pismo/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Pismo Challenge API
// @version         1.0
// @description     REST API for customer account and transaction management.
// @host            localhost:8080
// @BasePath        /
func main() {
	cfg := config.Load()

	logger.Setup(cfg.AppEnv, cfg.LogLevel)
	slog.Info("logger initialized", "env", cfg.AppEnv, "log_level", cfg.LogLevel)

	deps, err := initDependencies(cfg)
	if err != nil {
		slog.Error("failed to initialize dependencies", "error", err)
		os.Exit(1)
	}
	defer deps.cleanup()

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      setupRoutes(deps, cfg),
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

func setupRoutes(deps *dependencies, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	accountRepo := pg.NewAccountRepository(deps.pool)
	accountService := service.NewAccountService(accountRepo)
	accountHandler := handler.NewAccountHandler(accountService)

	transactionRepo := pg.NewTransactionRepository(deps.pool)
	transactionService := service.NewTransactionService(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	mux.HandleFunc("POST /accounts", accountHandler.Create)
	mux.HandleFunc("GET /accounts/{accountId}", accountHandler.FindByID)
	mux.HandleFunc("POST /transactions", transactionHandler.Create)

	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:"+cfg.AppPort+"/swagger/doc.json"),
	))

	return middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.RequestID,
		middleware.Logging,
	)
}
