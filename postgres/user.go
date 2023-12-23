package postgres

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type UserService struct {
	pool  *pgxpool.Pool
	locsv *yeahapi.LocalizerService
}

func NewUserService(pool *pgxpool.Pool) *UserService {
	return &UserService{
		pool: pool,
	}
}

func (s *UserService) User(ctx context.Context, id yeahapi.UserID) (*yeahapi.User, error) {
	var user yeahapi.User
	err := s.pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where id = $1`,
		id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	return &user, err
}

func (s *UserService) ByEmail(ctx context.Context, email string) (*yeahapi.User, error) {
	var user yeahapi.User
	err := s.pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where email = $1`,
		email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	return &user, err
}

func (s *UserService) ByPhone(ctx context.Context, phone string) (*yeahapi.User, error) {
	var user yeahapi.User
	err := s.pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where phone = $1`,
		phone).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	return &user, err
}

func (s *UserService) Account(ctx context.Context, id string) (*yeahapi.Account, error) {
	var account yeahapi.Account

	err := s.pool.QueryRow(
		ctx,
		"select id, provider, user_id, provider_account_id from accounts where id = $1",
		id,
	).Scan(&account.ID, &account.Provider, &account.UserID, &account.ProviderAccountID)

	return &account, err
}

func (s *UserService) CreateUser(ctx context.Context, user *yeahapi.User) (*yeahapi.User, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	user.ID = yeahapi.UserID(id.String())

	_, err = s.pool.Exec(ctx,
		"insert into users (id, first_name, last_name, email, phone, email_verified, phone_verified) values ($1, $2, $3, $4, $5, $6, $7)",
		id, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.EmailVerified, user.PhoneVerified,
	)

	return user, err
}

func (s *UserService) LinkAccount(ctx context.Context, account *yeahapi.Account) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	account.ID = id.String()

	_, err = s.pool.Exec(ctx,
		"insert into accounts (id, user_id, provider, provider_account_id) values ($1, $2, $3, $4) returning id",
		account.ID, account.UserID, account.Provider, account.ProviderAccountID,
	)

	return err
}
