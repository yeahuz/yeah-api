package common

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
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

func MakeHandlerFunc(fn ApiFunc, method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			WriteJSON(w, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}

		if err := fn(w, r); err != nil {
			if e, ok := err.(ApiError); ok {
				WriteJSON(w, e.Code, e)
				return
			}
			WriteJSON(w, ErrInternal.Code, ErrInternal)
		}
	}
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
