package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Guizzs26/pismo/internal/domain"
	"github.com/Guizzs26/pismo/internal/middleware"
	"github.com/Guizzs26/pismo/internal/service"
	"github.com/Guizzs26/pismo/pkg/httpx"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

type createAccountRequest struct {
	DocumentNumber string `json:"document_number" validate:"required,numeric"`
}

type accountResponse struct {
	AccountID      int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.Decode[createAccountRequest](w, r)
	if err != nil {
		if de, ok := httpx.IsValidationError(err); ok {
			httpx.ValidationFailed(w, de.Details)
			return
		}
		httpx.BadRequest(w, err.Error())
		return
	}

	acc, err := h.service.Create(r.Context(), req.DocumentNumber)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentAlreadyExists) {
			httpx.Conflict(w, "document number already exists")
			return
		}
		slog.Error("unexpected error creating account",
			"error", err,
			"request_id", middleware.GetRequestID(r.Context()),
		)

		httpx.InternalServerError(w)
		return
	}

	httpx.Success(w, http.StatusCreated, accountResponse{
		AccountID:      acc.ID,
		DocumentNumber: acc.DocumentNumber,
	})
}

func (h *AccountHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("accountId"), 10, 64)
	if err != nil || id <= 0 {
		httpx.BadRequest(w, "invalid account id")
		return
	}

	acc, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			httpx.NotFound(w, "account not found")
			return
		}

		slog.Error("unexpected error fetching account",
			"error", err,
			"account_id", id,
			"request_id", middleware.GetRequestID(r.Context()),
		)
		httpx.InternalServerError(w)
		return
	}

	httpx.Success(w, http.StatusOK, accountResponse{
		AccountID:      acc.ID,
		DocumentNumber: acc.DocumentNumber,
	})
}
