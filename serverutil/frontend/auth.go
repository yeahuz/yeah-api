package frontend

import (
	"net/http"

	"github.com/yeahuz/yeah-api/templ/auth"
)

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth/login", get(s.handleGetLogin()))
}

func (s *Server) handleGetLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return auth.Login().Render(r.Context(), w)
	}
}
