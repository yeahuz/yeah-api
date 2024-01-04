package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	yeahapi "github.com/yeahuz/yeah-api"
)

const ShutdownTimeout = 1 * time.Second

type Handler func(w http.ResponseWriter, r *http.Request) error

type Server struct {
	mux    *http.ServeMux
	server *http.Server
	ln     net.Listener
	Addr   string

	AuthService       yeahapi.AuthService
	UserService       yeahapi.UserService
	CQRSService       yeahapi.CQRSService
	CredentialService yeahapi.CredentialService
	LocalizerService  yeahapi.LocalizerService
	ClientService     yeahapi.ClientService
	ListingService    yeahapi.ListingService
	KVService         yeahapi.KVService
}

type errorResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		server: &http.Server{},
	}

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.registerAuthRoutes()
	s.registerCredentialRoutes()
	return s
}

func (s *Server) Open() (err error) {
	// TODO: some validations
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

	if s.CQRSService != nil {
		err = s.CQRSService.Close()
	}

	err = s.server.Shutdown(ctx)
	return err
}

func JSON(w http.ResponseWriter, r *http.Request, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		resp := &errorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}

		if e, ok := err.(*yeahapi.Error); ok {
			resp.Message = yeahapi.ErrorMessage(e)
			resp.StatusCode = errStatusCode(yeahapi.ErrorKind(e))
			JSON(w, r, errStatusCode(e.Kind), resp)
			return
		}

		JSON(w, r, http.StatusInternalServerError, resp)
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

type ok interface {
	Ok() error
}

func decode(r *http.Request, v ok) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return yeahapi.E("json: parsing error")
	}

	return v.Ok()
}

func (s *Server) clientOnly(next Handler) Handler {
	const op yeahapi.Op = "http/server.ClientOnly"
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		clientId, err := uuid.FromString(r.Header.Get("X-Client-Id"))
		if err != nil {
			return yeahapi.E(op, yeahapi.EUnathorized, "X-Client-Id header is missing or invalid")
		}

		client, err := s.ClientService.Client(ctx, yeahapi.ClientID{clientId})

		if err != nil {
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				return yeahapi.E(op, err, fmt.Sprintf("Client with id %s not found", clientId))
			}
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		clientSecret := r.Header.Get("X-Client-Secret")

		if err := s.ClientService.VerifySecret(client, clientSecret); err != nil {
			return yeahapi.E(op, err, "Invalid client secret")
		}

		r = r.WithContext(yeahapi.NewContextWithClient(r.Context(), client))

		return next(w, r)
	}
}

func (s *Server) userOnly(next Handler) Handler {
	const op yeahapi.Op = "http/server.userOnly"
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		sessionId, err := uuid.FromString(r.Header.Get("X-Session-Id"))
		if err != nil {
			return yeahapi.E(op, yeahapi.EUnathorized, "X-Session-Id header is missing or invalid")
		}

		session, err := s.AuthService.Session(ctx, sessionId)
		if err != nil {
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				return yeahapi.E(op, err, fmt.Sprintf("Session with id %s not found", sessionId))
			}
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}
		if !session.Active {
			return yeahapi.E(op, yeahapi.EUnathorized, "Session is not active or expired")
		}

		r = r.WithContext(yeahapi.NewContextWithSession(r.Context(), session))

		return next(w, r)
	}
}

var statusCodes = map[yeahapi.Kind]int{
	yeahapi.EInternal:         http.StatusInternalServerError,
	yeahapi.EInvalid:          http.StatusBadRequest,
	yeahapi.EPermission:       http.StatusForbidden,
	yeahapi.EUnathorized:      http.StatusUnauthorized,
	yeahapi.ENotFound:         http.StatusNotFound,
	yeahapi.EFound:            http.StatusConflict,
	yeahapi.ENotImplemented:   http.StatusNotImplemented,
	yeahapi.EMethodNotAllowed: http.StatusMethodNotAllowed,
	yeahapi.EOther:            http.StatusInternalServerError,
}

func post(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return yeahapi.E(yeahapi.EMethodNotAllowed)
		}

		return next(w, r)
	}
}

func get(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return yeahapi.E(yeahapi.EMethodNotAllowed)
		}

		return next(w, r)
	}
}

func errStatusCode(kind yeahapi.Kind) int {
	if v, ok := statusCodes[kind]; ok {
		return v
	}

	return http.StatusInternalServerError
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) == 0 {
		ip = r.Header.Get("X-Real-Ip")
	}
	if len(ip) == 0 {
		ip = r.RemoteAddr
	}

	return ip
}
