package frontend

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
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

	ClientID       yeahapi.ClientID
	AuthService    yeahapi.AuthService
	ListingService yeahapi.ListingService
	UserService    yeahapi.UserService
	CQRSService    yeahapi.CQRSService
	CookieService  CookieService
}

func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		server: &http.Server{},
	}

	s.server.Handler = http.HandlerFunc(s.serveHTTP)
	s.mux.Handle("/assets/", http.StripPrefix("/assets/", hashfs.FileServer(assets.FS)))

	gob.Register(&yeahapi.Flash{})
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
	h = loadFlash(h)
	if err := h(w, r); err != nil {
		fmt.Println(err)
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func routes(routeMap map[string]Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		handler := routeMap[r.Method]
		if handler != nil {
			return handler(w, r)
		}

		return nil
	}
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

func getIP(r *http.Request) string {
	addr := r.Header.Get("X-Forwarded-For")
	if len(addr) == 0 {
		addr = r.Header.Get("X-Real-Ip")
	}

	if len(addr) == 0 {
		addr = r.RemoteAddr
	}

	host, _, _ := net.SplitHostPort(addr)

	return host
}

func setFlash(w http.ResponseWriter, flash yeahapi.Flash) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(&flash); err != nil {
		return err
	}

	encoded := base64.URLEncoding.EncodeToString(buf.Bytes())
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    encoded,
		HttpOnly: true,
	})

	return nil
}

func errFlash(w http.ResponseWriter, err error) error {
	message := "Internal server error"
	if e, ok := err.(*yeahapi.Error); ok {
		message = yeahapi.ErrorMessage(e)
	}

	return setFlash(w, yeahapi.Flash{Kind: yeahapi.ErrFlashKind, Message: message})
}

func loadFlash(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if cookie, _ := r.Cookie("flash"); cookie != nil {
			var flash yeahapi.Flash
			decoded, err := base64.URLEncoding.DecodeString(cookie.Value)
			if err != nil {
				return err
			}
			if err := gob.NewDecoder(bytes.NewReader(decoded)).Decode(&flash); err != nil {
				return err
			}
			// clean up
			setFlash(w, yeahapi.Flash{})
			r = r.WithContext(yeahapi.NewContextWithFlash(r.Context(), flash))
		}

		return next(w, r)
	}
}
