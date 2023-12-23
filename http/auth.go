package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/cqrs"
)

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("POST /auth.sendPhoneCode", s.handleSendPhoneCode(nil))
	s.mux.Handle("POST /auth.sendEmailCode", s.handleSendEmailCode(nil))
}

func (s *Server) handleSendPhoneCode(cmdSender cqrs.Sender) common.ApiFunc {
	const op yeahapi.Op = "auth.handleSendPhoneCode"
	type request struct {
		PhoneNumber string `json:"phone_number"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		// if err := req.validate(); err != nil {
		// 	return err
		// }

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		otp, err := s.AuthService.CreateOtp(ctx, time.Minute*15, req.PhoneNumber)
		if err != nil {
			return err
		}

		if err := s.CQRSService.Send(ctx, cqrs.NewSendPhoneCodeCommand(req.PhoneNumber, otp.Code)); err != nil {
			return yeahapi.E(op, err)
		}

		sentCode := sentCode{Type: sentCodeSms{Length: len(otp.Code)}, Hash: otp.Hash}
		return common.JSON(w, http.StatusOK, sentCode)
	}
}

func (s *Server) handleSendEmailCode(cmdSender cqrs.Sender) common.ApiFunc {
	const op yeahapi.Op = "auth.handleSendEmailCode"
	type request struct {
		Email string `json:"email"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		// if err := emailCodeData.validate(); err != nil {
		// 	return err
		// }

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		otp, err := s.AuthService.CreateOtp(ctx, time.Minute*15, req.Email)
		if err != nil {
			return err
		}

		if err := s.CQRSService.Send(ctx, cqrs.NewSendEmailCodeCommand(req.Email, otp.Code)); err != nil {
			return yeahapi.E(op, err)
		}

		sentCode := sentCode{Type: sentCodeEmail{Length: len(otp.Code)}, Hash: otp.Hash}
		return common.JSON(w, http.StatusOK, sentCode)
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
