package domain

import (
	"context"
	"time"
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
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) (Transaction, error)
}
