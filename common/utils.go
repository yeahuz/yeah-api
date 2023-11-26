package common

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
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
			errors.MethodNotAllowed.Message = l.T(errors.MethodNotAllowed.Message)
			WriteJSON(w, http.StatusMethodNotAllowed, errors.MethodNotAllowed)
			return
		}

		if err := fn(w, r); err != nil {
			if e, ok := err.(errors.AppError); ok {
				e.SetError(l.T(e.Error()))
				errorMap := e.ErrorMap()
				for k, v := range errorMap {
					errorMap[k] = l.T(v)
				}
				WriteJSON(w, e.Status(), e)
				return
			}
			errors.Internal.Message = l.T(errors.Internal.Message)
			WriteJSON(w, errors.Internal.StatusCode, errors.Internal)
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
