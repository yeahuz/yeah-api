package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yeahuz/yeah-api/auth/otp"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

func HandleSendPhoneCode(w http.ResponseWriter, r *http.Request) error {
	var phoneCodeData PhoneCodeData
	err := json.NewDecoder(r.Body).Decode(&phoneCodeData)
	defer r.Body.Close()

	l := r.Context().Value("localizer").(localizer.Localizer)

	if err != nil {
		return c.ErrInternal
	}

	if err := phoneCodeData.validate(&l); err != nil {
		return err
	}

	otp := otp.New(time.Minute * 15)

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

	l := r.Context().Value("localizer").(localizer.Localizer)

	if err := emailCodeData.validate(&l); err != nil {
		return err
	}

	otp := otp.New(time.Minute * 15)

	if err := otp.Save(emailCodeData.Email); err != nil {
		return err
	}

	sentCode := SentCode{Hash: otp.Hash, Type: SentCodeEmail{Length: otp.CodeLen}}
	return c.WriteJSON(w, http.StatusOK, sentCode)
}

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {
	return nil
}
