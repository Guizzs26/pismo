package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Guizzs26/pismo/internal/domain"
	"github.com/Guizzs26/pismo/internal/middleware"
	"github.com/Guizzs26/pismo/internal/service"
	"github.com/Guizzs26/pismo/pkg/httpx"
)

type TransactionHandler struct {
	service *service.TransactionService
}

func NewTransactionHandler(service *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

type CreateTransactionRequest struct {
	AccountID       int64   `json:"account_id"        validate:"required,gt=0"`
	OperationTypeID int     `json:"operation_type_id" validate:"required,gt=0"`
	Amount          float64 `json:"amount"            validate:"required,gt=0"`
}

type TransactionResponse struct {
	TransactionID   int64   `json:"transaction_id"`
	AccountID       int64   `json:"account_id"`
	OperationTypeID int     `json:"operation_type_id"`
	Amount          float64 `json:"amount"`
}

// @Summary     Create transaction
// @Description Creates a new financial transaction for an account
// @Tags        transactions
// @Accept      json
// @Produce     json
// @Param       Idempotency-Key header string true "Unique key to prevent duplicate transactions"
// @Param       request body CreateTransactionRequest true "Transaction payload"
// @Success     201 {object} TransactionResponse
// @Success     200 {object} TransactionResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Router      /transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		httpx.BadRequest(w, "Idempotency-Key header is required")
		return
	}

	req, err := httpx.Decode[CreateTransactionRequest](w, r)
	if err != nil {
		if de, ok := httpx.IsValidationError(err); ok {
			httpx.ValidationFailed(w, de.Details)
			return
		}
		httpx.BadRequest(w, err.Error())
		return
	}

	t := domain.Transaction{
		AccountID:       req.AccountID,
		OperationTypeID: req.OperationTypeID,
		Amount:          req.Amount,
	}

	created, isNew, err := h.service.Create(r.Context(), t, idempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAccountNotFound):
			httpx.NotFound(w, "account not found")
		case errors.Is(err, domain.ErrOperationTypeNotFound):
			httpx.NotFound(w, "operation type not found")
		case errors.Is(err, domain.ErrIdempotencyKeyOwnerMismatch):
			httpx.Conflict(w, "idempotency key belongs to a different account")
		default:
			slog.Error("unexpected error creating transaction",
				"error", err,
				"account_id", req.AccountID,
				"operation_type_id", req.OperationTypeID,
				"request_id", middleware.GetRequestID(r.Context()),
			)
			httpx.InternalServerError(w)
		}
		return
	}

	status := http.StatusCreated
	if !isNew {
		status = http.StatusOK
	}

	httpx.Success(w, status, TransactionResponse{
		TransactionID:   created.ID,
		AccountID:       created.AccountID,
		OperationTypeID: created.OperationTypeID,
		Amount:          created.Amount,
	})
}
