package frontend

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth/login", get(s.handleGetLogin))
}

func (s *Server) handleGetLogin() {
}
