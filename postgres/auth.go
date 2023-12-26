package postgres

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type AuthService struct {
	pool    *pgxpool.Pool
	locsrv  *yeahapi.LocalizerService
	cqrssrv *yeahapi.CQRSService
}

func NewAuthService(pool *pgxpool.Pool) *AuthService {
	return &AuthService{
		pool: pool,
	}
}

func (a *AuthService) CreateOtp(ctx context.Context, duration time.Duration, identifier string) (*yeahapi.Otp, error) {
	expiresAt := time.Now().Add(duration)
	code := strconv.Itoa(100000 + rand.Intn(999999-100000))

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	otp := &yeahapi.Otp{
		ID:        id,
		Code:      code,
		ExpiresAt: expiresAt,
	}

	hashedCode := "hashed-code"
	// hashedCode, err := argon.Hash(code)

	// if err != nil {
	// 	return nil, err
	// }

	hash, err := genHash([]byte(identifier + hashedCode))

	if err != nil {
		return nil, err
	}

	otp.Hash = hash
	_, err = a.pool.Exec(ctx,
		"insert into otps (id, code, hash, expires_at) values ($1, $2, $3, $4) returning id",
		otp.ID, hashedCode, otp.Hash, otp.ExpiresAt,
	)

	return otp, err
}

func (a *AuthService) VerifyOtp(ctx context.Context) error {
	return nil
}

func (a *AuthService) CreateAuth(ctx context.Context) {}
func (a *AuthService) DeleteAuth(ctx context.Context) {}

func genHash(bytes []byte) (string, error) {
	return "", nil
	// key, err := hex.DecodeString(config.Config.HighwayHashKey)
	// if err != nil {
	// 	return "", errors.Internal
	// }

	// hash, err := highwayhash.New(key)
	// if err != nil {
	// 	return "", errors.Internal
	// }
	// hash.Write(bytes)
	// return hex.EncodeToString(hash.Sum(nil)), nil
}
