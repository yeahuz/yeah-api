package auth

import "time"

type PhoneCodeData struct {
	phoneNumber string
}

type EmailCodeData struct{ email string }

type SignInData struct {
	phoneNumber       string
	phoneCode         string
	phoneCodeHash     string
	emailVerification string
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
