package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Guizzs26/pismo/internal/domain"
)

type TransactionService struct {
	repo domain.TransactionRepository
}

func NewTransactionService(repo domain.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Create(
	ctx context.Context,
	t domain.Transaction,
	idempotencyKey string,
) (domain.Transaction, bool, error) {
	t.Amount = s.applySign(t.OperationTypeID, t.Amount)
	t.IdempotencyKey = idempotencyKey

	createdT, err := s.repo.Create(ctx, t)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateIdempotencyKey) {
			existing, fetchErr := s.repo.FindByIdempotencyKey(ctx, idempotencyKey)
			if fetchErr != nil {
				return domain.Transaction{}, false, fmt.Errorf("recovering idempotent transaction for key %q: %w", idempotencyKey, fetchErr)
			}

			if existing.AccountID != t.AccountID {
				return domain.Transaction{}, false, domain.ErrIdempotencyKeyOwnerMismatch
			}

			return existing, false, nil
		}

		return domain.Transaction{}, false, fmt.Errorf("creating transaction for account %d: %w", t.AccountID, err)
	}

	return createdT, true, nil
}

func (s *TransactionService) applySign(opTypeID int, amount float64) float64 {
	absAmount := math.Abs(amount)
	switch opTypeID {
	case domain.OperationPurchase,
		domain.OperationInstallmentPurchase,
		domain.OperationWithdrawal:
		return -absAmount
	default:
		return absAmount
	}
}
