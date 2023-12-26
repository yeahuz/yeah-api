package http

import (
	"net/http"

	yeahapi "github.com/yeahuz/yeah-api"
)

type Handler func(w http.ResponseWriter, r *http.Request) error
type Server struct {
	mux  *http.ServeMux
	Addr string

	AuthService       yeahapi.AuthService
	UserService       yeahapi.UserService
	CQRSService       yeahapi.CQRSService
	CredentialService yeahapi.CredentialService
}

func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}

	s.registerAuthRoutes()
	s.registerCredentialRoutes()
	return s
}

func (s *Server) Open() error {
	// TODO: some validations
	return http.ListenAndServe(s.Addr, s.mux)
}

func JSON(w http.ResponseWriter, status int, v any) error {
	return nil
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// lang := r.Header.Get("Accept-Language")
		if e, ok := err.(yeahapi.Error); ok {
			JSON(w, http.StatusInternalServerError, e)
		}
	}
}

// if err := fn(w, r); err != nil {
// 	lang := r.Header.Get("Accept-Language")
// 	// l := localizer.Get(lang)
// 	if e, ok := err.(yeahapi.Error); ok {
// 		e.SetError(l.T(e.Error()))
// 		errorMap := e.ErrorMap()
// 		for k, v := range errorMap {
// 			errorMap[k] = l.T(v)
// 		}
// 		JSON(w, e.Status(), e)
// 		return
// 	}
// 	JSON(w, errors.Internal.StatusCode, errors.NewInternal(l.T("Internal server error")))
// }
