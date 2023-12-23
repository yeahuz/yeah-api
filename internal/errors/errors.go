package errors

import (
	"bytes"

	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/client"
	"github.com/yeahuz/yeah-api/internal/localizer"
	"github.com/yeahuz/yeah-api/user"
)

var l = localizer.Get("en")

type Op string
type Kind uint8

const Separator = ":\n\t"

const (
	Other Kind = iota
	Invalid
	Permission
	NotExist
	Exist
	Internal
)

func (k Kind) String() string {
	switch k {
	case Internal:
		return "internal error"
	case Other:
		return "other error"
	case NotExist:
		return "item does not exit"
	case Exist:
		return "item already exists"
	case Permission:
		return "permission denied"
	}
	return "unknown error"
}

type Error struct {
	Op       Op
	Kind     Kind
	Err      error
	UserID   yeahapi.UserID
	ClientID yeahapi.ClientID
}

func (e *Error) isZero() bool {
	return (e.UserID == "" || e.ClientID == "") && e.Op == "" && e.Kind == 0 && e.Err == nil
}

func Ops(e *Error) []Op {
	res := []Op{e.Op}

	subErr, ok := e.Err.(*Error)

	if !ok {
		return res
	}

	res = append(res, Ops(subErr)...)
	return res
}

func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
}

func (e Error) Error() string {
	b := new(bytes.Buffer)

	if e.Op != "" {
		b.WriteString(string(e.Op))
	}
	if e.UserID != "" {
		pad(b, ": ")
		b.WriteString("user ")
		b.WriteString(string(e.UserID))
	}

	if e.Kind != 0 {
		pad(b, ": ")
		b.WriteString(e.Kind.String())
	}

	if e.Err != nil {
		if prevErr, ok := e.Err.(*Error); ok {
			if !prevErr.isZero() {
				pad(b, Separator)
				b.WriteString(e.Err.Error())
			}
		} else {
			pad(b, ": ")
			b.WriteString(e.Err.Error())
		}
	}

	if b.Len() == 0 {
		return "no error"
	}

	return b.String()
}

func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}

	e := &Error{}

	for _, arg := range args {
		switch arg := arg.(type) {
		case user.UserID:
			e.UserID = arg
		case client.ClientID:
			e.ClientID = arg
		case Op:
			e.Op = arg
		case error:
			e.Err = arg
		case Kind:
			e.Kind = arg
		case string:
			e.Err = Str(arg)
		case *Error:
			copy := *arg
			e.Err = &copy
		default:
			panic("bad call to errors.E")
		}
	}

	prev, ok := e.Err.(*Error)

	if !ok {
		return e
	}

	if prev.UserID == e.UserID {
		prev.UserID = ""
	}

	if prev.Kind == e.Kind {
		prev.Kind = Other
	}

	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}

	return e
}

