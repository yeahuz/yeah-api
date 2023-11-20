package common

import "net/http"

var (
	ErrInternal         = ApiError{Message: "Internal server error", Code: http.StatusInternalServerError}
	ErrMethodNotAllowed = ApiError{Message: "Method not allowed", Code: http.StatusMethodNotAllowed}
)

func ErrBadRequest(message string) ApiError {
	return ApiError{
		Message: message,
		Code:    http.StatusBadRequest,
	}
}
