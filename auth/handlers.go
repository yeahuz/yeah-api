package auth

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/yeahuz/yeah-api/auth/credential"
	"github.com/yeahuz/yeah-api/auth/otp"
	"github.com/yeahuz/yeah-api/client"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/cqrs"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/user"
)

var (
	tos            = termsOfService{Text: "this is a terms of service"}
	signupRequired = authorizationSignUpRequired{TermsOfService: tos}
	userId         = "10"
)

func HandleSendPhoneCode(cmdSender cqrs.Sender) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var phoneCodeData phoneCodeData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&phoneCodeData); err != nil {
			return err
		}

		if err := phoneCodeData.validate(); err != nil {
			return err
		}

		otp, err := otp.New(time.Minute * 15)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := otp.Save(ctx, phoneCodeData.PhoneNumber); err != nil {
			return err
		}

		cmd := cqrs.NewSendPhoneCodeCommand(phoneCodeData.PhoneNumber, otp.Code)
		if err := cmdSender.Send(ctx, cmd); err != nil {
			return errors.Internal
		}

		sentCode := sentCode{Hash: otp.Hash, Type: sentCodeSms{Length: len(otp.Code)}}
		return c.JSON(w, http.StatusOK, sentCode)
	}
}

func HandleSendEmailCode(cmdSender cqrs.Sender) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var emailCodeData emailCodeData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&emailCodeData); err != nil {
			return err
		}

		if err := emailCodeData.validate(); err != nil {
			return err
		}

		otp, err := otp.New(time.Minute * 15)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := otp.Save(ctx, emailCodeData.Email); err != nil {
			return err
		}

		cmd := cqrs.NewSendEmailCodeCommand(emailCodeData.Email, otp.Code)
		if err := cmdSender.Send(ctx, cmd); err != nil {
			return errors.Internal
		}

		sentCode := sentCode{Hash: otp.Hash, Type: sentCodeEmail{Length: len(otp.Code)}}
		return c.JSON(w, http.StatusOK, sentCode)
	}
}

func HandleSignInWithEmail() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var signInData signInEmailData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&signInData); err != nil {
			return err
		}

		if err := signInData.validate(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		otp, err := otp.GetByHash(ctx, signInData.Hash, false)
		if err != nil {
			return err
		}

		if err := otp.VerifyHash([]byte(signInData.Email + otp.Code)); err != nil {
			return err
		}

		if err := otp.Verify(signInData.Code); err != nil {
			return err
		}

		if err := otp.Confirm(ctx); err != nil {
			return err
		}

		u, err := user.GetByEmail(ctx, signInData.Email)
		if err != nil {
			if e.As(err, &errors.NotFound) {
				return c.JSON(w, http.StatusOK, signupRequired)
			}
			return err
		}

		client := r.Context().Value("client").(*client.Client)

		sess, err := newSession(u.ID, client.ID, r.UserAgent(), getIP(r))
		if err != nil {
			return err
		}

		if err := sess.save(ctx); err != nil {
			return err
		}

		authorization := authorization{User: u, Session: sess}
		return c.JSON(w, http.StatusOK, authorization)
	}
}

func HandleSignInWithPhone() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var signInData signInPhoneData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&signInData); err != nil {
			return err
		}

		if err := signInData.validate(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		otp, err := otp.GetByHash(ctx, signInData.Hash, false)
		if err != nil {
			return err
		}

		if err := otp.VerifyHash([]byte(signInData.PhoneNumber + otp.Code)); err != nil {
			return err
		}

		if err := otp.Verify(signInData.Code); err != nil {
			return err
		}

		if err := otp.Confirm(ctx); err != nil {
			return err
		}

		u, err := user.GetByPhone(ctx, signInData.PhoneNumber)
		if err != nil {
			if e.As(err, &errors.NotFound) {
				return c.JSON(w, http.StatusOK, signupRequired)
			}
			return err
		}

		client := r.Context().Value("client").(*client.Client)

		sess, err := newSession(u.ID, client.ID, r.UserAgent(), getIP(r))
		if err != nil {
			return err
		}

		if err := sess.save(ctx); err != nil {
			return err
		}

		authorization := authorization{User: u, Session: sess}
		return c.JSON(w, http.StatusOK, authorization)
	}
}

