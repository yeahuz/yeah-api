package frontend

import (
	"net"
	"net/http"

	yeahapi "github.com/yeahuz/yeah-api"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

type Server struct {
	mux         *http.ServeMux
	server      *http.Server
	ln          net.Listener
	Addr        string
	AuthService yeahapi.AuthService
}

func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		server: &http.Server{},
	}

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	// s.registerAuthRoutes()

	return s
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func get(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
		}
		return next(w, r)
	}
}
