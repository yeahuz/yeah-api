package postgres

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	const op yeahapi.Op = "postgres/UserService.User"
	var user yeahapi.User
	err := s.pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where id = $1`,
		id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}

		return nil, yeahapi.E(op, err)
	}

	return &user, nil
}

func (s *UserService) ByEmail(ctx context.Context, email string) (*yeahapi.User, error) {
	const op yeahapi.Op = "postgres/UserService.ByEmail"
	var user yeahapi.User
	err := s.pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where email = $1`,
		email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}

		return nil, yeahapi.E(op, err)
	}

	return &user, nil
}

func (s *UserService) ByPhone(ctx context.Context, phone string) (*yeahapi.User, error) {
	const op yeahapi.Op = "postgres/UserService.ByPhone"
	var user yeahapi.User
	err := s.pool.QueryRow(
		ctx,
		`select id, first_name, last_name, coalesce(phone, ''), coalesce(email, ''), coalesce(username, '') from users where phone = $1`,
		phone).Scan(&user.ID, &user.FirstName, &user.LastName, &user.PhoneNumber, &user.Email, &user.Username)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}

		return nil, yeahapi.E(op, err)
	}

	return &user, nil
}

func (s *UserService) Account(ctx context.Context, id uuid.UUID) (*yeahapi.Account, error) {
	const op yeahapi.Op = "postgres/UserService.Account"
	var account yeahapi.Account
	err := s.pool.QueryRow(
		ctx,
		"select id, provider, user_id, provider_account_id from accounts where id = $1",
		id,
	).Scan(&account.ID, &account.Provider, &account.UserID, &account.ProviderAccountID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}

		return nil, yeahapi.E(op, err)
	}

	return &account, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *yeahapi.User) (*yeahapi.User, error) {
	const op yeahapi.Op = "postgres/UserService.CreateUser"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	defer tx.Rollback(ctx)

	if err := createUser(ctx, tx, user); err != nil {
		return nil, yeahapi.E(op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return user, nil
}

func (s *UserService) LinkAccount(ctx context.Context, account *yeahapi.Account) error {
	const op yeahapi.Op = "postgres/UserService.LinkAccount"
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return yeahapi.E(op, err)
	}

	defer tx.Rollback(ctx)

	if err := linkAccount(ctx, tx, account); err != nil {
		return yeahapi.E(op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}

func linkAccount(ctx context.Context, tx pgx.Tx, account *yeahapi.Account) error {
	const op yeahapi.Op = "postgres/UserService.linkAccount"
	if err := account.Ok(); err != nil {
		return yeahapi.E(op, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	account.ID = id
	_, err = tx.Exec(ctx,
		"insert into accounts (id, user_id, provider, provider_account_id) values ($1, $2, $3, $4) returning id",
		account.ID, account.UserID, account.Provider, account.ProviderAccountID,
	)

	if err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}

func createUser(ctx context.Context, tx pgx.Tx, user *yeahapi.User) error {
	const op yeahapi.Op = "postgres/UserService.createUser"
	if err := user.Ok(); err != nil {
		return yeahapi.E(op, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return yeahapi.E(op, err)
	}

	user.ID = yeahapi.UserID{id}

	_, err = tx.Exec(ctx,
		"insert into users (id, first_name, last_name, email, phone, email_verified, phone_verified) values ($1, $2, $3, $4, $5, $6, $7)",
		user.ID, user.FirstName, user.LastName, user.Email, user.PhoneNumber, user.EmailVerified, user.PhoneVerified,
	)

	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) && pgerrcode.IsIntegrityConstraintViolation(pgerr.Code) {
			return yeahapi.E(op, yeahapi.EFound)
		}

		return yeahapi.E(op, err)
	}

	return nil
}
