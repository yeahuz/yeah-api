package http

import (
	"encoding/json"
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
	LocalizerService  yeahapi.LocalizerService
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

func JSON(w http.ResponseWriter, r *http.Request, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func POST(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost {
			return yeahapi.E(yeahapi.EMethodNotAllowed)
		}

		return next(w, r)
	}
}

func GET(next Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodGet {
			return yeahapi.E(yeahapi.EMethodNotAllowed)
		}

		return next(w, r)
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// lang := r.Header.Get("Accept-Language")
		if e, ok := err.(yeahapi.Error); ok {
			JSON(w, r, errStatusCode(e.Kind), e)
			return
		}

		JSON(w, r, http.StatusInternalServerError, nil)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

var statusCodes = map[yeahapi.Kind]int{
	yeahapi.EInternal:         http.StatusInternalServerError,
	yeahapi.EInvalid:          http.StatusBadRequest,
	yeahapi.EPermission:       http.StatusForbidden,
	yeahapi.EUnathorized:      http.StatusUnauthorized,
	yeahapi.ENotExist:         http.StatusNotFound,
	yeahapi.EExist:            http.StatusConflict,
	yeahapi.ENotImplemented:   http.StatusNotImplemented,
	yeahapi.EMethodNotAllowed: http.StatusMethodNotAllowed,
	yeahapi.EOther:            http.StatusInternalServerError,
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
