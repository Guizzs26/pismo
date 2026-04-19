package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/Guizzs26/pismo/internal/domain"
	db "github.com/Guizzs26/pismo/internal/infra/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TransactionRepository struct {
	db db.DBTX
}

func NewTransactionRepository(db db.DBTX) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(
	ctx context.Context,
	t domain.Transaction,
) (domain.Transaction, error) {
	const query = `
		INSERT INTO transactions (account_id, operation_type_id, amount, idempotency_key)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (idempotency_key) DO NOTHING
		RETURNING transaction_id, account_id, operation_type_id, amount, event_date
	`

	if err := r.db.QueryRow(ctx, query, t.AccountID, t.OperationTypeID, t.Amount, t.IdempotencyKey).Scan(
		&t.ID,
		&t.AccountID,
		&t.OperationTypeID,
		&t.Amount,
		&t.EventDate,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, domain.ErrDuplicateIdempotencyKey
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "fk_account":
				return domain.Transaction{}, domain.ErrAccountNotFound
			case "fk_operation_type":
				return domain.Transaction{}, domain.ErrOperationTypeNotFound
			}
		}
		return domain.Transaction{}, fmt.Errorf("inserting transaction for account %d: %v", t.AccountID, err)
	}

	return t, nil
}

func (r *TransactionRepository) FindByIdempotencyKey(
	ctx context.Context,
	key string,
) (domain.Transaction, error) {
	const query = `
		SELECT transaction_id, account_id, operation_type_id, amount, event_date
		FROM transactions
		WHERE idempotency_key = $1
	`

	var t domain.Transaction
	err := r.db.QueryRow(ctx, query, key).Scan(
		&t.ID,
		&t.AccountID,
		&t.OperationTypeID,
		&t.Amount,
		&t.EventDate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Transaction{}, domain.ErrTransactionNotFound
		}
		return domain.Transaction{}, fmt.Errorf("querying transaction by idempotency key %q: %w", key, err)
	}

	return t, nil
}
