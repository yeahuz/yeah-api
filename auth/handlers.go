package auth

import (
	"encoding/json"
	"net/http"

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

	if err != nil {
		return c.ErrInternal
	}

	if err := emailCodeData.validate(); err != nil {
		return err
	}

	return nil
}

func HandleSignIn(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func HandleSignUp(w http.ResponseWriter, r *http.Request) error {
	return nil
}
