package errors

import (
	"net/http"

	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.Get("en")

var (
	MethodNotAllowed = ErrMethodNotAllowed{Message: l.T("Method not allowed"), StatusCode: http.StatusMethodNotAllowed}
	Internal         = ErrInternal{Message: l.T("Internal server error"), StatusCode: http.StatusInternalServerError}
)

type AppError interface {
	error
	Name() string
	ErrorMap() map[string]string
	Status() int
	SetError(message string)
}

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

func (emna ErrMethodNotAllowed) ErrorMap() map[string]string {
	return nil
}

func (emna *ErrMethodNotAllowed) SetError(message string) {
	emna.Message = message
}

func (emna ErrMethodNotAllowed) Status() int {
	return emna.StatusCode
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

func (ei ErrInternal) ErrorMap() map[string]string {
	return nil
}

func (ei ErrInternal) SetError(message string) {
	ei.Message = message
}

func (ei ErrInternal) Status() int {
	return ei.StatusCode
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

func (enf ErrNotFound) ErrorMap() map[string]string {
	return nil
}

func (enf ErrNotFound) SetError(message string) {
	enf.Message = message
}

func (enf ErrNotFound) Status() int {
	return enf.StatusCode
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

func (ebr ErrBadRequest) ErrorMap() map[string]string {
	return nil
}

func (ebr ErrBadRequest) SetError(message string) {
	ebr.Message = message
}

func (ebr ErrBadRequest) Status() int {
	return ebr.StatusCode
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

func (ev ErrValidation) ErrorsMap() map[string]string {
	return ev.Errors
}

func (ev ErrValidation) SetError(message string) {
	ev.Message = message
}

func (ev ErrValidation) Status() int {
	return ev.StatusCode
}
