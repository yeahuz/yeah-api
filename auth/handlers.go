package auth

import (
	"context"
	"encoding/json"
	e "errors"
	"net/http"
	"time"

	"github.com/yeahuz/yeah-api/auth/credential"
	"github.com/yeahuz/yeah-api/auth/otp"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/cqrs"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/user"
)

var (
	termsOfService              = TermsOfService{Text: "this is a terms of service"}
	authorizationSignUpRequired = AuthorizationSignUpRequired{TermsOfService: termsOfService}
	userId                      = "10"
)

func HandleSendPhoneCode(w http.ResponseWriter, r *http.Request) error {
	var phoneCodeData PhoneCodeData
	err := json.NewDecoder(r.Body).Decode(&phoneCodeData)
	defer r.Body.Close()

	if err != nil {
		return errors.Internal
	}

	if err := phoneCodeData.validate(); err != nil {
		return err
	}

	otp := otp.New(time.Minute * 15)

	if err := otp.Save(phoneCodeData.PhoneNumber); err != nil {
		return err
	}
	cmd := cqrs.NewSendPhoneCodeCommand(phoneCodeData.PhoneNumber, otp.Code)
	if err := cqrs.Send(cmd); err != nil {
		return errors.Internal
	}

	sentCode := SentCode{Hash: otp.Hash, Type: SentCodeSms{Length: len(otp.Code)}}
	return c.JSON(w, http.StatusOK, sentCode)
}

func HandleSendEmailCode(w http.ResponseWriter, r *http.Request) error {
	var emailCodeData EmailCodeData
	err := json.NewDecoder(r.Body).Decode(&emailCodeData)
	defer r.Body.Close()

	if err != nil {
		return errors.Internal
	}

	if err := emailCodeData.validate(); err != nil {
		return err
	}

	otp := otp.New(time.Minute * 15)

	if err := otp.Save(emailCodeData.Email); err != nil {
		return err
	}

	cmd := cqrs.NewSendEmailCodeCommand(emailCodeData.Email, otp.Code)
	if err := cqrs.Send(cmd); err != nil {
		return errors.Internal
	}

	sentCode := SentCode{Hash: otp.Hash, Type: SentCodeEmail{Length: len(otp.Code)}}
	return c.JSON(w, http.StatusOK, sentCode)
}

func HandleSignInWithEmail(w http.ResponseWriter, r *http.Request) error {
	var signInData SignInEmailData
	err := json.NewDecoder(r.Body).Decode(&signInData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
	}

	if err := signInData.validate(); err != nil {
		return err
	}

	otp, err := otp.GetByHash(signInData.Hash, false)
	if err != nil {
		return err
	}

	if err := otp.VerifyHash([]byte(signInData.Email + otp.Code)); err != nil {
		return err
	}

	if err := otp.Verify(signInData.Code); err != nil {
		return err
	}

	if err := otp.Confirm(); err != nil {
		return err
	}

	u, err := user.GetByEmail(signInData.Email)
	if err != nil {
		if e.As(err, &errors.NotFound) {
			return c.JSON(w, http.StatusOK, authorizationSignUpRequired)
		}
		return err
	}

	authorization := Authorization{User: u}
	return c.JSON(w, http.StatusOK, authorization)
}

func HandleSignInWithPhone(w http.ResponseWriter, r *http.Request) error {
	var signInData SignInPhoneData
	err := json.NewDecoder(r.Body).Decode(&signInData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
	}

	if err := signInData.validate(); err != nil {
		return err
	}

	otp, err := otp.GetByHash(signInData.Hash, false)
	if err != nil {
		return err
	}

	if err := otp.VerifyHash([]byte(signInData.PhoneNumber + otp.Code)); err != nil {
		return err
	}

	if err := otp.Verify(signInData.Code); err != nil {
		return err
	}

	if err := otp.Confirm(); err != nil {
		return err
	}

	u, err := user.GetByPhone(signInData.PhoneNumber)
	if err != nil {
		if e.As(err, &errors.NotFound) {
			return c.JSON(w, http.StatusOK, authorizationSignUpRequired)
		}
		return err
	}

	authorization := Authorization{User: u}
	return c.JSON(w, http.StatusOK, authorization)
}

