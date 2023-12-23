package http

import (
	"net/http"

	yeahapi "github.com/yeahuz/yeah-api"
)

type Server struct {
	mux *http.ServeMux

	AuthService yeahapi.AuthService
	UserService yeahapi.UserService
	CQRSService yeahapi.CQRSService
}

func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}

	s.registerAuthRoutes()
	return s
}
