package user

import "github.com/gofrs/uuid"

type NewUserOpts struct {
	PhoneNumber   string `json:"phone_number"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	EmailVerified bool   `json:"-"`
	PhoneVerified bool   `json:"-"`
}

type User struct {
	ID uuid.UUID `json:"id"`
	NewUserOpts
}

type Account struct {
	ID                uuid.UUID
	Provider          string
	UserID            uuid.UUID
	ProviderAccountID string
}
