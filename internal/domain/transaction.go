package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDuplicateIdempotencyKey     = errors.New("duplicate idempotency key")
	ErrTransactionNotFound         = errors.New("transaction not found")
	ErrOperationTypeNotFound       = errors.New("operation type not found")
	ErrIdempotencyKeyOwnerMismatch = errors.New("idempotency key belongs to a different account")
)

const (
	OperationPurchase            = 1
	OperationInstallmentPurchase = 2
	OperationWithdrawal          = 3
	OperationPayment             = 4
)

type Transaction struct {
	ID              int64     `json:"transaction_id"`
	AccountID       int64     `json:"account_id"`
	OperationTypeID int       `json:"operation_type_id"`
	Amount          float64   `json:"amount"`
	EventDate       time.Time `json:"event_date"`
	IdempotencyKey  string    `json:"-"`
}

type TransactionRepository interface {
	Create(ctx context.Context, t Transaction) (Transaction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (Transaction, error)
}
