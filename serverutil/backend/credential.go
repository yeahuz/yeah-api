package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerCredentialRoutes() {
	s.mux.Handle("/credentials.pubKeyCreateRequest", post(s.userOnly(s.handlePubKeyCreateRequest())))
	s.mux.Handle("/credentials.pubKeyGetRequest", post(s.clientOnly(s.handlePubKeyGetRequest())))
	s.mux.Handle("/credentials.createPubKey", post(s.userOnly(s.handleCreatePubKey())))
	s.mux.Handle("/credentials.verifyPubKey", post(s.clientOnly(s.handleVerifyPubKey())))
}

func (s *Server) handlePubKeyCreateRequest() Handler {
	const op yeahapi.Op = "http/credentials.handlePubKeyCreateRequest"
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		session := yeahapi.SessionFromContext(r.Context())
		u, err := s.UserService.User(ctx, session.UserID)
		if err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		request, err := s.CredentialService.PubKeyCreateRequest(ctx, u)
		if err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		return JSON(w, r, http.StatusOK, request)
	}
}

func (s *Server) handlePubKeyGetRequest() Handler {
	const op yeahapi.Op = "http/credentials.handlePubKeyGetRequest"
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		session := yeahapi.SessionFromContext(r.Context())
		request, err := s.CredentialService.PubKeyGetRequest(ctx, session.UserID)
		if err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		return JSON(w, r, http.StatusOK, request)
	}
}

func (s *Server) handleCreatePubKey() Handler {
	const op yeahapi.Op = "http/credentials.handleCreatePubKey"
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
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				return yeahapi.E(op, err, fmt.Sprintf("Credential request with id %s not found", req.ReqID))
			}
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		if _, err := s.CredentialService.ValidateClientData(req.Credential.Response.ClientDataJSON, credRequest); err != nil {
			return yeahapi.E(op, err, "Unable to validate client date")
		}

		authnData, err := s.CredentialService.ValidateAuthnData(req.Credential.Response.AuthenticatorData)
		if err != nil {
			return yeahapi.E(op, err, "Unable to validate authnticator data")
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
			return yeahapi.E(op, err, "Unable to save pubkey. Please, try again later")
		}

		return JSON(w, r, http.StatusOK, nil)
	}
}

func (s *Server) handleVerifyPubKey() Handler {
	const op yeahapi.Op = "http/credentials.handleVerifyPubKey"
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
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				return yeahapi.E(op, err, fmt.Sprintf("Credential request with id %s not found", req.ReqID))
			}
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		clientData, err := s.CredentialService.ValidateClientData(req.Credential.Response.ClientDataJSON, credRequest)
		if err != nil {
			return yeahapi.E(op, err, "Unable to validate client date")
		}

		authnData, err := s.CredentialService.ValidateAuthnData(req.Credential.Response.AuthenticatorData)
		if err != nil {
			return yeahapi.E(op, err, "Unable to validate authnticator data")
		}

		credential, err := s.CredentialService.Credential(ctx, req.Credential.ID)

		if err != nil {
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				return yeahapi.E(op, err, "Credential with not found")
			}
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		if err := credential.Verify(clientData.Raw, authnData.Raw, req.Credential.Response.Signature); err != nil {
			return yeahapi.E(op, err, "Couldn't verify the credential")
		}

		return JSON(w, r, http.StatusOK, nil)
	}
}
