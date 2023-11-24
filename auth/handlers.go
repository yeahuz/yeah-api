package auth

import (
	"encoding/json"
	"fmt"
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

	result := SentCode{Kind: "SMS", Hash: "hash"}
	resp := c.Response{Object: result.Name(), Data: result}
	return c.WriteJSON(w, http.StatusOK, resp)
}

func HandleSendEmailCode(w http.ResponseWriter, r *http.Request) error {
	var emailCodeData EmailCodeData
	err := json.NewDecoder(r.Body).Decode(&emailCodeData)

	defer r.Body.Close()

	l := r.Context().Value("localizer").(localizer.Localizer)

	if err != nil {
		return c.ErrInternal
	}

	if err := emailCodeData.validate(&l); err != nil {
		return err
	}

	otp, err := otp.New(emailCodeData.Email, time.Minute*15)
	if err != nil {
		return err
	}

	fmt.Printf("Code: %s\n", otp.Code)
	fmt.Printf("Hash: %s\n", otp.Hash)

	return nil
}

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {
	return nil
}
