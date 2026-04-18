package service

import (
	"context"

	"github.com/Guizzs26/pismo/internal/domain"
)

type AccountService struct {
	repo domain.AccountRepository
}

func NewAccountService(repo domain.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) Create(ctx context.Context, documentNumber string) (domain.Account, error) {
	acc := &domain.Account{
		DocumentNumber: documentNumber,
	}

	return s.repo.Create(ctx, acc)
}

func (s *AccountService) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	return s.repo.FindByID(ctx, id)
}
