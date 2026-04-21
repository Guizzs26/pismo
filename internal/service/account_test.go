package service

import (
	"context"
	"testing"

	"github.com/Guizzs26/pismo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(ctx context.Context, acc domain.Account) (domain.Account, error) {
	args := m.Called(ctx, acc)
	return args.Get(0).(domain.Account), args.Error(1)
}

func (m *MockAccountRepository) FindByID(ctx context.Context, id int64) (domain.Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Account), args.Error(1)
}

func TestAccountService_Create_Success_ReturnsAccountWithID(t *testing.T) {
	repo := new(MockAccountRepository)
	svc := NewAccountService(repo)

	input := domain.Account{DocumentNumber: "12345678900"}
	returned := domain.Account{ID: 1, DocumentNumber: "12345678900"}

	repo.On("Create", mock.Anything, input).Return(returned, nil)

	result, err := svc.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "12345678900", result.DocumentNumber)
	repo.AssertExpectations(t)
}

func TestAccountService_Create_DuplicateDocument_ReturnsErrDocumentAlreadyExists(t *testing.T) {
	repo := new(MockAccountRepository)
	svc := NewAccountService(repo)

	input := domain.Account{DocumentNumber: "12345678900"}

	repo.On("Create", mock.Anything, input).Return(domain.Account{}, domain.ErrDocumentAlreadyExists)

	result, err := svc.Create(context.Background(), input)

	assert.ErrorIs(t, err, domain.ErrDocumentAlreadyExists)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestAccountService_FindByID_Success_ReturnsAccount(t *testing.T) {
	repo := new(MockAccountRepository)
	svc := NewAccountService(repo)

	expected := domain.Account{ID: 1, DocumentNumber: "12345678900"}

	repo.On("FindByID", mock.Anything, int64(1)).Return(expected, nil)

	result, err := svc.FindByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "12345678900", result.DocumentNumber)
	repo.AssertExpectations(t)
}

func TestAccountService_FindByID_NotFound_ReturnsErrAccountNotFound(t *testing.T) {
	repo := new(MockAccountRepository)
	svc := NewAccountService(repo)

	repo.On("FindByID", mock.Anything, int64(999)).Return(domain.Account{}, domain.ErrAccountNotFound)

	result, err := svc.FindByID(context.Background(), 999)

	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}
