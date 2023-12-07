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

func JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
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
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != method {
			JSON(w, http.StatusMethodNotAllowed, errors.NewMethodNotAllowed(l.T("Method not allowed")))
			return
		}

		if err := fn(w, r); err != nil {
			if e, ok := err.(errors.AppError); ok {
				e.SetError(l.T(e.Error()))
				errorMap := e.ErrorMap()
				for k, v := range errorMap {
					errorMap[k] = l.T(v)
				}
				JSON(w, e.Status(), e)
				return
			}
			JSON(w, errors.Internal.StatusCode, errors.NewInternal(l.T("Internal server error")))
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
