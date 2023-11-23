package common

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/yeahuz/yeah-api/internal/localizer"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func NewResponse(object Object, data interface{}) Response {
	return Response{
		Object: object.Name(),
		Data:   data,
	}
}

func LocalizerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Header.Get("Accept-Language")
		l := localizer.Get(lang)
		ctx := context.WithValue(r.Context(), "localizer", l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MakeHandler(fn ApiFunc, method string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Header.Get("Accept-Language")
		l := localizer.Get(lang)
		if r.Method != method {
			ErrMethodNotAllowed.Message = l.T(ErrMethodNotAllowed.Message)
			WriteJSON(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}

		if err := fn(w, r); err != nil {
			if e, ok := err.(ApiError); ok {
				e.Message = l.T(e.Message)
				for k, v := range e.Errors {
					e.Errors[k] = l.T(v)
				}
				WriteJSON(w, e.Code, e)
				return
			}
			ErrInternal.Message = l.T(ErrInternal.Message)
			WriteJSON(w, ErrInternal.Code, ErrInternal)
		}
	})
}

func GetEnvInt(key string, fallback int) int {
	s := os.Getenv(key)

	if len(s) == 0 {
		return fallback
	}

	v, err := strconv.Atoi(s)

	if err != nil {
		return fallback
	}

	return v
}

func GetEnvStr(key, fallback string) string {
	s := os.Getenv(key)

	if len(s) == 0 {
		return fallback
	}

	return s
}