func Str(text string) error {
	return &errorString{text}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// type AppError interface {
// 	error
// 	Name() string
// 	ErrorMap() map[string]string
// 	Status() int
// 	SetError(message string)
// }

// type errMethodNotAllowed struct {
// 	Message    string `json:"message"`
// 	StatusCode int    `json:"status_code"`
// }

// func (emna errMethodNotAllowed) Name() string {
// 	return "error.methodNotAllowed"
// }

// func (emna errMethodNotAllowed) Error() string {
// 	return emna.Message
// }

// func (emna errMethodNotAllowed) ErrorMap() map[string]string {
// 	return nil
// }

// func (emna *errMethodNotAllowed) SetError(message string) {
// 	emna.Message = message
// }

// func (emna errMethodNotAllowed) Status() int {
// 	return emna.StatusCode
// }

// func NewMethodNotAllowed(message string) errMethodNotAllowed {
// 	return errMethodNotAllowed{
// 		Message:    message,
// 		StatusCode: http.StatusMethodNotAllowed,
// 	}
// }

// type errUnauthorized struct {
// 	Message    string `json:"message"`
// 	StatusCode int    `json:"status_code"`
// }

// func (e errUnauthorized) Name() string {
// 	return "error.internal"
// }

// func (e errUnauthorized) Error() string {
// 	return e.Message
// }

// func (e errUnauthorized) ErrorMap() map[string]string {
// 	return nil
// }

// func (e errUnauthorized) SetError(message string) {
// 	e.Message = message
// }

// func (e errUnauthorized) Status() int {
// 	return e.StatusCode
// }

// func NewUnauthorized(message string) errUnauthorized {
// 	return errUnauthorized{
// 		Message:    message,
// 		StatusCode: http.StatusUnauthorized,
// 	}
// }

// type errForbidden struct {
// 	Message    string `json:"message"`
// 	StatusCode int    `json:"status_code"`
// }

// func (e errForbidden) Name() string {
// 	return "error.internal"
// }

// func (e errForbidden) Error() string {
// 	return e.Message
// }

// func (e errForbidden) ErrorMap() map[string]string {
// 	return nil
// }

// func (e errForbidden) SetError(message string) {
// 	e.Message = message
// }

// func (e errForbidden) Status() int {
// 	return e.StatusCode
// }

// func NewForbidden(message string) errUnauthorized {
// 	return errUnauthorized{
// 		Message:    message,
// 		StatusCode: http.StatusForbidden,
// 	}
// }

// type errInternal struct {
// 	Message    string `json:"message"`
// 	StatusCode int    `json:"status_code"`
// }

// func (ei errInternal) Name() string {
// 	return "error.internal"
// }

// func (ei errInternal) Error() string {
// 	return ei.Message
// }

// func (ei errInternal) ErrorMap() map[string]string {
// 	return nil
// }

// func (ei errInternal) SetError(message string) {
// 	ei.Message = message
// }

// func (ei errInternal) Status() int {
// 	return ei.StatusCode
// }

// func NewInternal(message string) errInternal {
// 	return errInternal{
// 		Message:    message,
// 		StatusCode: http.StatusInternalServerError,
// 	}
// }

// type errNotFound struct {
// 	Message    string `json:"message"`
// 	StatusCode int    `json:"status_code"`
// }

// func (enf errNotFound) Name() string {
// 	return "error.notFound"
// }

// func (enf errNotFound) Error() string {
// 	return enf.Message
// }

// func (enf errNotFound) ErrorMap() map[string]string {
// 	return nil
// }

// func (enf errNotFound) SetError(message string) {
// 	enf.Message = message
// }

// func (enf errNotFound) Status() int {
// 	return enf.StatusCode
// }

// func NewNotFound(message string) errNotFound {
// 	return errNotFound{
// 		Message:    message,
// 		StatusCode: http.StatusNotFound,
// 	}
// }

// type errBadRequest struct {
// 	Message    string `json:"message"`
// 	StatusCode int    `json:"status_code"`
// }

// func (ebr errBadRequest) Name() string {
// 	return "error.badRequest"
// }

// func (ebr errBadRequest) Error() string {
// 	return ebr.Message
// }

// func (ebr errBadRequest) ErrorMap() map[string]string {
// 	return nil
// }

// func (ebr errBadRequest) SetError(message string) {
// 	ebr.Message = message
// }

// func (ebr errBadRequest) Status() int {
// 	return ebr.StatusCode
// }

// func NewBadRequest(message string) errBadRequest {
// 	return errBadRequest{
// 		Message:    message,
// 		StatusCode: http.StatusBadRequest,
// 	}
// }

// type errValidation struct {
// 	Message    string            `json:"message"`
// 	StatusCode int               `json:"status_code"`
// 	Errors     map[string]string `json:"errors,omitempty"`
// }

// func (ev errValidation) Name() string {
// 	return "error.valdation"
// }

// func (ev errValidation) Error() string {
// 	return ev.Message
// }

// func (ev errValidation) ErrorMap() map[string]string {
// 	return ev.Errors
// }

// func (ev errValidation) SetError(message string) {
// 	ev.Message = message
// }

// func (ev errValidation) Status() int {
// 	return ev.StatusCode
// }

// func NewValidation(errors map[string]string) errValidation {
// 	return errValidation{
// 		Message:    l.T("Validation failed"),
// 		Errors:     errors,
// 		StatusCode: http.StatusUnprocessableEntity,
// 	}
// }
