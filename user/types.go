package user

import "time"

type User struct {
	id            int       `json:"id"`
	phone         string    `json:"phone"`
	phoneVerified bool      `json:"phone_verified"`
	name          string    `json:"name"`
	username      string    `json:"username"`
	bio           string    `json:"bio"`
	websiteUrl    string    `json:"website_url"`
	photoUrl      string    `json:"photo_url"`
	email         string    `json:"email"`
	emailVerified bool      `json:"email_verified"`
	password      string    `json:"password"`
	profileUrl    string    `json:"profile_url"`
	verified      bool      `json:"verified"`
	createdAt     time.Time `json:"created_at"`
	updatedAt     time.Time `json:"updated_at"`
}

type Account struct {
	id                int
	provider          string
	userID            int
	providerAccountId string
	createdAt         time.Time
	updatedAt         time.Time
}
