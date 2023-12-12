package common

import (
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

func HandleError(fn ApiFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Header.Get("Accept-Language")
		l := localizer.Get(lang)
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

type handlerConfig struct {
	Public bool
}

func MakeHandler(fn ApiFunc, method string, opts ...func(config *handlerConfig)) http.Handler {
	config := &handlerConfig{
		Public: false,
	}

	for _, fn := range opts {
		fn(config)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.WriteHeader(http.StatusOK)
			return
		}

		lang := r.Header.Get("Accept-Language")
		l := localizer.Get(lang)

		//TODO: in the future, make all handlers deny by default!
		// if !config.Public {
		// 	JSON(w, http.StatusUnauthorized, errors.NewUnauthorized(l.T("Not authorized")))
		// 	return
		// }

		if r.Method != method {
			JSON(w, http.StatusMethodNotAllowed, errors.NewMethodNotAllowed(l.T("Method not allowed")))
			return
		}

		handler := HandleError(fn)
		handler.ServeHTTP(w, r)
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
