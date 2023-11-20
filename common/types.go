package common

import "net/http"

type ApiError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func (e ApiError) Error() string {
	return e.Message
}

type ApiFunc func(w http.ResponseWriter, r *http.Request) error
