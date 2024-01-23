package frontend

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/benbjohnson/hashfs"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/serverutil/frontend/assets"
)

const ShutdownTimeout = 1 * time.Second

type Handler func(w http.ResponseWriter, r *http.Request) error

type Server struct {
	mux    *http.ServeMux
	server *http.Server
	ln     net.Listener
	Addr   string

	AuthService    yeahapi.AuthService
	ListingService yeahapi.ListingService
	UserService    yeahapi.UserService
}

func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		server: &http.Server{},
	}

	s.server.Handler = http.HandlerFunc(s.serveHTTP)
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", hashfs.FileServer(assets.FS)))

	s.registerAuthRoutes()

	return s
}

func (s *Server) Open() (err error) {
	fmt.Printf("Server started at %s\n", s.Addr)
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}
	go s.server.Serve(s.ln)
	return nil
}

func (s *Server) Close() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
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
			return nil
		}
		return next(w, r)
	}
}

func post(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return nil
		}

		return next(w, r)
	}
}
