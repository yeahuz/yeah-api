package frontend

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/serverutil/frontend/templ/auth"
)

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

func (s *Server) handleGetLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		flash := yeahapi.FlashFromContext(r.Context())
		method := r.URL.Query().Get("method")
		loginToken, err := s.AuthService.CreateLoginToken(time.Now().Add(time.Second * 45))
		if err != nil {
			fmt.Fprintf(w, "unable to create login token")
		}
		url, err := generateQRDataURL(loginToken.Token)
		if err != nil {
			fmt.Fprintf(w, "unable to generate qr data url")
		}
		return auth.Login(auth.LoginProps{Method: fallbackStr(method, "phone"), QRDataUrl: url, Flash: flash}).Render(r.Context(), w)
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
				errFlash(w, "Unable to create otp")
				//TODO: redirect
				return nil
			}
			if err := s.CQRSService.Publish(ctx, yeahapi.NewSendPhoneCodeCmd(data.phone, otp.Code)); err != nil {
				errFlash(w, "Unable to publish")
				//TODO: redirect
				return nil
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
				errFlash(w, "Unable to create otp")
				//TODO: redirect
				return nil
			}

			if err := s.CQRSService.Publish(ctx, yeahapi.NewSendEmailCodeCmd(data.email, otp.Code)); err != nil {
				errFlash(w, "Unable to publish")
				//TODO: redirect
				return nil
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
