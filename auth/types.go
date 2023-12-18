package auth

import (
	"encoding/json"
	"time"

	"github.com/yeahuz/yeah-api/user"
)

type session struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Active    bool   `json:"-"`
	ClientID  string `json:"-"`
	UserAgent string `json:"-"`
	IP        string `json:"-"`
}

type sentCodeType interface{}

type loginToken struct {
	sig       []byte
	payload   []byte
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type loginTokenData struct {
	Token string `json:"token"`
}

type phoneCodeData struct {
	PhoneNumber string `json:"phone_number"`
}

type emailCodeData struct {
	Email string `json:"email"`
}

type sentCode struct {
	Type sentCodeType `json:"type"`
	Hash string       `json:"hash"`
}

type sentCodeSms struct {
	Length int `json:"length"`
}

type sentCodeEmail struct {
	Length int `json:"length"`
}

type signInData struct {
	Code string `json:"code"`
	Hash string `json:"hash"`
}

type signUpData struct {
	signInData
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SignInPhoneData struct {
	signInData
	PhoneNumber string `json:"phone_number"`
}

type signInEmailData struct {
	signInData
	Email string `json:"email"`
}

type termsOfService struct {
	Text string `json:"text"`
}

type authorizationSignUpRequired struct {
	TermsOfService termsOfService `json:"terms_of_service"`
}

type authorization struct {
	User    *user.User `json:"user"`
	Session *session   `json:"session"`
}

type signUpEmailData struct {
	signUpData
	Email string `json:"email"`
}

type signUpPhoneData struct {
	signUpData
	PhoneNumber string `json:"phone_number"`
}

type createOAuthFlowData struct {
	Provider providerName `json:"provider"`
}

type providerName string

const (
	providerGoogle   providerName = "google"
	providerTelegram providerName = "telegram"
)

type provider struct {
	name      providerName
	logoUrl   string
	active    bool
	createdAt time.Time
	updatedAt time.Time
}

func (a authorization) MarshalJSON() ([]byte, error) {
	type Alias authorization
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.authorization",
		Alias: Alias(a),
	})
}

func (ar authorizationSignUpRequired) MarshalJSON() ([]byte, error) {
	type Alias authorizationSignUpRequired
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.authorizationSignUpRequired",
		Alias: Alias(ar),
	})
}

func (scs sentCodeSms) MarshalJSON() ([]byte, error) {
	type Alias sentCodeSms
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.sentCodeSms",
		Alias: Alias(scs),
	})
}

func (sce sentCodeEmail) MarshalJSON() ([]byte, error) {
	type Alias sentCodeEmail
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.sentCodeEmail",
		Alias: Alias(sce),
	})
}

func (sc sentCode) MarshalJSON() ([]byte, error) {
	type Alias sentCode
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.sentCode",
		Alias: Alias(sc),
	})
}

func (lt loginToken) MarshalJSON() ([]byte, error) {
	type Alias loginToken
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.loginToken",
		Alias: Alias(lt),
	})
}
