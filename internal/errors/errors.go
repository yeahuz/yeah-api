package errors

import (
	"net/http"

	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.Get("en")

var (
	MethodNotAllowed = NewMethodNotAllowed(l.T("Method not allowed"))
	Internal         = NewInternal(l.T("Internal server error"))
	NotFound         = NewNotFound(l.T("Resource not found"))
	Unauthorized     = NewUnauthorized(l.T("Not authorized"))
)

type AppError interface {
	error
	Name() string
	ErrorMap() map[string]string
	Status() int
	SetError(message string)
}

type errMethodNotAllowed struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (emna errMethodNotAllowed) Name() string {
	return "error.methodNotAllowed"
}

func (emna errMethodNotAllowed) Error() string {
	return emna.Message
}

func (emna errMethodNotAllowed) ErrorMap() map[string]string {
	return nil
}

func (emna *errMethodNotAllowed) SetError(message string) {
	emna.Message = message
}

func (emna errMethodNotAllowed) Status() int {
	return emna.StatusCode
}

func NewMethodNotAllowed(message string) errMethodNotAllowed {
	return errMethodNotAllowed{
		Message:    message,
		StatusCode: http.StatusMethodNotAllowed,
	}
}

type errUnauthorized struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e errUnauthorized) Name() string {
	return "error.internal"
}

func (e errUnauthorized) Error() string {
	return e.Message
}

func (e errUnauthorized) ErrorMap() map[string]string {
	return nil
}

func (e errUnauthorized) SetError(message string) {
	e.Message = message
}

func (e errUnauthorized) Status() int {
	return e.StatusCode
}

func NewUnauthorized(message string) errUnauthorized {
	return errUnauthorized{
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

type errForbidden struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (e errForbidden) Name() string {
	return "error.internal"
}

func (e errForbidden) Error() string {
	return e.Message
}

func (e errForbidden) ErrorMap() map[string]string {
	return nil
}

func (e errForbidden) SetError(message string) {
	e.Message = message
}

func (e errForbidden) Status() int {
	return e.StatusCode
}

func NewForbidden(message string) errUnauthorized {
	return errUnauthorized{
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

type errInternal struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (ei errInternal) Name() string {
	return "error.internal"
}

func (ei errInternal) Error() string {
	return ei.Message
}

func (ei errInternal) ErrorMap() map[string]string {
	return nil
}

func (ei errInternal) SetError(message string) {
	ei.Message = message
}

func (ei errInternal) Status() int {
	return ei.StatusCode
}

func NewInternal(message string) errInternal {
	return errInternal{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

type errNotFound struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (enf errNotFound) Name() string {
	return "error.notFound"
}

func (enf errNotFound) Error() string {
	return enf.Message
}

func (enf errNotFound) ErrorMap() map[string]string {
	return nil
}

func (enf errNotFound) SetError(message string) {
	enf.Message = message
}

func (enf errNotFound) Status() int {
	return enf.StatusCode
}

func NewNotFound(message string) errNotFound {
	return errNotFound{
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

type errBadRequest struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (ebr errBadRequest) Name() string {
	return "error.badRequest"
}

func (ebr errBadRequest) Error() string {
	return ebr.Message
}

func (ebr errBadRequest) ErrorMap() map[string]string {
	return nil
}

func (ebr errBadRequest) SetError(message string) {
	ebr.Message = message
}

func (ebr errBadRequest) Status() int {
	return ebr.StatusCode
}

func NewBadRequest(message string) errBadRequest {
	return errBadRequest{
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

type errValidation struct {
	Message    string            `json:"message"`
	StatusCode int               `json:"status_code"`
	Errors     map[string]string `json:"errors,omitempty"`
}

func (ev errValidation) Name() string {
	return "error.valdation"
}

func (ev errValidation) Error() string {
	return ev.Message
}

func (ev errValidation) ErrorMap() map[string]string {
	return ev.Errors
}

func (ev errValidation) SetError(message string) {
	ev.Message = message
}

func (ev errValidation) Status() int {
	return ev.StatusCode
}

func NewValidation(errors map[string]string) errValidation {
	return errValidation{
		Message:    l.T("Validation failed"),
		Errors:     errors,
		StatusCode: http.StatusUnprocessableEntity,
	}
}
