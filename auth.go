package yeahapi

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type Session struct {
	ID        string   `json:"id"`
	UserID    UserID   `json:"-"`
	Active    bool     `json:"-"`
	ClientID  ClientID `json:"-"`
	UserAgent string   `json:"-"`
	IP        string   `json:"-"`
}

type Auth struct {
	User    *User    `json:"user"`
	Session *Session `json:"session"`
}

type Otp struct {
	ID        uuid.UUID
	Code      string
	Hash      string
	Confirmed bool
	ExpiresAt time.Time
}

type AuthService interface {
	CreateOtp(ctx context.Context, duration time.Duration, identifier string) (*Otp, error)
	VerifyOtp(ctx context.Context) error
	CreateAuth(ctx context.Context)
	DeleteAuth(ctx context.Context)
}
