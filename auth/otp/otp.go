package otp

import (
	"context"
	"encoding/hex"
	e "errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
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
	id        uuid.UUID
	Code      string
	Hash      string
	Confirmed bool
	ExpiresAt time.Time
}

func randomIn(min, max int) int {
	return min + rand.Intn(max-min)
}

func New(duration time.Duration) (*Otp, error) {
	expiresAt := time.Now().Add(duration)
	code := strconv.Itoa(randomIn(100000, 999999))

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	otp := &Otp{
		Code:      code,
		ExpiresAt: expiresAt,
		id:        id,
	}

	return otp, nil
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

	return argon.Verify(code, o.Code)
}

func (o *Otp) Confirm() error {
	err := db.Pool.QueryRow(context.Background(), "update otps set confirmed = true where id = $1 returning confirmed", o.id).Scan(&o.Confirmed)
	if err != nil {
		return errors.Internal
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

	_, err = db.Pool.Exec(
		context.Background(),
		"insert into otps (id, code, hash, expires_at) values ($1, $2, $3, $4) returning id",
		o.id, code, o.Hash, o.ExpiresAt,
	)

	if err != nil {
		return errors.Internal
	}

	return nil
}

func GetByHash(hash string, confirmed bool) (*Otp, error) {
	var otp Otp
	err := db.Pool.QueryRow(
		context.Background(),
		"select id, hash, code, expires_at, confirmed from otps where hash = $1 and confirmed = $2 order by id desc limit 1",
		hash, confirmed).Scan(&otp.id, &otp.Hash, &otp.Code, &otp.ExpiresAt, &otp.Confirmed)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("Hash not found"))
		}
		return nil, errors.Internal
	}

	return &otp, nil
}
