package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth.sendPhoneCode", post(s.handleSendPhoneCode()))
	s.mux.Handle("/auth.sendEmailCode", post(s.handleSendEmailCode()))
	s.mux.Handle("/auth.signInWithEmail", post(s.handleSignInWithEmail()))
	s.mux.Handle("/auth.signInWithPhone", post(s.handleSignInWithPhone()))
	s.mux.Handle("/auth.signUpWithEmail", post(s.handleSignUpWithEmail()))
	s.mux.Handle("/auth.signUpWithPhone", post(s.handleSignUpWithPhone()))
}

type signInData struct {
	Code string `json:"code"`
	Hash string `json:"hash"`
}

type termsOfService struct {
	Text string `json:"text"`
}

type authSignUpRequired struct {
	TermsOfService termsOfService `json:"terms_of_service"`
}

func (s *Server) handleSignInWithPhone() Handler {
	const op yeahapi.Op = "http/auth.handleSignInWithPhone"
	type request struct {
		signInData
		PhoneNumber string `json:"phone_number"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		if err := s.AuthService.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       req.Hash,
			Code:       req.Code,
			Identifier: req.PhoneNumber,
		}); err != nil {
			return yeahapi.E(op, err, "Unable to verify otp code. Make sure code and hash is correct")
		}

		u, err := s.UserService.ByPhone(ctx, req.PhoneNumber)
		if yeahapi.EIs(yeahapi.ENotExist, err) {
			return JSON(w, r, http.StatusOK, authSignUpRequired{
				TermsOfService: termsOfService{
					Text: "terms of service",
				},
			})
		} else if err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		client := yeahapi.ClientFromContext(r.Context())

		auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
			User: u,
			Session: &yeahapi.Session{
				UserID:    u.ID,
				ClientID:  client.ID,
				UserAgent: r.UserAgent(),
				IP:        getIP(r),
			},
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create a session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, auth)
	}
}

func (s *Server) handleSignInWithEmail() Handler {
	const op yeahapi.Op = "http/auth.handleSignInWithEmail"
	type request struct {
		signInData
		Email string `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		if err := s.AuthService.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       req.Hash,
			Code:       req.Code,
			Identifier: req.Email,
		}); err != nil {
			return yeahapi.E(op, err, "Unable to verify otp code. Make sure code and hash is correct")
		}

		u, err := s.UserService.ByEmail(ctx, req.Email)
		if yeahapi.EIs(yeahapi.ENotExist, err) {
			return JSON(w, r, http.StatusOK, authSignUpRequired{
				TermsOfService: termsOfService{
					Text: "terms of service",
				},
			})
		} else if err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		client := yeahapi.ClientFromContext(r.Context())

		auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
			User: u,
			Session: &yeahapi.Session{
				UserID:    u.ID,
				ClientID:  client.ID,
				UserAgent: r.UserAgent(),
				IP:        getIP(r),
			},
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create a session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, auth)
	}
}

type signUpData struct {
	signInData
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (s *Server) handleSignUpWithEmail() Handler {
	const op yeahapi.Op = "http/auth.handleSignUpWithEmail"
	type request struct {
		signUpData
		Email string `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		if err := s.AuthService.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       req.Hash,
			Code:       req.Code,
			Identifier: req.Email,
		}); err != nil {
			return yeahapi.E(op, err, "Unable to verify otp code. Make sure code and hash is correct")
		}

		u, err := s.UserService.CreateUser(ctx, &yeahapi.User{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			Email:         req.Email,
			EmailVerified: true,
		})

		if err != nil {
			// TODO: check if there is constraint violation with phone number
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		client := yeahapi.ClientFromContext(r.Context())

		auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
			User: u,
			Session: &yeahapi.Session{
				UserID:    u.ID,
				ClientID:  client.ID,
				UserAgent: r.UserAgent(),
				IP:        getIP(r),
			},
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create a session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, auth)
	}
}

func (s *Server) handleSignUpWithPhone() Handler {
	const op yeahapi.Op = "http/auth.handleSignUpWithPhone"
	type request struct {
		signUpData
		PhoneNumber string `json:"phone_number"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		if err := s.AuthService.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       req.Hash,
			Code:       req.Code,
			Identifier: req.PhoneNumber,
		}); err != nil {
			return yeahapi.E(op, err, "Unable to verify otp code. Make sure code and hash is correct")
		}

		u, err := s.UserService.CreateUser(ctx, &yeahapi.User{
			FirstName:     req.FirstName,
			LastName:      req.LastName,
			PhoneNumber:   req.PhoneNumber,
			PhoneVerified: true,
		})

		if err != nil {
			// TODO: check if there is constraint violation with phone number
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		client := yeahapi.ClientFromContext(r.Context())

		auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
			User: u,
			Session: &yeahapi.Session{
				UserID:    u.ID,
				ClientID:  client.ID,
				UserAgent: r.UserAgent(),
				IP:        getIP(r),
			},
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create a session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, auth)
	}
}

type sentCodeType interface{}

type sentCode struct {
	Type sentCodeType `json:"type"`
	Hash string       `json:"hash"`
}

type sentCodeSms struct {
	Length int `json:"length"`
}

type sentCodeEmail struct {
	Length int `json:"length"`
}

func (s *Server) handleSendPhoneCode() Handler {
	const op yeahapi.Op = "http/auth.handleSendPhoneCode"
	type request struct {
		PhoneNumber string `json:"phone_number"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		// if err := req.validate(); err != nil {
		// 	return err
		// }

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		otp, err := s.AuthService.CreateOtp(ctx, &yeahapi.Otp{
			Identifier: req.PhoneNumber,
			ExpiresAt:  time.Now().Add(time.Minute * 15),
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create otp code. Please try again")
		}

		if err := s.CQRSService.Publish(ctx, yeahapi.NewSendPhoneCodeCmd(req.PhoneNumber, otp.Code)); err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please try again after some time")
		}

		sentCode := sentCode{Type: sentCodeSms{Length: len(otp.Code)}, Hash: otp.Hash}
		return JSON(w, r, http.StatusOK, sentCode)
	}
}

func (s *Server) handleSendEmailCode() Handler {
	const op yeahapi.Op = "http/auth.handleSendEmailCode"
	type request struct {
		Email string `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		// if err := emailCodeData.validate(); err != nil {
		// 	return err
		// }

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		otp, err := s.AuthService.CreateOtp(ctx, &yeahapi.Otp{
			Identifier: req.Email,
			ExpiresAt:  time.Now().Add(time.Minute * 15),
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create otp code. Please, try again")
		}

		if err := s.CQRSService.Publish(ctx, yeahapi.NewSendEmailCodeCmd(req.Email, otp.Code)); err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again after some time")
		}

		sentCode := sentCode{Type: sentCodeEmail{Length: len(otp.Code)}, Hash: otp.Hash}
		return JSON(w, r, http.StatusOK, sentCode)
	}
}

func (s *Server) handleLogOut() Handler {
	const op yeahapi.Op = "http/auth.handleLogOut"
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		session := yeahapi.SessionFromContext(r.Context())

		if err := s.AuthService.DeleteAuth(ctx, session.ID); err != nil {
			return yeahapi.E(op, err, "Couldn't delete session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, nil)
	}
}
