package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yeahuz/yeah-api/auth/otp"
	c "github.com/yeahuz/yeah-api/common"
)

var (
	termsOfService        = TermsOfService{Text: "this is a terms of service"}
	authorizationRequired = AuthorizationRequired{TermsOfService: termsOfService}
)

func HandleSendPhoneCode(w http.ResponseWriter, r *http.Request) error {
	var phoneCodeData PhoneCodeData
	err := json.NewDecoder(r.Body).Decode(&phoneCodeData)
	defer r.Body.Close()

	if err != nil {
		return c.ErrInternal
	}

	if err := phoneCodeData.validate(); err != nil {
		return err
	}

	otp := otp.New(time.Minute * 15)
	fmt.Printf("Code: %s\n", otp.Code)

	if err := otp.Save(phoneCodeData.PhoneNumber); err != nil {
		return err
	}

	sentCode := SentCode{Hash: otp.Hash, Type: SentCodeSms{Length: otp.CodeLen}}
	return c.WriteJSON(w, http.StatusOK, sentCode)
}

func HandleSendEmailCode(w http.ResponseWriter, r *http.Request) error {
	var emailCodeData EmailCodeData
	err := json.NewDecoder(r.Body).Decode(&emailCodeData)
	defer r.Body.Close()

	if err != nil {
		return c.ErrInternal
	}

	if err := emailCodeData.validate(); err != nil {
		return err
	}

	otp := otp.New(time.Minute * 15)

	if err := otp.Save(emailCodeData.Email); err != nil {
		return err
	}

	sentCode := SentCode{Hash: otp.Hash, Type: SentCodeEmail{Length: otp.CodeLen}}
	return c.WriteJSON(w, http.StatusOK, sentCode)
}

func HandleSignInWithEmail(w http.ResponseWriter, r *http.Request) error {
	var signInData SignInEmailData
	err := json.NewDecoder(r.Body).Decode(&signInData)
	defer r.Body.Close()
	if err != nil {
		return c.ErrInternal
	}

	if err := signInData.validate(); err != nil {
		return err
	}

	otp, err := otp.GetByHash(signInData.Hash)
	if err != nil {
		return err
	}

	err = otp.Verify(signInData.Code)
	if err != nil {
		return err
	}

	return c.WriteJSON(w, http.StatusOK, authorizationRequired)
}

func HandleSignInWithPhone(w http.ResponseWriter, r *http.Request) error {
	var signInData SignInPhoneData
	err := json.NewDecoder(r.Body).Decode(&signInData)
	defer r.Body.Close()
	if err != nil {
		return c.ErrInternal
	}

	if err := signInData.validate(); err != nil {
		return err
	}

	otp, err := otp.GetByHash(signInData.Hash)
	if err != nil {
		return err
	}

	if err := otp.VerifyHash([]byte(signInData.PhoneNumber + otp.Code)); err != nil {
		return err
	}

	if err := otp.Verify(signInData.Code); err != nil {
		return err
	}

	return c.WriteJSON(w, http.StatusOK, authorizationRequired)
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {
	return nil
}
