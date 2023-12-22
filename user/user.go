package user

import (
	"context"
	e "errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func New(opts NewUserOpts) (*User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:          id,
		NewUserOpts: opts,
	}, nil
}

func (u *User) Save(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		"insert into users (id, first_name, last_name, email, phone, email_verified, phone_verified) values ($1, $2, $3, $4, $5, $6, $7)",
		u.ID, u.FirstName, u.LastName, u.Email, u.PhoneNumber, u.EmailVerified, u.PhoneVerified,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if e.As(err, &pgErr) && pgErr.Code == "23505" {
			return errors.NewBadRequest(l.T("User already exists"))
		}
		return errors.Internal
	}

	return nil
}

func GetByPhone(ctx context.Context, phone string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where phone = $1`,
		phone).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("User with phone %s not found", phone))
		}
		return nil, errors.Internal
	}

	return &user, nil
}

func GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where email = $1`,
		email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("User with email %s not found", email))
		}
		return nil, errors.Internal
	}

	return &user, nil
}

func GetById(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := db.Pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where id = $1`,
		id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("User with email %s not found", id))
		}
		return nil, errors.Internal
	}

	return &user, nil
}

func GetByAccountId(ctx context.Context, id string) (*Account, error) {
	var account Account

	err := db.Pool.QueryRow(
		ctx,
		"select id, provider, user_id, provider_account_id from accounts where id = $1",
		id,
	).Scan(&account.ID, &account.Provider, &account.UserID, &account.ProviderAccountID)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("Account with id %s not found", id))
		}

		return nil, errors.Internal
	}

	return &account, nil
}

func newAccount(provider, providerID string, userID uuid.UUID) (*Account, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	account := Account{
		Provider:          provider,
		ProviderAccountID: providerID,
		UserID:            userID,
		ID:                id,
	}

	return &account, nil
}

func (a *Account) save(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		"insert into accounts (id, user_id, provider, provider_account_id) values ($1, $2, $3, $4) returning id",
		a.ID, a.UserID, a.Provider, a.ProviderAccountID,
	)

	return err
}

func (u *User) LinkAccount(ctx context.Context, provider string, providerID string) (*Account, error) {
	account, err := newAccount(provider, providerID, u.ID)
	if err != nil {
		return nil, err
	}

	if err := account.save(ctx); err != nil {
		return nil, err
	}

	return account, nil
}
