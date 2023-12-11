package auth

import (
	"encoding/json"
	"time"

	"github.com/yeahuz/yeah-api/user"
)

type SentCodeType interface{}

type LoginToken struct {
	sig       []byte
	payload   []byte
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

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

type SignInData struct {
	Code string `json:"code"`
	Hash string `json:"hash"`
}

type SignUpData struct {
	SignInData
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
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

type AuthorizationSignUpRequired struct {
	TermsOfService TermsOfService `json:"terms_of_service"`
}

type Authorization struct {
	User *user.User `json:"user"`
}

type SignUpEmailData struct {
	SignUpData
	Email string `json:"email"`
}

type SignUpPhoneData struct {
	SignUpData
	PhoneNumber string `json:"phone_number"`
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

func (a Authorization) MarshalJSON() ([]byte, error) {
	type Alias Authorization
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.authorization",
		Alias: Alias(a),
	})
}

func (ar AuthorizationSignUpRequired) MarshalJSON() ([]byte, error) {
	type Alias AuthorizationSignUpRequired
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.authorizationSignUpRequired",
		Alias: Alias(ar),
	})
}

func (scs SentCodeSms) MarshalJSON() ([]byte, error) {
	type Alias SentCodeSms
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.sentCodeSms",
		Alias: Alias(scs),
	})
}

func (sce SentCodeEmail) MarshalJSON() ([]byte, error) {
	type Alias SentCodeEmail
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.sentCodeEmail",
		Alias: Alias(sce),
	})
}

func (sc SentCode) MarshalJSON() ([]byte, error) {
	type Alias SentCode
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.sentCode",
		Alias: Alias(sc),
	})
}

func (lt LoginToken) MarshalJSON() ([]byte, error) {
	type Alias LoginToken
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.loginToken",
		Alias: Alias(lt),
	})
}
