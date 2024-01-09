package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

var emailRegex = regexp.MustCompile(`(?i)^(([^<>()[\].,;:\s@"]+(\.[^<>()[\].,;:\s@"]+)*)|(".+"))@(([^<>()[\].,;:\s@"]+\.)+[^<>()[\].,;:\s@"]{2,})$`)

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth.sendPhoneCode", post(s.clientOnly(s.handleSendPhoneCode())))
	s.mux.Handle("/auth.sendEmailCode", post(s.clientOnly(s.handleSendEmailCode())))
	s.mux.Handle("/auth.signInWithEmail", post(s.clientOnly(s.handleSignInWithEmail())))
	s.mux.Handle("/auth.signInWithPhone", post(s.clientOnly(s.handleSignInWithPhone())))
	s.mux.Handle("/auth.signUpWithEmail", post(s.clientOnly(s.handleSignUpWithEmail())))
	s.mux.Handle("/auth.signUpWithPhone", post(s.clientOnly(s.handleSignUpWithPhone())))
	s.mux.Handle("/auth.logOut", post(s.userOnly(s.handleLogOut())))
}

type sentCodeData struct {
	Code string `json:"code"`
	Hash string `json:"hash"`
}

type phoneData struct {
	PhoneNumber string `json:"phone_number"`
}

type emailData struct {
	Email string `json:"email"`
}

type termsOfService struct {
	Text string `json:"text"`
}

type authSignUpRequired struct {
	TermsOfService termsOfService `json:"terms_of_service"`
}

func (d sentCodeData) Ok() error {
	if d.Code == "" {
		return yeahapi.E(yeahapi.EInvalid, "Code is required")
	}
	if d.Hash == "" {
		return yeahapi.E(yeahapi.EInvalid, "Hash is required")
	}
	return nil
}

func (d phoneData) Ok() error {
	if d.PhoneNumber == "" {
		return yeahapi.E(yeahapi.EInvalid, "Phone number is required")
	}
	if len(d.PhoneNumber) != 13 {
		return yeahapi.E(yeahapi.EInvalid, "Phone number is invalid")
	}
	return nil
}

func (d emailData) Ok() error {
	if d.Email == "" {
		return yeahapi.E(yeahapi.EInvalid, "Email is required")
	}
	if !emailRegex.MatchString(d.Email) {
		return yeahapi.E(yeahapi.EInvalid, "Email is invalid")
	}
	return nil
}

type signInPhoneData struct {
	sentCodeData
	phoneData
}

func (d signInPhoneData) Ok() error {
	if err := d.sentCodeData.Ok(); err != nil {
		return err
	}
	if err := d.phoneData.Ok(); err != nil {
		return err
	}
	return nil
}

type authorization struct {
	*yeahapi.Auth
}

func (s *Server) handleSignInWithPhone() Handler {
	const op yeahapi.Op = "http/auth.handleSignInWithPhone"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req signInPhoneData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
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
		if yeahapi.EIs(yeahapi.ENotFound, err) {
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

		return JSON(w, r, http.StatusOK, authorization{auth})
	}
}

type signInEmailData struct {
	sentCodeData
	emailData
}

