package frontend

import (
	"net/http"

	"github.com/yeahuz/yeah-api/serverutil/frontend/templ/auth"
)

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth/login", get(s.handleGetLogin()))
	s.mux.Handle("/auth/login", post(s.handleLogin()))
}

func (s *Server) handleGetLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return auth.Login().Render(r.Context(), w)
	}
}

func (s *Server) handleLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}
