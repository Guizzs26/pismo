package domain

import (
	"context"
	"errors"
)

var (
	ErrDocumentAlreadyExists = errors.New("document number already exists")
	ErrAccountNotFound       = errors.New("account not found")
)

type Account struct {
	ID             int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

type AccountRepository interface {
	Create(ctx context.Context, account *Account) (Account, error)
	FindByID(ctx context.Context, id int64) (Account, error)
}
