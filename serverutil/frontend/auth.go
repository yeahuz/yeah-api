package frontend

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/skip2/go-qrcode"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/serverutil/frontend/templ/auth"
)

func Loading() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<div id='loading'>Loading</div>")
		w.(http.Flusher).Flush()
		return err
	})
}

func Header() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<div>This is the header</div>")
		w.(http.Flusher).Flush()
		return err
	})
}

func Footer() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<div>This is the footer</div>")
		w.(http.Flusher).Flush()
		return err
	})
}

func Content(ch chan struct{}) templ.Component {
	go func() {
		time.Sleep(time.Second * 2)
		ch <- struct{}{}
	}()

	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<div id='content'>Content</div>
    <script>
      let content = document.getElementById('content');
      let loading = document.getElementById('loading');
      loading.replaceWith(content);
    </script>`)
		return err
	})
}

type SuspendibleComponentFunc func(ch chan struct{}) templ.Component

type suspense struct {
	ch       chan struct{}
	fallback templ.Component
	content  templ.Component
}

func (s suspense) Suspend() <-chan struct{} {
	return s.ch
}

func Suspense(fallback templ.Component, content SuspendibleComponentFunc) *suspense {
	ch := make(chan struct{})
	return &suspense{
		ch:       ch,
		fallback: fallback,
		content:  content(ch),
	}
}

func Page() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := Header().Render(ctx, w); err != nil {
			return err
		}

		sus := Suspense(Loading(), Content)
		if err := sus.fallback.Render(ctx, w); err != nil {
			return err
		}

		if err := Footer().Render(ctx, w); err != nil {
			return err
		}

		<-sus.Suspend()
		return sus.content.Render(ctx, w)
	})
}

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth/login", routes(map[string]Handler{
		http.MethodGet:  s.handleGetLogin(),
		http.MethodPost: s.handleLogin(),
	}))
	s.mux.Handle("/auth/login/otp", routes(map[string]Handler{
		http.MethodGet:  s.handleGetLoginCode(),
		http.MethodPost: s.handleSignin(),
	}))
	s.mux.Handle("/auth/login/info", routes(map[string]Handler{
		http.MethodGet:  s.handleGetLoginInfo(),
		http.MethodPost: s.handleSignup(),
	}))
}

func fallbackStr(str, fallback string) string {
	if str == "" {
		return fallback
	}
	return str
}

func (s *Server) handleGetLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		method := r.URL.Query().Get("method")
		q, err := qrcode.New("https://google.com", qrcode.Highest)
		q.DisableBorder = true
		png, err := q.PNG(256)
		if err != nil {
			fmt.Fprintf(w, "Unable to generate qr code")
		}
		b64 := base64.RawStdEncoding.EncodeToString(png)
		url := "data:image/png;base64," + b64
		return auth.Login(auth.LoginProps{Method: fallbackStr(method, "phone"), QRDataUrl: url}).Render(r.Context(), w)
	}
}

func (s *Server) handleGetLoginCode() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		cookieValue, err := s.CookieService.ReadCookie(r, "login-data")
		if err != nil {
			return err
		}
		parts := strings.Split(cookieValue, "|")
		method := parts[0]
		identifier := parts[1]
		hash := parts[2]
		return auth.LoginCode(method, identifier, hash).Render(r.Context(), w)
	}
}

func (s *Server) handleGetLoginInfo() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return auth.LoginInfo().Render(r.Context(), w)
	}
}

type loginData struct {
	method      string
	email       string
	phone       string
	countryCode string
}

func (d loginData) ok() error {
	if d.method == "email" && d.email == "" {
		return yeahapi.E(yeahapi.EInvalid, "Login email is required")
	}
	if d.method == "phone" && d.phone == "" {
		return yeahapi.E(yeahapi.EInvalid, "Login phone is required")
	}
	if d.method == "phone" && d.countryCode == "" {
		return yeahapi.E(yeahapi.EInvalid, "Phone country code is required")
	}
	return nil
}

func (s *Server) handleLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		data := loginData{
			method:      r.PostFormValue("method"),
			phone:       r.PostFormValue("phone"),
			email:       r.PostFormValue("email"),
			countryCode: r.PostFormValue("country_code"),
		}
		if err := data.ok(); err != nil {
			return err
		}

		switch data.method {
		case "phone":
			otp, err := s.AuthService.CreateOtp(ctx, &yeahapi.Otp{
				Identifier: data.phone,
				ExpiresAt:  time.Now().Add(time.Minute * 15),
			})
			if err != nil {
				fmt.Fprintf(w, "unable to create otp")
			}
			if err := s.CQRSService.Publish(ctx, yeahapi.NewSendPhoneCodeCmd(data.phone, otp.Code)); err != nil {
				fmt.Fprintf(w, "Unable to publish")
			}
			s.CookieService.SetCookie(w, &http.Cookie{
				Name:     "login-data",
				Expires:  otp.ExpiresAt,
				Value:    "phone|" + otp.Identifier + "|" + otp.Hash,
				HttpOnly: true,
			})
			http.Redirect(w, r, "/auth/login/otp", http.StatusSeeOther)
			break
		case "email":
			otp, err := s.AuthService.CreateOtp(ctx, &yeahapi.Otp{
				Identifier: data.email,
				ExpiresAt:  time.Now().Add(time.Minute * 60),
			})
			if err != nil {
				fmt.Fprintf(w, "unable to create otp")
			}

			if err := s.CQRSService.Publish(ctx, yeahapi.NewSendEmailCodeCmd(data.email, otp.Code)); err != nil {
				fmt.Fprintf(w, "Unable to publish")
			}
			s.CookieService.SetCookie(w, &http.Cookie{
				Name:     "login-data",
				Expires:  otp.ExpiresAt,
				Value:    "email|" + otp.Identifier + "|" + otp.Hash,
				HttpOnly: true,
			})
			http.Redirect(w, r, "/auth/login/otp", http.StatusSeeOther)
			break
		default:
			break
		}
		return nil
	}
}

type signInData struct {
	loginData
	otp  string
	hash string
}

func (d signInData) ok() error {
	if err := d.loginData.ok(); err != nil {
		return err
	}
	if d.otp == "" {
		return yeahapi.E(yeahapi.EInvalid, "Otp code is required")
	}
	if d.hash == "" {
		return yeahapi.E(yeahapi.EInvalid, "Otp hash is required")
	}
	return nil
}

func (s *Server) handleSignin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		data := signInData{
			loginData: loginData{
				method:      r.PostFormValue("method"),
				phone:       r.PostFormValue("phone"),
				email:       r.PostFormValue("email"),
				countryCode: r.PostFormValue("country_code"),
			},
			otp:  r.PostFormValue("otp"),
			hash: r.PostFormValue("hash"),
		}

		if err := data.ok(); err != nil {
			fmt.Fprintf(w, "unable to validate data")
			return nil
		}

		if err := s.AuthService.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       data.hash,
			Code:       data.otp,
			Identifier: "",
		}); err != nil {
			fmt.Fprintf(w, "unable to verify otp")
			return nil
		}

		switch data.loginData.method {
		case "email":
			u, err := s.UserService.ByEmail(ctx, data.loginData.email)
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				http.Redirect(w, r, "/auth/login/info", http.StatusSeeOther)
				break
			}
			auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
				User: u,
				Session: &yeahapi.Session{
					ClientID:  s.ClientID,
					UserID:    u.ID,
					UserAgent: r.UserAgent(),
					IP:        getIP(r),
				},
			})

			if err != nil {
				fmt.Fprintf(w, "unable to create auth")
			}
			s.CookieService.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    auth.Session.ID.String(),
				HttpOnly: true,
			})
			break
		case "phone":
			u, err := s.UserService.ByEmail(ctx, data.loginData.phone)
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				http.Redirect(w, r, "/auth/login/info", http.StatusSeeOther)
				break
			}
			auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
				User: u,
				Session: &yeahapi.Session{
					ClientID:  s.ClientID,
					UserID:    u.ID,
					UserAgent: r.UserAgent(),
					IP:        getIP(r),
				},
			})

			if err != nil {
				fmt.Fprintf(w, "unable to create auth")
			}

			s.CookieService.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    auth.Session.ID.String(),
				HttpOnly: true,
			})
			break
		default:
			break
		}

		return nil
	}
}

func (s *Server) handleSignup() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}
