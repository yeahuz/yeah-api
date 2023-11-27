package errors

import (
	"net/http"

	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.Get("en")

var (
	MethodNotAllowed = NewMethodNotAllowed()
	Internal         = NewInternal()
	NotFound         = NewNotFound(l.T("Resource not found"))
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

func NewMethodNotAllowed() errMethodNotAllowed {
	return errMethodNotAllowed{
		Message:    l.T("Method not allowed"),
		StatusCode: http.StatusMethodNotAllowed,
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

func NewInternal() errInternal {
	return errInternal{
		Message:    l.T("Internal server error"),
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

func (ev errValidation) ErrorsMap() map[string]string {
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
