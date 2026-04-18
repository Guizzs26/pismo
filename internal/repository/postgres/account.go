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

type AccountRepository struct {
	db db.DBTX
}

func NewAccountRepository(db db.DBTX) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(
	ctx context.Context,
	account domain.Account,
) (domain.Account, error) {
	const query = `
		INSERT INTO accounts (document_number)
		VALUES ($1)
		RETURNING account_id, document_number
	`

	if err := r.db.QueryRow(ctx, query, account.DocumentNumber).Scan(
		&account.ID,
		&account.DocumentNumber,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Account{}, domain.ErrDocumentAlreadyExists
		}
		return domain.Account{}, fmt.Errorf("inserting account with document %q: %w", account.DocumentNumber, err)
	}

	return account, nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	const query = `
		SELECT account_id, document_number
		FROM accounts
		WHERE account_id = $1
	`

	var account domain.Account
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.DocumentNumber,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}
		return domain.Account{}, fmt.Errorf("querying account with id %d: %v", id, err)
	}

	return account, nil
}
