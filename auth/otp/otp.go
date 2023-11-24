package otp

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/yeahuz/yeah-api/auth/argon"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/db"
)

type Otp struct {
	Code      string
	Hash      string
	Used      bool
	ExpiresAt time.Time
}

func randomIn(min, max int) int {
	return min + rand.Intn(max-min)
}

func New(identifier string, duration time.Duration) (*Otp, error) {
	expiresAt := time.Now().Add(time.Minute * duration)
	code, err := argon.Hash(strconv.Itoa(randomIn(100000, 999999)))
	if err != nil {
		return nil, err
	}

	hash, err := argon.Hash(identifier + code)
	if err != nil {
		return nil, err
	}

	otp := &Otp{
		Code:      code,
		Hash:      hash,
		ExpiresAt: expiresAt,
	}

	return otp, nil
}

func (o *Otp) Save() (*Otp, error) {
	err := db.Pool.QueryRow(
		context.Background(),
		"insert into otps (code, hash, expires_at) values ($1, $2, $3)",
		o.Code, o.Hash, o.ExpiresAt,
	).Scan(o)

	if err != nil {
		return nil, c.ErrInternal
		// var pgerr *pgconn.PgError
		// if errors.As(err, &pgerr) {
		// 	return nil, c.ErrInternal
		// }
	}

	return o, nil
}

func GetByHash(hash string) (*Otp, error) {
	var otp Otp
	err := db.Pool.QueryRow(
		context.Background(),
		"select hash, code, expires_at, used from otps where hash = $1",
		hash).Scan(&otp.Hash, &otp.Code, &otp.ExpiresAt, &otp.Used)

	if err != nil {
		return nil, err
	}

	return &otp, nil
}
