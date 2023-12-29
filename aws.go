package yeahapi

import "github.com/nats-io/nats.go/jetstream"

type EmailService interface {
	SendEmailCode(m jetstream.Msg) error
}
