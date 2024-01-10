package yeahapi

import (
	"context"

	"github.com/gofrs/uuid"
)

type UserID struct {
	uuid.UUID
}

const (
	authProviderGoogle   string = "google"
	authProviderTelegram string = "telegram"
)

type User struct {
	ID            UserID `json:"id"`
	PhoneNumber   string `json:"phone_number"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	EmailVerified bool   `json:"-"`
	PhoneVerified bool   `json:"-"`
}

type Account struct {
	ID                uuid.UUID
	Provider          string
	UserID            UserID
	ProviderAccountID string
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	ByPhone(ctx context.Context, phone string) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
	User(ctx context.Context, id UserID) (*User, error)
	Account(ctx context.Context, id uuid.UUID) (*Account, error)
	LinkAccount(ctx context.Context, account *Account) error
}

func (a *Account) Ok() error {
	if a.Provider == "" {
		return E(EInvalid, "Provider is required")
	} else if a.Provider != authProviderGoogle && a.Provider != authProviderTelegram {
		return E(EInvalid, "Unsupported auth provider")
	} else if a.ProviderAccountID == "" {
		return E(EInvalid, "Provider account id is required")
	} else if a.UserID.IsNil() {
		return E(EInvalid, "User id is required")
	}
	return nil
}

func (u *User) Ok() error {
	if u.PhoneNumber == "" && u.Email == "" {
		return E(EInvalid, "Either phone number or email is required")
	} else if u.FirstName == "" {
		return E(EInvalid, "First name is required")
	} else if u.LastName == "" {
		return E(EInvalid, "Last name is required")
	}
	return nil
}
