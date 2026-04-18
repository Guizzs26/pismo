package httpx

const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeBadRequest       = "BAD_REQUEST"
	ErrCodeNotFound         = "NOT_FOUND"
	ErrCodeConflict         = "CONFLICT"
	ErrCodeInternalServer   = "INTERNAL_SERVER_ERROR"
)

type ErrorDetail struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func newErrorResponse(code, message string, details []ErrorDetail) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