func HandleSignUpWithEmail() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		var signUpData signUpEmailData
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&signUpData); err != nil {
			return err
		}

		if err := signUpData.validate(); err != nil {
			return err
		}

		otp, err := otp.GetByHash(ctx, signUpData.Hash, true)
		if err != nil {
			return err
		}

		if err := otp.VerifyHash([]byte(signUpData.Email + otp.Code)); err != nil {
			return err
		}

		if err := otp.Verify(signUpData.Code); err != nil {
			return err
		}

		u, err := user.New(user.NewUserOpts{
			FirstName:     signUpData.FirstName,
			LastName:      signUpData.LastName,
			Email:         signUpData.Email,
			EmailVerified: true,
		})

		if err != nil {
			return err
		}

		if err := u.Save(ctx); err != nil {
			return err
		}

		client := r.Context().Value("client").(*client.Client)

		sess, err := newSession(u.ID, client.ID, r.UserAgent(), getIP(r))
		if err != nil {
			return err
		}

		if err := sess.save(ctx); err != nil {
			return err
		}

		authorization := authorization{User: u, Session: sess}
		return c.JSON(w, http.StatusOK, authorization)
	}
}

func HandleSignUpWithPhone() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		var signUpData signUpPhoneData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&signUpData); err != nil {
			return err
		}

		if err := signUpData.validate(); err != nil {
			return err
		}

		otp, err := otp.GetByHash(ctx, signUpData.Hash, true)
		if err != nil {
			return err
		}

		if err := otp.VerifyHash([]byte(signUpData.PhoneNumber + otp.Code)); err != nil {
			return err
		}

		if err := otp.Verify(signUpData.Code); err != nil {
			return err
		}

		u, err := user.New(user.NewUserOpts{
			FirstName:     signUpData.FirstName,
			LastName:      signUpData.LastName,
			PhoneNumber:   signUpData.PhoneNumber,
			PhoneVerified: true,
		})

		if err != nil {
			return err
		}

		if err := u.Save(ctx); err != nil {
			return err
		}

		client := r.Context().Value("client").(*client.Client)

		sess, err := newSession(u.ID, client.ID, r.UserAgent(), getIP(r))
		if err != nil {
			return err
		}

		if err := sess.save(ctx); err != nil {
			return err
		}

		authorization := authorization{User: u, Session: sess}
		return c.JSON(w, http.StatusOK, authorization)
	}
}

func HandlePubKeyCreateRequest() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		session := r.Context().Value("session").(*Session)
		u, err := user.GetById(ctx, session.UserID)
		if err != nil {
			return err
		}

		request, err := credential.NewPubKeyCreateRequest(u.ID, u.FirstName)

		if err != nil {
			return err
		}

		if err := request.Save(ctx); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, request)
	}
}

func HandlePubKeyGetRequest() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			return errors.NewBadRequest(l.T("user_id is required"))
		}

		u, err := user.GetById(ctx, uuid.FromStringOrNil(userID))
		if err != nil {
			return err
		}

		credentials, err := credential.GetAll(ctx, u.ID)
		if err != nil {
			return err
		}

		request, err := credential.NewPubKeyGetRequest(u.ID, credentials)

		if err != nil {
			return errors.Internal
		}

		if err := request.Save(ctx); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, request)
	}
}

func HandleCreatePubKey() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var createData credential.CreatePubKeyData
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&createData); err != nil {
			return err
		}

		request, err := credential.GetRequestById(ctx, createData.ReqID)

		if err != nil {
			return err
		}

		if _, err := credential.ValidateClientData(createData.Credential.Response.ClientDataJSON, request); err != nil {
			return err
		}

		authnData, err := credential.ValidateAuthenticatorData(createData.Credential.Response.AuthenticatorData)
		if err != nil {
			return err
		}

		pubKeyCredential, err := credential.NewPubKeyCredential(&credential.PubKeyCredentialOpts{
			CredentialID:        createData.Credential.ID,
			Counter:             authnData.Counter,
			UserID:              request.UserID,
			PubKey:              createData.Credential.Response.PubKey,
			PubKeyAlg:           createData.Credential.Response.PubKeyAlg,
			Transports:          createData.Credential.Response.Transports,
			CredentialRequestID: createData.ReqID,
			Title:               createData.Title,
		})

		if err != nil {
			return err
		}

		if err := pubKeyCredential.Save(ctx); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, nil)
	}
}

