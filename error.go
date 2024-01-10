package yeahapi

import (
	"bytes"
	"fmt"

	"github.com/gofrs/uuid"
)

type Op string
type Kind int

type Error struct {
	Op       Op
	Kind     Kind
	Err      error
	Message  string
	UserID   UserID
	ClientID ClientID
}

const Separator = ":\n\t"

const (
	EOther Kind = iota
	EInvalid
	EPermission
	ENotFound
	EFound
	ENotImplemented
	EUnathorized
	EMethodNotAllowed
	EOtpCodeExpired
	EOtpHashNotMatched
	EInternal
)

func (k Kind) String() string {
	switch k {
	case EInternal:
		return "internal error"
	case EOther:
		return "other error"
	case ENotFound:
		return "item does not exit"
	case EFound:
		return "item already exists"
	case EInvalid:
		return "invalid error"
	case ENotImplemented:
		return "not implemented"
	case EUnathorized:
		return "unauthorized"
	case EMethodNotAllowed:
		return "method not allowed"
	case EOtpCodeExpired:
		return "otp code expired"
	case EOtpHashNotMatched:
		return "otp hash not matched"
	case EPermission:
		return "permission denied"
	}
	return "unknown error"
}

func (e Error) Error() string {
	b := new(bytes.Buffer)

	if e.Op != "" {
		b.WriteString(string(e.Op))
	}
	if !e.UserID.IsNil() {
		pad(b, ": ")
		b.WriteString("user ")
		b.WriteString(e.UserID.String())
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

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Message != "" {
		return e.Message
	} else if ok && e.Err != nil {
		return ErrorMessage(e.Err)
	}
	return ""
}

func ErrorKind(err error) Kind {
	if err == nil {
		return EOther
	} else if e, ok := err.(*Error); ok && e.Kind != 0 {
		return e.Kind
	} else if ok && e.Err != nil {
		return ErrorKind(e.Err)
	}
	return EOther
}

func E(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}

	e := &Error{}

	for _, arg := range args {
		switch arg := arg.(type) {
		case UserID:
			e.UserID = arg
		case ClientID:
			e.ClientID = arg
		case Op:
			e.Op = arg
		case error:
			e.Err = arg
		case Kind:
			e.Kind = arg
		case string:
			e.Message = arg
			// e.Err = Str(arg)
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
		prev.UserID.UUID = uuid.Nil
	}

	if prev.Kind == e.Kind {
		prev.Kind = EOther
	}

	if e.Kind == EOther {
		e.Kind = prev.Kind
		prev.Kind = EOther
	}

	return e
}

func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}

func EIs(kind Kind, err error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}

	if e.Kind != EOther {
		return e.Kind == kind
	}

	if e.Err != nil {
		return EIs(kind, e.Err)
	}

	return false
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

func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
}

func (e *Error) isZero() bool {
	return (e.UserID.IsNil() || e.ClientID.IsNil()) && e.Op == "" && e.Kind == 0 && e.Err == nil
}
