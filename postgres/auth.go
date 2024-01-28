package postgres

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type AuthService struct {
	pool          *pgxpool.Pool
	argonHasher   yeahapi.ArgonHasher
	highwayHasher yeahapi.HighwayHasher
	signingKey    []byte
}

func NewAuthService(pool *pgxpool.Pool, argonHasher yeahapi.ArgonHasher, highwayHasher yeahapi.HighwayHasher, signingKey string) *AuthService {
	return &AuthService{
		pool:          pool,
		argonHasher:   argonHasher,
		highwayHasher: highwayHasher,
		signingKey:    []byte(signingKey),
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
		return yeahapi.E(op, yeahapi.EOtpHashNotMatched)
	}

	savedOtp, err := a.Otp(ctx, otp.Hash, false)

	if err != nil {
		return yeahapi.E(op, err)
	}

	if time.Now().After(savedOtp.ExpiresAt) {
		return yeahapi.E(op, yeahapi.EOtpCodeExpired)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}

		return nil, yeahapi.E(op, err)
	}

	return &otp, nil
}

func (a *AuthService) CreateAuth(ctx context.Context, auth *yeahapi.Auth) (*yeahapi.Auth, error) {
	const op yeahapi.Op = "postgres/AuthService.CreateAuth"

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	if auth.Session.UserID.IsNil() {
		if err := createUser(ctx, tx, auth.User); err != nil {
			return nil, yeahapi.E(op, err)
		}

		auth.Session.UserID = auth.User.ID
	}

	if err := createSession(ctx, tx, auth); err != nil {
		return nil, yeahapi.E(op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return auth, nil
}

func (a *AuthService) DeleteAuth(ctx context.Context, sessionID uuid.UUID) error {
	const op yeahapi.Op = "postgres/AuthService.DeleteAuth"
	if _, err := a.pool.Exec(ctx, "delete from sessions where id = $1", sessionID); err != nil {
		return yeahapi.E(op, err)
	}
	return nil
}

func (a *AuthService) Session(ctx context.Context, sessionID uuid.UUID) (*yeahapi.Session, error) {
	const op yeahapi.Op = "postgres/AuthService.Session"
	var session yeahapi.Session

	err := a.pool.QueryRow(ctx,
		"select id, user_id, active, client_id from sessions where id = $1",
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.Active, &session.ClientID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}
		return nil, yeahapi.E(op, err)
	}

	return &session, nil
}

func (a *AuthService) CreateLoginToken(expiresAt time.Time) (*yeahapi.LoginToken, error) {
	const op yeahapi.Op = "postgres/AuthService.CreateLoginToken"

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return nil, yeahapi.E(yeahapi.EInternal, "Unable to generate bytes")
	}

	payload := make([]byte, 24)
	copy(payload, b)
	binary.BigEndian.PutUint64(payload[16:], uint64(expiresAt.Unix()))
	h := hmac.New(sha256.New, a.signingKey)
	h.Write(payload)
	sig := h.Sum(nil)

	loginToken := &yeahapi.LoginToken{
		Token:     fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(payload), base64.RawURLEncoding.EncodeToString(sig)),
		ExpiresAt: expiresAt,
	}

	return loginToken, nil
}

func (a *AuthService) VerifyLoginToken(token string) error {
	const op yeahapi.Op = "postgres/AuthService.VerifyLoginToken"
	loginToken, err := decodeLoginToken(token)
	if err != nil {
		return err
	}

	h := hmac.New(sha256.New, a.signingKey)
	h.Write(loginToken.Payload)
	checksum := h.Sum(nil)

	if time.Now().After(loginToken.ExpiresAt) {
		return yeahapi.E(yeahapi.EInvalid, "Login token expired")
	}

	if bytes.Equal(checksum, loginToken.Sig) {
		return yeahapi.E(yeahapi.EInvalid, "Login token invalid")
	}

	return nil
}

func decodeLoginToken(token string) (*yeahapi.LoginToken, error) {
	parts := strings.Split(token, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var expiresAt int64
	binary.Read(bytes.NewReader(payload[16:]), binary.BigEndian, &expiresAt)

	loginToken := &yeahapi.LoginToken{
		Payload:   payload,
		Sig:       sig,
		ExpiresAt: time.Unix(expiresAt, 0),
	}

	return loginToken, nil
}

func createSession(ctx context.Context, tx pgx.Tx, auth *yeahapi.Auth) error {
	const op yeahapi.Op = "postgres/AuthService.createSession"

	if err := auth.Ok(); err != nil {
		return yeahapi.E(op, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return yeahapi.E(op, err, "unable to generate uuid")
	}

	auth.Session.ID = id

	_, err = tx.Exec(ctx,
		"insert into sessions (id, user_id, client_id, user_agent, ip) values ($1, $2, $3, $4, $5)",
		auth.Session.ID, auth.Session.UserID, auth.Session.ClientID, auth.Session.UserAgent, auth.Session.IP,
	)

	if err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}
