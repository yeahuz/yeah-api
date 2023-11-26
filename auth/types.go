package auth

import (
	"encoding/json"
	"time"

	"github.com/yeahuz/yeah-api/user"
)

type SentCodeType interface{}

type PhoneCodeData struct {
	PhoneNumber string `json:"phone_number"`
}

type EmailCodeData struct {
	Email string `json:"email"`
}

type SentCodeSms struct {
	Length int `json:"length"`
}

type SentCodeEmail struct {
	Length int `json:"length"`
}

type SentCode struct {
	Type SentCodeType `json:"type"`
	Hash string       `json:"hash"`
}

func (scs SentCodeSms) MarshalJSON() ([]byte, error) {
	type Alias SentCodeSms
	return json.Marshal(struct {
		Typ string `json:"_"`
		Alias
	}{
		Typ:   "auth.sentCodeSms",
		Alias: Alias(scs),
	})
}

func (sce SentCodeEmail) MarshalJSON() ([]byte, error) {
	type Alias SentCodeEmail
	return json.Marshal(struct {
		Typ string `json:"_"`
		Alias
	}{
		Typ:   "auth.sentCodeEmail",
		Alias: Alias(sce),
	})
}

func (sc SentCode) MarshalJSON() ([]byte, error) {
	type Alias SentCode
	return json.Marshal(struct {
		Typ string `json:"_"`
		Alias
	}{
		Typ:   "auth.sentCode",
		Alias: Alias(sc),
	})
}

type SignInData struct {
	Code string `json:"code"`
	Hash string `json:"hash"`
}

type SignInPhoneData struct {
	SignInData
	PhoneNumber string `json:"phone_number"`
}

type SignInEmailData struct {
	SignInData
	Email string `json:"email"`
}

type TermsOfService struct {
	Text string `json:"text"`
}

type AuthorizationRequired struct {
	TermsOfService TermsOfService `json:"terms_of_service"`
}

type Authorization struct {
	user user.User
}

func (a Authorization) MarshalJSON() ([]byte, error) {
	type Alias Authorization
	return json.Marshal(struct {
		Typ string `json:"_"`
		Alias
	}{
		Typ:   "auth.authorization",
		Alias: Alias(a),
	})
}

func (ar AuthorizationRequired) MarshalJSON() ([]byte, error) {
	type Alias AuthorizationRequired
	return json.Marshal(struct {
		Typ string `json:"_"`
		Alias
	}{
		Typ:   "auth.authorizationSignUpRequired",
		Alias: Alias(ar),
	})
}

type SignUpData struct {
	phoneNumber   string
	phoneCodeHash string
	firstName     string
	lastName      string
}

type providerName string

const (
	providerGoogle   providerName = "google"
	providerTelegram providerName = "telegram"
)

type Provider struct {
	name      providerName
	logoUrl   string
	active    bool
	createdAt time.Time
	updatedAt time.Time
}
