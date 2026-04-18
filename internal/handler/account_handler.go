package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Guizzs26/pismo/internal/domain"
	"github.com/Guizzs26/pismo/internal/service"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(service *service.AccountService) *AccountHandler {
	return &AccountHandler{service: service}
}

type createAccountRequest struct {
	DocumentNumber string `json:"document_number"`
}

type createAccountResponse struct {
	AccountID      int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DocumentNumber == "" {
		writeError(w, http.StatusBadRequest, "document_number is required")
		return
	}

	acc, err := h.service.Create(r.Context(), req.DocumentNumber)
	if err != nil {
		if errors.Is(err, domain.ErrDocumentAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, createAccountResponse{
		AccountID:      acc.ID,
		DocumentNumber: acc.DocumentNumber,
	})
}

func (h *AccountHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("accountId")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid account id")
		return
	}

	acc, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, createAccountResponse{
		AccountID:      acc.ID,
		DocumentNumber: acc.DocumentNumber,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
