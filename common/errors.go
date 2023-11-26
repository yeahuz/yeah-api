package common

import (
	"net/http"

	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.Get("en")

var (
	MethodNotAllowed = ErrMethodNotAllowed{Message: l.T("Method not allowed"), StatusCode: http.StatusMethodNotAllowed}
	Internal         = ErrInternal{Message: l.T("Internal server error"), StatusCode: http.StatusInternalServerError}
)

type ErrMethodNotAllowed struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (emna ErrMethodNotAllowed) Name() string {
	return "error.methodNotAllowed"
}

func (emna ErrMethodNotAllowed) Error() string {
	return emna.Message
}

type ErrInternal struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (ei ErrInternal) Name() string {
	return "error.internal"
}

func (ei ErrInternal) Error() string {
	return ei.Message
}

type ErrNotFound struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (enf ErrNotFound) Name() string {
	return "error.notFound"
}

func (enf ErrNotFound) Error() string {
	return enf.Message
}

type ErrBadRequest struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (ebr ErrBadRequest) Name() string {
	return "error.badRequest"
}

func (ebr ErrBadRequest) Error() string {
	return ebr.Message
}

type ErrValidation struct {
	Message    string            `json:"message"`
	StatusCode int               `json:"status_code"`
	Errors     map[string]string `json:"errors,omitempty"`
}

func (ev ErrValidation) Name() string {
	return "error.valdation"
}

func (ev ErrValidation) Error() string {
	return ev.Message
}
