package user

import (
	"context"
	e "errors"

	"github.com/jackc/pgx/v5"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func New(opts NewUserOpts) *User {
	return &User{
		FirstName:   opts.FirstName,
		LastName:    opts.LastName,
		Email:       opts.Email,
		PhoneNumber: opts.PhoneNumber,
	}
}

func (u *User) Save() error {
	err := db.Pool.QueryRow(context.Background(), "insert into users (first_name, last_name, email, phone) values ($1, $2, nullif($3, ''), nullif($4, '')) returning id",
		u.FirstName, u.LastName, u.Email, u.PhoneNumber,
	).Scan(&u.ID)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func GetByPhone(phone string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(
		context.Background(),
		"select id from users where phone = $1",
		phone,
	).Scan(&user.ID)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("User with phone %s not found", phone))
		}
		return nil, errors.Internal
	}

	return &user, nil
}

func GetByEmail(email string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(
		context.Background(),
		"select id from users where email = $1",
		email,
	).Scan(&user.ID)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("User with email %s not found", email))
		}
		return nil, errors.Internal
	}

	return &user, nil
}
