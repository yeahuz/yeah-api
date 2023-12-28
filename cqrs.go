package yeahapi

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	sendPhoneCode = "auth.sendPhoneCode"
	sendEmailCode = "auth.sendEmailCode"
	emailCodeSent = "auth.emailCodeSent"
	phoneCodeSent = "auth.phoneCodeSent"
)

type CQRSConfig struct {
	NatsURL       string
	NatsAuthToken string
	Streams       map[string][]string
}

type CQRSMessage interface {
	Subject() string
}

type CQRSHandler func(message jetstream.Msg) error

type CQRSService interface {
	Publish(ctx context.Context, message CQRSMessage) error
	Handle(subject string, handler CQRSHandler)
	Close() error
}

type subject struct {
	subject string
}

func (s subject) Subject() string {
	return s.subject
}

type SendPhoneCodeCmd struct {
	subject
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

type EmailCodeSentEvent struct {
	subject
	Email string `json:"email"`
}

type PhoneCodeSentEvent struct {
	subject
	PhoneNumber string `json:"phone_number"`
}

type SendEmailCodeCmd struct {
	subject
	Email string `json:"email"`
	Code  string `json:"code"`
}

func NewSendPhoneCodeCmd(phoneNumber string, code string) SendPhoneCodeCmd {
	return SendPhoneCodeCmd{
		subject:     subject{sendPhoneCode},
		PhoneNumber: phoneNumber,
		Code:        code,
	}
}

func NewSendEmailCodeCmd(email string, code string) SendEmailCodeCmd {
	return SendEmailCodeCmd{
		subject: subject{sendEmailCode},
		Email:   email,
		Code:    code,
	}
}

func NewEmailCodeSentEvent(email string) EmailCodeSentEvent {
	return EmailCodeSentEvent{
		subject: subject{emailCodeSent},
		Email:   email,
	}
}

func NewPhoneCodeSentEvent(phoneNumber string) PhoneCodeSentEvent {
	return PhoneCodeSentEvent{
		subject:     subject{phoneCodeSent},
		PhoneNumber: phoneNumber,
	}
}
