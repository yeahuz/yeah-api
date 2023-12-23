package yeahapi

import (
	"context"
)

type CQRSConfig struct {
	NatsURL       string
	NatsAuthToken string
	Streams       map[string][]string
}

type CQRSMessage interface {
	Subject() string
}

type CQRSService interface {
	Send(ctx context.Context, message CQRSMessage) error
}
