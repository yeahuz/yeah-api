package user

import "time"

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
	ID int `json:"id"`
	NewUserOpts
}

type Account struct {
	id                int
	provider          string
	userID            int
	providerAccountId string
	createdAt         time.Time
	updatedAt         time.Time
}
