package yeahapi

import (
	"context"

	"github.com/gofrs/uuid"
)

type UserID struct {
	uuid.UUID
}

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
	ID                string
	Provider          string
	UserID            UserID
	ProviderAccountID string
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	ByPhone(ctx context.Context, phone string) (*User, error)
	ByEmail(ctx context.Context, email string) (*User, error)
	User(ctx context.Context, id UserID) (*User, error)
	Account(ctx context.Context, id string) (*Account, error)
	LinkAccount(ctx context.Context, account *Account) error
}
