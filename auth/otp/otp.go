package otp

import (
	"context"
	"encoding/hex"
	e "errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/minio/highwayhash"
	"github.com/yeahuz/yeah-api/auth/argon"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

type Otp struct {
	id        int
	Code      string
	Hash      string
	Used      bool
	ExpiresAt time.Time
}

func randomIn(min, max int) int {
	return min + rand.Intn(max-min)
}

func New(duration time.Duration) *Otp {
	expiresAt := time.Now().Add(duration)
	code := strconv.Itoa(randomIn(100000, 999999))

	otp := &Otp{
		Code:      code,
		ExpiresAt: expiresAt,
	}

	return otp
}

func genHash(bytes []byte) (string, error) {
	key, err := hex.DecodeString(config.Config.HighwayHashKey)
	if err != nil {
		return "", errors.Internal
	}

	hash, err := highwayhash.New(key)
	if err != nil {
		return "", errors.Internal
	}
	hash.Write(bytes)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (o *Otp) VerifyHash(bytes []byte) error {
	hash, err := genHash(bytes)
	if err != nil {
		return err
	}

	if hash != o.Hash {
		return errors.NewBadRequest(l.T("Hash invalid"))
	}

	return nil
}

func (o *Otp) Verify(code string) error {
	if time.Now().Compare(o.ExpiresAt) > 0 {
		return errors.NewBadRequest(l.T("Code expired"))
	}

	match, err := argon.Verify(code, o.Code)
	if err != nil {
		return errors.Internal
	}

	if !match {
		return errors.NewBadRequest(l.T("Code is invalid"))
	}

	return nil
}

func (o *Otp) Save(identifier string) error {
	code, err := argon.Hash(o.Code)
	if err != nil {
		return errors.Internal
	}

	hash, err := genHash([]byte(identifier + code))

	if err != nil {
		return err
	}

	o.Hash = hash

	err = db.Pool.QueryRow(
		context.Background(),
		"insert into otps (code, hash, expires_at) values ($1, $2, $3) returning id",
		code, o.Hash, o.ExpiresAt,
	).Scan(&o.id)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func GetByHash(hash string) (*Otp, error) {
	var otp Otp
	err := db.Pool.QueryRow(
		context.Background(),
		"select hash, code, expires_at, used from otps where hash = $1 and used = false order by id desc limit 1",
		hash).Scan(&otp.Hash, &otp.Code, &otp.ExpiresAt, &otp.Used)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("Hash not found"))
		}
		return nil, errors.Internal
	}

	return &otp, nil
}
