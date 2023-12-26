package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerCredentialRoutes() {
	s.mux.Handle("POST /credentials.pubKeyCreateRequest", s.handlePubKeyCreateRequest())
	s.mux.Handle("POST /credentials.pubKeyGetRequest", s.handlePubKeyGetRequest())
	s.mux.Handle("POST /credentials.createPubKey", s.handleCreatePubKey())
	s.mux.Handle("POST /credentials.verifyPubKey", s.handleVerifyPubKey())
}

func (s *Server) handlePubKeyCreateRequest() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		session := r.Context().Value("session").(*yeahapi.Session)
		u, err := s.UserService.User(ctx, session.UserID)
		if err != nil {
			return err
		}

		request, err := s.CredentialService.PubKeyCreateRequest(ctx, u)
		if err != nil {
			return err
		}

		return JSON(w, http.StatusOK, request)
	}
}

func (s *Server) handlePubKeyGetRequest() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		session := r.Context().Value("session").(*yeahapi.Session)
		request, err := s.CredentialService.PubKeyGetRequest(ctx, session.UserID)
		if err != nil {
			return err
		}

		return JSON(w, http.StatusOK, request)
	}
}

func (s *Server) handleCreatePubKey() Handler {
	type request struct {
		ReqID      uuid.UUID                              `json:"req_id"`
		Credential yeahapi.RawPubKeyCredentialAttestation `json:"credential"`
		Title      string
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		credRequest, err := s.CredentialService.Request(ctx, req.ReqID)
		if err != nil {
			return err
		}

		if _, err := s.CredentialService.ValidateClientData(req.Credential.Response.ClientDataJSON, credRequest); err != nil {
			return err
		}

		authnData, err := s.CredentialService.ValidateAuthnData(req.Credential.Response.AuthenticatorData)
		if err != nil {
			return err
		}

		pubKeyCredential := &yeahapi.PubKeyCredential{
			CredentialID:        req.Credential.ID,
			Counter:             authnData.Counter,
			UserID:              credRequest.UserID,
			PubKey:              req.Credential.Response.PubKey,
			PubKeyAlg:           req.Credential.Response.PubKeyAlg,
			Transports:          req.Credential.Response.Transports,
			CredentialRequestID: req.ReqID,
			Title:               req.Title,
		}

		if err := s.CredentialService.CreatePubKey(ctx, pubKeyCredential); err != nil {
			return err
		}

		return JSON(w, http.StatusOK, nil)
	}
}

func (s *Server) handleVerifyPubKey() Handler {
	type request struct {
		ReqID      uuid.UUID                            `json:"req_id"`
		Credential yeahapi.RawPubKeyCredentialAssertion `json:"credential"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return err
		}

		credRequest, err := s.CredentialService.Request(ctx, req.ReqID)
		if err != nil {
			return err
		}

		clientData, err := s.CredentialService.ValidateClientData(req.Credential.Response.ClientDataJSON, credRequest)
		if err != nil {
			return err
		}

		authnData, err := s.CredentialService.ValidateAuthnData(req.Credential.Response.AuthenticatorData)
		if err != nil {
			return err
		}

		credential, err := s.CredentialService.Credential(ctx, req.Credential.ID)
		if err != nil {
			return err
		}

		if err := credential.Verify(clientData.Raw, authnData.Raw, req.Credential.Response.Signature); err != nil {
			return err
		}

		return JSON(w, http.StatusOK, nil)
	}
}