func HandleVerifyPubKey() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		var assertData credential.AssertPubKeyData
		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&assertData); err != nil {
			return err
		}

		request, err := credential.GetRequestById(ctx, assertData.ReqID)

		if err != nil {
			return err
		}

		clientData, err := credential.ValidateClientData(assertData.Credential.Response.ClientDataJSON, request)

		if err != nil {
			return err
		}

		authnData, err := credential.ValidateAuthenticatorData(assertData.Credential.Response.AuthenticatorData)

		if err != nil {
			return err
		}
		credential, err := credential.GetById(ctx, assertData.Credential.ID)
		if err != nil {
			return err
		}

		if err := credential.Verify(clientData.Raw, authnData.Raw, assertData.Credential.Response.Signature); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, nil)
	}
}

func HandleCreateLoginToken() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		fmt.Printf("HERE")
		loginToken, err := newLoginToken()
		if err != nil {
			return err
		}
		return c.JSON(w, http.StatusOK, loginToken)
	}
}

func HandleAcceptLoginToken(cmdSender cqrs.Sender) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var loginTokenData loginTokenData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&loginTokenData); err != nil {
			return err
		}

		token, err := parseLoginToken(loginTokenData.Token)
		if err != nil {
			return err
		}

		if err := token.verify(); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := cmdSender.Send(ctx, cqrs.NewLoginTokenAcceptedEvent(loginTokenData.Token)); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, nil)
	}
}

func HandleRejectLoginToken(cmdSender cqrs.Sender) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var loginTokenData loginTokenData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&loginTokenData); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := cmdSender.Send(ctx, cqrs.NewLoginTokenRejectedEvent(loginTokenData.Token)); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, nil)
	}
}

func HandleScanLoginToken() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var loginTokenData loginTokenData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&loginTokenData); err != nil {
			return err
		}

		token, err := parseLoginToken(loginTokenData.Token)
		if err != nil {
			return err
		}

		if err := token.verify(); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, nil)
	}
}

func HandleLogOut() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		session := r.Context().Value("session").(*Session)

		if err := session.remove(ctx); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, nil)
	}

}

func HandleCreateOAuthFlow() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var data oAuthFlowData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return err
		}

		if err := data.validate(); err != nil {
			return err
		}

		flow := newOAuthFlow(data)
		return c.JSON(w, http.StatusOK, flow)
	}
}

func HandleSignInWithGoogle() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		var signInData signInGoogleData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&signInData); err != nil {
			return err
		}

		if err := signInData.validate(); err != nil {
			return err
		}

		conf := config.Config.GoogleOAuthConf
		tok, err := conf.Exchange(ctx, signInData.Code)
		if err != nil {
			return err
		}

		oauthClient := conf.Client(ctx, tok)
		resp, err := oauthClient.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		var info userInfo
		if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
			return err
		}

		client := r.Context().Value("client").(*client.Client)

		account, err := user.GetByAccountId(ctx, info.Sub)
		if account != nil {

			sess, err := newSession(account.UserID, client.ID, r.UserAgent(), getIP(r))
			if err != nil {
				return err
			}

			if err := sess.save(ctx); err != nil {
				return err
			}

			u, err := user.GetById(ctx, account.UserID)
			if err != nil {
				return err
			}

			authorization := authorization{User: u, Session: sess}
			return c.JSON(w, http.StatusOK, authorization)
		}

		if !e.As(err, &errors.NotFound) {
			return err
		}

		existingUser, err := user.GetByEmail(ctx, info.Email)
		if existingUser != nil {
			sess, err := newSession(account.UserID, client.ID, r.UserAgent(), getIP(r))
			if err != nil {
				return err
			}

			if _, err := existingUser.LinkAccount(ctx, "google", info.Sub); err != nil {
				return err
			}

			if err := sess.save(ctx); err != nil {
				return err
			}

			authorization := authorization{User: existingUser, Session: sess}
			return c.JSON(w, http.StatusOK, authorization)
		}

		if !e.As(err, &errors.NotFound) {
			return err
		}

		newUser, err := user.New(user.NewUserOpts{
			Email:         info.Email,
			FirstName:     info.GivenName,
			LastName:      info.FamilyName,
			EmailVerified: true,
		})

		if err != nil {
			return err
		}

		if err := newUser.Save(ctx); err != nil {
			return err
		}

		sess, err := newSession(account.UserID, client.ID, r.UserAgent(), getIP(r))
		if err != nil {
			return err
		}

		if _, err := newUser.LinkAccount(ctx, "google", info.Sub); err != nil {
			return err
		}

		if err := sess.save(ctx); err != nil {
			return err
		}

		authorization := authorization{User: newUser, Session: sess}

		return c.JSON(w, http.StatusOK, authorization)
	}
}

func HandleSignInWithTelegram() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}
