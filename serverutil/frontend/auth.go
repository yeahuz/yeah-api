package frontend

import "net/http"

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth/login", get(s.handleGetLogin()))
}

func (s *Server) handleGetLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}
