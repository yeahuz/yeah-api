package auth

import (
	"encoding/json"
	"net/http"

	"github.com/yeahuz/yeah-api/common"
)

func HandleSendPhoneCode(w http.ResponseWriter, r *http.Request) error {
	var phoneCodeData PhoneCodeData
	err := json.NewDecoder(r.Body).Decode(&phoneCodeData)
	defer r.Body.Close()

	if err != nil {
		return common.ErrInternal
	}

	if err := phoneCodeData.validate(); err != nil {
		return err
	}

	result := SentCode{Kind: "SMS", Hash: "hash"}
	resp := common.Response{Object: result.Name(), Data: result}
	return common.WriteJSON(w, http.StatusOK, resp)
}

func HandleSendEmailCode(w http.ResponseWriter, r *http.Request) error {
	var emailCodeData EmailCodeData
	err := json.NewDecoder(r.Body).Decode(&emailCodeData)

	if err != nil {
		return common.ErrInternal
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
