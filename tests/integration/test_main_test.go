package integration

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Guizzs26/pismo/internal/handler"
	"github.com/Guizzs26/pismo/internal/middleware"
	pg "github.com/Guizzs26/pismo/internal/repository/postgres"
	"github.com/Guizzs26/pismo/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	testPool   *pgxpool.Pool
	testServer *httptest.Server
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pismodb_test"),
		postgres.WithUsername("testusr"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("to start postgres container: %v", err)
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("terminate container: %v", err)
		}
	}()

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("get connection string: %v", err)
	}

	testPool, err = pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("create pool: %v", err)
	}
	defer testPool.Close()

	if err := runMigrations(ctx, testPool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	testServer = setupTestServer(testPool)
	defer testServer.Close()

	code := m.Run()

	os.Exit(code)
}

func setupTestServer(pool *pgxpool.Pool) *httptest.Server {
	mux := http.NewServeMux()

	accountRepo := pg.NewAccountRepository(pool)
	accountService := service.NewAccountService(accountRepo)
	accountHandler := handler.NewAccountHandler(accountService)

	transactionRepo := pg.NewTransactionRepository(pool)
	transactionService := service.NewTransactionService(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	mux.HandleFunc("POST /accounts", accountHandler.Create)
	mux.HandleFunc("GET /accounts/{accountId}", accountHandler.FindByID)
	mux.HandleFunc("POST /transactions", transactionHandler.Create)

	chainedHandler := middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.RequestID,
	)

	return httptest.NewServer(chainedHandler)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationsPath := filepath.Join("..", "..", "migrations")

	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			sqlBytes, err := os.ReadFile(filepath.Join(migrationsPath, file.Name()))
			if err != nil {
				return err
			}

			_, err = pool.Exec(ctx, string(sqlBytes))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func clearDB(t *testing.T) {
	t.Helper()

	query := `
		TRUNCATE TABLE transactions RESTART IDENTITY CASCADE;
		TRUNCATE TABLE accounts RESTART IDENTITY CASCADE;
	`
	_, err := testPool.Exec(context.Background(), query)
	if err != nil {
		t.Fatalf("truncate database: %v", err)
	}
}