func HandleSignUpWithEmail(w http.ResponseWriter, r *http.Request) error {
	var signUpData SignUpEmailData
	err := json.NewDecoder(r.Body).Decode(&signUpData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
	}

	if err := signUpData.validate(); err != nil {
		return err
	}

	otp, err := otp.GetByHash(signUpData.Hash, true)
	if err != nil {
		return err
	}

	if err := otp.VerifyHash([]byte(signUpData.Email + otp.Code)); err != nil {
		return err
	}

	if err := otp.Verify(signUpData.Code); err != nil {
		return err
	}

	u := user.New(user.NewUserOpts{
		FirstName:     signUpData.FirstName,
		LastName:      signUpData.LastName,
		Email:         signUpData.Email,
		EmailVerified: true,
	})

	if err := u.Save(); err != nil {
		return err
	}

	authorization := Authorization{User: u}
	return c.JSON(w, http.StatusOK, authorization)
}

func HandleSignUpWithPhone(w http.ResponseWriter, r *http.Request) error {
	var signUpData SignUpPhoneData
	err := json.NewDecoder(r.Body).Decode(&signUpData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
	}

	if err := signUpData.validate(); err != nil {
		return err
	}

	otp, err := otp.GetByHash(signUpData.Hash, true)
	if err != nil {
		return err
	}

	if err := otp.VerifyHash([]byte(signUpData.PhoneNumber + otp.Code)); err != nil {
		return err
	}

	if err := otp.Verify(signUpData.Code); err != nil {
		return err
	}

	u := user.New(user.NewUserOpts{
		FirstName:     signUpData.FirstName,
		LastName:      signUpData.LastName,
		PhoneNumber:   signUpData.PhoneNumber,
		PhoneVerified: true,
	})

	if err := u.Save(); err != nil {
		return err
	}

	authorization := Authorization{User: u}
	return c.JSON(w, http.StatusOK, authorization)
}

// func HandleGetCredentials(w http.ResponseWriter, r *http.Request) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
// 	defer cancel()
// 	credentials, err := credential.GetAll(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return c.JSON(w, http.StatusOK, credentials)
// }

func HandlePubKeyCreateRequest(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		return errors.NewBadRequest(l.T("user_id is required"))
	}

	u, err := user.GetById(userID)
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

func HandlePubKeyGetRequest(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		return errors.NewBadRequest(l.T("user_id is required"))
	}

	u, err := user.GetById(userID)
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

func HandleCreatePubKey(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var createData credential.CreatePubKeyData
	err := json.NewDecoder(r.Body).Decode(&createData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
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

	pubKeyCredential := &credential.PubKeyCredential{
		CredentialID:        createData.Credential.ID,
		Counter:             authnData.Counter,
		UserID:              request.UserID,
		PubKey:              createData.Credential.Response.PubKey,
		PubKeyAlg:           createData.Credential.Response.PubKeyAlg,
		Transports:          createData.Credential.Response.Transports,
		CredentialRequestID: createData.ReqID,
		Title:               createData.Title,
	}

	if err := pubKeyCredential.Save(ctx); err != nil {
		return err
	}

	return c.JSON(w, http.StatusOK, nil)
}

func HandleVerifyPubKey(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var assertData credential.AssertPubKeyData
	err := json.NewDecoder(r.Body).Decode(&assertData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
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

func HandleCreateLoginToken(w http.ResponseWriter, r *http.Request) error {
	loginToken, err := newLoginToken()
	if err != nil {
		return err
	}

	return c.JSON(w, http.StatusOK, loginToken)
}

func HandleAcceptLoginToken(w http.ResponseWriter, r *http.Request) error {
	type data struct {
		Token string `json:"token"`
	}

	var d data
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return err
	}

	token, err := parseLoginToken(d.Token)
	if err != nil {
		return err
	}

	valid := token.verify()

	return c.JSON(w, http.StatusOK, valid)
}

func HandleVerifyLoginToken(w http.ResponseWriter, r *http.Request) error {
	return nil
}
