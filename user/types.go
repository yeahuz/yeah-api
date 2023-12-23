package user

type UserID string

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
	ID UserID `json:"id"`
	NewUserOpts
}

type Account struct {
	ID                string
	Provider          string
	UserID            UserID
	ProviderAccountID string
}
