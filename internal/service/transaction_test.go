package service

import (
	"context"
	"testing"
	"time"

	"github.com/Guizzs26/pismo/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, t domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, t)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByIdempotencyKey(ctx context.Context, key string) (domain.Transaction, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func TestTransactionService_Create_DebitOperation_AmountBecomesNegative(t *testing.T) {
	debitOperations := []struct {
		name        string
		operationID int
	}{
		{"purchase", domain.OperationPurchase},
		{"installment purchase", domain.OperationInstallmentPurchase},
		{"withdrawal", domain.OperationWithdrawal},
	}

	for _, op := range debitOperations {
		t.Run(op.name, func(t *testing.T) {
			repo := new(MockTransactionRepository)
			svc := NewTransactionService(repo)

			input := domain.Transaction{
				AccountID:       1,
				OperationTypeID: op.operationID,
				Amount:          100.0,
			}

			expectedStored := input
			expectedStored.Amount = -100.0
			expectedStored.IdempotencyKey = "test-key"

			returned := expectedStored
			returned.ID = 1
			returned.EventDate = time.Now()

			repo.On("Create", mock.Anything, expectedStored).Return(returned, nil)

			result, isNew, err := svc.Create(context.Background(), input, "test-key")

			assert.NoError(t, err)
			assert.True(t, isNew)
			assert.Equal(t, -100.0, result.Amount)
			assert.Equal(t, int64(1), result.ID)
			repo.AssertExpectations(t)
		})
	}
}

func TestTransactionService_Create_PaymentOperation_AmountStaysPositive(t *testing.T) {
	repo := new(MockTransactionRepository)
	svc := NewTransactionService(repo)

	input := domain.Transaction{
		AccountID:       1,
		OperationTypeID: domain.OperationPayment,
		Amount:          100.0,
	}

	expectedStored := input
	expectedStored.IdempotencyKey = "test-key"

	returned := expectedStored
	returned.ID = 1
	returned.EventDate = time.Now()

	repo.On("Create", mock.Anything, expectedStored).Return(returned, nil)

	result, isNew, err := svc.Create(context.Background(), input, "test-key")

	assert.NoError(t, err)
	assert.True(t, isNew)
	assert.Equal(t, 100.0, result.Amount)
	assert.Equal(t, int64(1), result.ID)
	repo.AssertExpectations(t)
}

func TestTransactionService_Create_InvalidAccountID_ReturnsErrAccountNotFound(t *testing.T) {
	repo := new(MockTransactionRepository)
	svc := NewTransactionService(repo)

	input := domain.Transaction{
		AccountID:       999,
		OperationTypeID: domain.OperationPayment,
		Amount:          100.0,
	}

	expectedStored := input
	expectedStored.IdempotencyKey = "test-key"

	repo.On("Create", mock.Anything, expectedStored).Return(domain.Transaction{}, domain.ErrAccountNotFound)

	result, isNew, err := svc.Create(context.Background(), input, "test-key")

	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
	assert.False(t, isNew)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestTransactionService_Create_InvalidOperationTypeID_ReturnsErrOperationTypeNotFound(t *testing.T) {
	repo := new(MockTransactionRepository)
	svc := NewTransactionService(repo)

	input := domain.Transaction{
		AccountID:       1,
		OperationTypeID: 999,
		Amount:          100.0,
	}

	expectedStored := input
	expectedStored.Amount = input.Amount
	expectedStored.IdempotencyKey = "test-key"

	repo.On("Create", mock.Anything, expectedStored).Return(domain.Transaction{}, domain.ErrOperationTypeNotFound)

	result, isNew, err := svc.Create(context.Background(), input, "test-key")

	assert.ErrorIs(t, err, domain.ErrOperationTypeNotFound)
	assert.False(t, isNew)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestTransactionService_Create_IdempotentReplay_ReturnsExisting(t *testing.T) {
	repo := new(MockTransactionRepository)
	svc := NewTransactionService(repo)

	input := domain.Transaction{
		AccountID:       1,
		OperationTypeID: domain.OperationPayment,
		Amount:          100,
	}

	expectedStored := input
	expectedStored.IdempotencyKey = "key"

	existing := expectedStored
	existing.ID = 1
	existing.EventDate = time.Now()

	repo.On("Create", mock.Anything, expectedStored).
		Return(domain.Transaction{}, domain.ErrDuplicateIdempotencyKey)

	repo.On("FindByIdempotencyKey", mock.Anything, "key").
		Return(existing, nil)

	result, isNew, err := svc.Create(context.Background(), input, "key")

	assert.NoError(t, err)
	assert.False(t, isNew)
	assert.Equal(t, int64(1), result.ID)
	repo.AssertExpectations(t)
}

func TestTransactionService_Create_IdempotencyOwnerMismatch_ReturnsErr(t *testing.T) {
	repo := new(MockTransactionRepository)
	svc := NewTransactionService(repo)

	input := domain.Transaction{
		AccountID:       2,
		OperationTypeID: domain.OperationPayment,
		Amount:          100,
	}

	existing := domain.Transaction{
		ID:              1,
		AccountID:       1,
		OperationTypeID: domain.OperationPayment,
		Amount:          100,
		IdempotencyKey:  "key",
	}

	repo.On("Create", mock.Anything, mock.Anything).
		Return(domain.Transaction{}, domain.ErrDuplicateIdempotencyKey)

	repo.On("FindByIdempotencyKey", mock.Anything, "key").
		Return(existing, nil)

	result, isNew, err := svc.Create(context.Background(), input, "key")

	assert.ErrorIs(t, err, domain.ErrIdempotencyKeyOwnerMismatch)
	assert.False(t, isNew)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}

func TestTransactionService_Create_IdempotencyFetchFails_ReturnsErr(t *testing.T) {
	repo := new(MockTransactionRepository)
	svc := NewTransactionService(repo)

	input := domain.Transaction{
		AccountID:       1,
		OperationTypeID: domain.OperationPayment,
		Amount:          100,
	}

	repo.On("Create", mock.Anything, mock.Anything).
		Return(domain.Transaction{}, domain.ErrDuplicateIdempotencyKey)

	repo.On("FindByIdempotencyKey", mock.Anything, "key").
		Return(domain.Transaction{}, assert.AnError)

	result, isNew, err := svc.Create(context.Background(), input, "key")

	assert.Error(t, err)
	assert.False(t, isNew)
	assert.Empty(t, result)
	repo.AssertExpectations(t)
}
