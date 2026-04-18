package service

import (
	"context"
	"fmt"

	"github.com/Guizzs26/pismo/internal/domain"
)

type AccountService struct {
	repo domain.AccountRepository
}

func NewAccountService(repo domain.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Create(ctx context.Context, documentNumber string) (domain.Account, error) {
	acc := domain.Account{
		DocumentNumber: documentNumber,
	}

	acc, err := s.repo.Create(ctx, acc)
	if err != nil {
		return domain.Account{}, fmt.Errorf("registering account: %w", err)
	}

	return acc, nil
}

func (s *AccountService) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	acc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Account{}, fmt.Errorf("retrieving account details: %w", err)
	}

	return acc, nil
}
