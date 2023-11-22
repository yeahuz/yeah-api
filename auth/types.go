package auth

import (
	"time"

	"github.com/yeahuz/yeah-api/user"
)

type PhoneCodeData struct {
	PhoneNumber string `json:"phone_number"`
}

type EmailCodeData struct {
	email string `json:"email"`
}

type SentCode struct {
	Kind string `json:"type"`
	Hash string `json:"hash"`
}

func (sc SentCode) Name() string {
	return "auth.sentCode"
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
