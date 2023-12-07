package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	e "errors"
	"fmt"
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

func HandleGetCredentials(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	credentials, err := credential.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(w, http.StatusOK, credentials)
}

func HandleCredentialCreateRequest(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	createRequest, err := credential.NewCreateRequest("200", "Avazbek", "Avazbek")

	if err != nil {
		return err
	}

	if err := createRequest.Save(ctx); err != nil {
		return err
	}

	return c.JSON(w, http.StatusOK, createRequest)
}

func HandleCredentialGetRequest(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	allowedCredentials, err := credential.GetAllowedCredentials(ctx, "10")
	if err != nil {
		return err
	}

	getRequest, err := credential.NewGetRequest(allowedCredentials)
	if err != nil {
		return err
	}

	if err := getRequest.Save(ctx); err != nil {
		return err
	}

	return c.JSON(w, http.StatusOK, getRequest)
}

func HandleCredentialVerify(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func HandleCreateCredential(w http.ResponseWriter, r *http.Request) error {
	fmt.Printf("HERE\n")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var credentialData credential.CreateCredentialData
	err := json.NewDecoder(r.Body).Decode(&credentialData)
	defer r.Body.Close()
	if err != nil {
		return errors.Internal
	}

	request, err := credential.GetRequestById(ctx, credentialData.ReqID)

	if err != nil {
		return err
	}

	fmt.Printf("CLIENTDATA: %s\n", credentialData.Credential.Response.ClientDataJSON)
	if _, err := request.VerifyClientData(credentialData.Credential.Response.ClientDataJSON); err != nil {
		return err
	}

	b, err := base64.RawURLEncoding.DecodeString(credentialData.Credential.Response.AttestationObject)

	if err != nil {
		return errors.Internal
	}

	p, err := credential.ParseAttestation(b)
	if err != nil {
		return errors.Internal
	}

	_ = p
	return nil
}
