package common

import (
	"net/http"

	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

type ApiFunc func(w http.ResponseWriter, r *http.Request) error

func (fn ApiFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		lang := r.Header.Get("Accept-Language")
		l := localizer.Get(lang)
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
}
