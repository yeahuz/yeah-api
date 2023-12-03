package user

import "time"

type User struct {
	ID          int    `json:"id"`
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
}

type NewUserOpts struct {
	PhoneNumber string
	Email       string
	FirstName   string
	LastName    string
}

type Account struct {
	id                int
	provider          string
	userID            int
	providerAccountId string
	createdAt         time.Time
	updatedAt         time.Time
}
