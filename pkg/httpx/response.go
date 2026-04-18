package httpx

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, status int, data any) {
	JSON(w, status, data)
}

func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, newErrorResponse(ErrCodeBadRequest, message, nil))
}

func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, newErrorResponse(ErrCodeNotFound, message, nil))
}

func Conflict(w http.ResponseWriter, message string) {
	JSON(w, http.StatusConflict, newErrorResponse(ErrCodeConflict, message, nil))
}

func InternalServerError(w http.ResponseWriter) {
	JSON(w, http.StatusInternalServerError, newErrorResponse(ErrCodeInternalServer, "internal server error", nil))
}

func ValidationFailed(w http.ResponseWriter, details []ErrorDetail) {
	JSON(w, http.StatusUnprocessableEntity, newErrorResponse(ErrCodeValidationFailed, "one or more fields are invalid", details))
}