func (d signInEmailData) Ok() error {
	if err := d.sentCodeData.Ok(); err != nil {
		return err
	}

	if err := d.emailData.Ok(); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleSignInWithEmail() Handler {
	const op yeahapi.Op = "http/auth.handleSignInWithEmail"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req signInEmailData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
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
		if yeahapi.EIs(yeahapi.ENotFound, err) {
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

		return JSON(w, r, http.StatusOK, authorization{auth})
	}
}

type signUpData struct {
	sentCodeData
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (d signUpData) Ok() error {
	if err := d.sentCodeData.Ok(); err != nil {
		return err
	}
	if d.LastName == "" {
		return yeahapi.E(yeahapi.EInvalid, "Last name is required")
	}
	if d.FirstName == "" {
		return yeahapi.E(yeahapi.EInvalid, "First name is required")
	}
	return nil
}

type signUpEmailData struct {
	signUpData
	emailData
}

func (d signUpEmailData) Ok() error {
	if err := d.signUpData.Ok(); err != nil {
		return err
	}
	if err := d.emailData.Ok(); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleSignUpWithEmail() Handler {
	const op yeahapi.Op = "http/auth.handleSignUpWithEmail"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req signUpEmailData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
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

		client := yeahapi.ClientFromContext(r.Context())

		auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
			User: &yeahapi.User{
				FirstName:     req.FirstName,
				LastName:      req.LastName,
				Email:         req.Email,
				EmailVerified: true,
			},
			Session: &yeahapi.Session{
				ClientID:  client.ID,
				UserAgent: r.UserAgent(),
				IP:        getIP(r),
			},
		})

		if err != nil {
			fmt.Println(err)
			return yeahapi.E(op, err, "Couldn't create a session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, authorization{auth})
	}
}

type signUpPhoneData struct {
	signUpData
	phoneData
}

func (d signUpPhoneData) Ok() error {
	if err := d.signUpData.Ok(); err != nil {
		return err
	}
	if err := d.phoneData.Ok(); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleSignUpWithPhone() Handler {
	const op yeahapi.Op = "http/auth.handleSignUpWithPhone"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req signUpPhoneData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
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

		client := yeahapi.ClientFromContext(r.Context())

		auth, err := s.AuthService.CreateAuth(ctx, &yeahapi.Auth{
			User: &yeahapi.User{
				FirstName:     req.FirstName,
				LastName:      req.LastName,
				PhoneNumber:   req.PhoneNumber,
				PhoneVerified: true,
			},
			Session: &yeahapi.Session{
				ClientID:  client.ID,
				UserAgent: r.UserAgent(),
				IP:        getIP(r),
			},
		})

		if err != nil {
			return yeahapi.E(op, err, "Couldn't create a session. Please, try again")
		}

		return JSON(w, r, http.StatusOK, authorization{auth})
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

func (s *Server) handleSendPhoneCode() Handler {
	const op yeahapi.Op = "http/auth.handleSendPhoneCode"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req phoneData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}

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

type sentCodeEmail struct {
	Length int `json:"length"`
}

func (s *Server) handleSendEmailCode() Handler {
	const op yeahapi.Op = "http/auth.handleSendEmailCode"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req emailData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}

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

func (s sentCode) MarshalJSON() ([]byte, error) {
	type Alias sentCode
	return json.Marshal(&struct {
		Type string `json:"_"`
		*Alias
	}{
		Type:  "auth.sentCode",
		Alias: (*Alias)(&s),
	})
}

func (s sentCodeSms) MarshalJSON() ([]byte, error) {
	type Alias sentCodeSms
	return json.Marshal(&struct {
		Type string `json:"_"`
		*Alias
	}{
		Type:  "auth.sentCodeSms",
		Alias: (*Alias)(&s),
	})
}

func (s sentCodeEmail) MarshalJSON() ([]byte, error) {
	type Alias sentCodeEmail
	return json.Marshal(&struct {
		Type string `json:"_"`
		*Alias
	}{
		Type:  "auth.sentCodeEmail",
		Alias: (*Alias)(&s),
	})
}

func (a authorization) MarshalJSON() ([]byte, error) {
	type Alias authorization
	return json.Marshal(&struct {
		Type string `json:"_"`
		*Alias
	}{
		Type:  "auth.authorization",
		Alias: (*Alias)(&a),
	})
}

func (a authSignUpRequired) MarshalJSON() ([]byte, error) {
	type Alias authSignUpRequired
	return json.Marshal(&struct {
		Type string `json:"_"`
		*Alias
	}{
		Type:  "auth.authorizationSignUpRequired",
		Alias: (*Alias)(&a),
	})
}
