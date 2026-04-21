package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const maxBodyBytes = 1 << 20 // 1MB

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "" || name == "-" {
			return fld.Name
		}
		return strings.Split(name, ",")[0]
	})
}

type DecodeError struct {
	Message string
	Details []ErrorDetail
}

func (e *DecodeError) Error() string {
	return e.Message
}

func IsValidationError(err error) (*DecodeError, bool) {
	var de *DecodeError
	if errors.As(err, &de) && de.Details != nil {
		return de, true
	}
	return nil, false
}

func Decode[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var payload T

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return payload, &DecodeError{Message: "invalid request body"}
	}

	if err := validate.Struct(payload); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := make([]ErrorDetail, len(ve))
			for i, fe := range ve {
				details[i] = ErrorDetail{
					Field: fe.Field(),
					Issue: validationMessage(fe),
				}
			}
			return payload, &DecodeError{
				Message: "one or more fields are invalid",
				Details: details,
			}
		}
	}

	return payload, nil
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "min":
		return "value is too short"
	case "max":
		return "value is too long"
	case "numeric":
		return "must be numeric"
	case "email":
		return "must be a valid email"
	default:
		return "invalid value"
	}
}
