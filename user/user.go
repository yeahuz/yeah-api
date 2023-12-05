package user

import (
	"context"
	e "errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func New(opts NewUserOpts) *User {
	return &User{
		NewUserOpts: opts,
	}
}

func (u *User) Save() error {
	err := db.Pool.QueryRow(context.Background(), "insert into users (first_name, last_name, email, phone, email_verified, phone_verified) values ($1, $2, $3, $4, $5, $6) returning id",
		u.FirstName, u.LastName, u.Email, u.PhoneNumber, u.EmailVerified, u.PhoneVerified,
	).Scan(&u.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if e.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.NewBadRequest(l.T("User already exists"))
		}
		return errors.Internal
	}

	return nil
}

func GetByPhone(phone string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(context.Background(),
		`select id, first_name, last_name, phone, email, coalesce(username, '') from users where phone = $1`,
		phone).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

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
		`select id, first_name, last_name, phone, email, coalesce(username, '') from users where email = $1`,
		email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("User with email %s not found", email))
		}
		return nil, errors.Internal
	}

	return &user, nil
}
