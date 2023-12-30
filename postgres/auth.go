package postgres

import (
	"context"
	"errors"
	"math/rand"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type AuthService struct {
	pool          *pgxpool.Pool
	argonHasher   yeahapi.ArgonHasher
	highwayHasher yeahapi.HighwayHasher
}

func NewAuthService(pool *pgxpool.Pool, argonHasher yeahapi.ArgonHasher, highwayHasher yeahapi.HighwayHasher) *AuthService {
	return &AuthService{
		pool:          pool,
		argonHasher:   argonHasher,
		highwayHasher: highwayHasher,
	}
}

func (a *AuthService) CreateOtp(ctx context.Context, otp *yeahapi.Otp) (*yeahapi.Otp, error) {
	const op yeahapi.Op = "postgres/AuthService.CreateOtp"
	code := strconv.Itoa(100000 + rand.Intn(999999-100000))

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	otp.ID = id
	otp.Code = code

	hash, err := a.highwayHasher.Hash([]byte(otp.Identifier + code))

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	otp.Hash = hash

	hashedCode, err := a.argonHasher.Hash([]byte(code))
	if err != nil {
		return nil, yeahapi.E(op, "unable to hash otp code")
	}

	_, err = a.pool.Exec(ctx,
		"insert into otps (id, code, hash, expires_at) values ($1, $2, $3, $4)",
		otp.ID, hashedCode, otp.Hash, otp.ExpiresAt,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return otp, nil
}

func (a *AuthService) VerifyOtp(ctx context.Context, otp *yeahapi.Otp) error {
	const op yeahapi.Op = "postgres/AuthService.VerifyOtp"
	hash, err := a.highwayHasher.Hash([]byte(otp.Identifier + otp.Code))
	if err != nil {
		return yeahapi.E(op, err)
	}

	if hash != otp.Hash {
		return yeahapi.E(op, yeahapi.EInvalid, "hashes don't match")
	}

	savedOtp, err := a.Otp(ctx, otp.Hash, false)

	if err != nil {
		return yeahapi.E(op, err)
	}

	if err := a.argonHasher.Verify(otp.Code, savedOtp.Code); err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}

func (a *AuthService) Otp(ctx context.Context, hash string, confirmed bool) (*yeahapi.Otp, error) {
	const op yeahapi.Op = "postgres/AuthService.Otp"
	var otp yeahapi.Otp

	err := a.pool.QueryRow(ctx,
		"select id, hash, code, expires_at, confirmed from otps where hash = $1 and confirmed = $2 order by id desc limit 1",
		hash, confirmed).Scan(&otp.ID, &otp.Hash, &otp.Code, &otp.ExpiresAt, &otp.Confirmed)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return &otp, nil
}

func (a *AuthService) CreateAuth(ctx context.Context, auth *yeahapi.Auth) (*yeahapi.Auth, error) {
	const op yeahapi.Op = "postgres/AuthService.CreateAuth"

	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err, "unable to generate uuid")
	}

	auth.Session.ID = id.String()

	_, err = a.pool.Exec(ctx,
		"insert into seessions (id, user_id, client_id, user_agent, ip) values ($1, $2, $3, $4, $5)",
		auth.Session.ID, auth.Session.UserID, auth.Session.ClientID, auth.Session.UserAgent, auth.Session.IP,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return auth, nil
}

func (a *AuthService) DeleteAuth(ctx context.Context, sessionID string) error {
	const op yeahapi.Op = "postgres/AuthService.DeleteAuth"
	if _, err := a.pool.Exec(ctx, "delete from sessions where id = $1", sessionID); err != nil {
		return yeahapi.E(op, err)
	}
	return nil
}

func (a *AuthService) Session(ctx context.Context, sessionID string) (*yeahapi.Session, error) {
	const op yeahapi.Op = "postgres/AuthService.Session"
	var session yeahapi.Session

	err := a.pool.QueryRow(ctx,
		"select id, user_id, active, client_id from sessions where id = $1",
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.Active, &session.ClientID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotExist)
		}
		return nil, yeahapi.E(op, err)
	}

	return &session, nil
}
