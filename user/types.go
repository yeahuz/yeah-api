package user

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
	ID string `json:"id"`
	NewUserOpts
}

type Account struct {
	ID                int
	Provider          string
	UserID            string
	ProviderAccountID string
}
