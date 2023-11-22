package common

import "net/http"

type ApiError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors,omitempty"`
}

type Object interface {
	Name() string
}

type Response struct {
	Object string      `json:"_,omitempty"`
	Data   interface{} `json:"data"`
}

func (e ApiError) Error() string {
	return e.Message
}

type ApiFunc func(w http.ResponseWriter, r *http.Request) error
