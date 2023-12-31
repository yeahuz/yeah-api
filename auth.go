package yeahapi

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    UserID    `json:"-"`
	Active    bool      `json:"-"`
	ClientID  ClientID  `json:"-"`
	UserAgent string    `json:"-"`
	IP        string    `json:"-"`
}

type Auth struct {
	User    *User    `json:"user"`
	Session *Session `json:"session"`
}

type Otp struct {
	ID         uuid.UUID
	Code       string
	Hash       string
	Confirmed  bool
	ExpiresAt  time.Time
	Identifier string
}

type AuthService interface {
	CreateOtp(ctx context.Context, otp *Otp) (*Otp, error)
	VerifyOtp(ctx context.Context, otp *Otp) error
	Otp(ctx context.Context, hash string, confirmed bool) (*Otp, error)
	CreateAuth(ctx context.Context, auth *Auth) (*Auth, error)
	DeleteAuth(ctx context.Context, sessionID uuid.UUID) error
	Session(ctx context.Context, sessionID uuid.UUID) (*Session, error)
}
