package auth

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/yeahuz/yeah-api/user"
)

type session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Active    bool      `json:"-"`
	ClientID  uuid.UUID `json:"-"`
	UserAgent string    `json:"-"`
	IP        string    `json:"-"`
}

type userInfo struct {
	Sub        string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Email      string `json:"email"`
	Profile    string `json:"profile"`
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

type signInGoogleData struct {
	Code string `json:"code"`
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

type signInPhoneData struct {
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

type oAuthFlowData struct {
	Provider    oAuthProvider `json:"provider"`
	State       string        `json:"state"`
	RedirectURL string        `json:"redirect_url"`
}

type oAuthFlow struct {
	URL string `json:"url"`
}

type oAuthProvider string

const (
	providerGoogle oAuthProvider = "google"
)

func (o oAuthFlow) MarshalJSON() ([]byte, error) {
	type Alias oAuthFlow
	return json.Marshal(struct {
		Type string `json:"_"`
		Alias
	}{
		Type:  "auth.oAuthFlow",
		Alias: Alias(o),
	})
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
