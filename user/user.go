package user

import (
	"context"
	e "errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func GetByPhone(phone string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(
		context.Background(),
		"select id from users where phone = $1",
		phone,
	).Scan(&user.id)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.ErrNotFound{Message: l.T("User with phone %s not found", phone), StatusCode: http.StatusNotFound}
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
	).Scan(&user.id)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, c.ErrNotFound{Message: l.T("User with email %s not found", email), StatusCode: http.StatusNotFound}
		}
		return nil, errors.Internal
	}

	return &user, nil
}
