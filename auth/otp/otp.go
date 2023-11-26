package otp

import (
	"context"
	"encoding/hex"
	"math/rand"
	"strconv"
	"time"

	"github.com/minio/highwayhash"
	"github.com/yeahuz/yeah-api/auth/argon"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
)

type Otp struct {
	id        int
	Code      string
	CodeLen   int
	Hash      string
	Used      bool
	ExpiresAt time.Time
}

func randomIn(min, max int) int {
	return min + rand.Intn(max-min)
}

func New(duration time.Duration) *Otp {
	expiresAt := time.Now().Add(time.Minute * duration)
	code := strconv.Itoa(randomIn(100000, 999999))

	otp := &Otp{
		Code:      code,
		ExpiresAt: expiresAt,
		CodeLen:   len(code),
	}

	return otp
}

func (o *Otp) Save(identifier string) error {
	key, err := hex.DecodeString(config.Config.HighwayHashKey)

	if err != nil {
		return c.ErrInternal
	}

	hash, err := highwayhash.New(key)
	if err != nil {
		return c.ErrInternal
	}

	hash.Write([]byte(identifier + o.Code))
	o.Hash = hex.EncodeToString(hash.Sum(nil))
	o.Code, err = argon.Hash(o.Code)

	if err != nil {
		return c.ErrInternal
	}

	err = db.Pool.QueryRow(
		context.Background(),
		"insert into otps (code, hash, expires_at) values ($1, $2, $3) returning id",
		o.Code, o.Hash, o.ExpiresAt,
	).Scan(&o.id)

	if err != nil {
		return c.ErrInternal
	}

	return nil
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
