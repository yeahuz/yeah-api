package common

import (
	"net/http"

	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.Get("en")
var (
	ErrInternal         = ApiError{Message: l.T("Internal server error"), Code: http.StatusInternalServerError}
	ErrMethodNotAllowed = ApiError{Message: l.T("Method not allowed"), Code: http.StatusMethodNotAllowed}
)

func ErrBadRequest(message string) ApiError {
	return ApiError{
		Message: message,
		Code:    http.StatusBadRequest,
	}
}

func ErrValidation(message string, errors map[string]string) ApiError {
	return ApiError{
		Message: message,
		Code:    http.StatusUnprocessableEntity,
		Errors:  errors,
	}
}
