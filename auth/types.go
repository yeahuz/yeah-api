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
	phoneNumber       string
	phoneCode         string
	phoneCodeHash     string
	emailVerification string
}

type Authorization struct {
	user user.User
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
