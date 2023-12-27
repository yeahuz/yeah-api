package inmem

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/minio/highwayhash"
	yeahapi "github.com/yeahuz/yeah-api"
	"golang.org/x/crypto/argon2"
)

type ArgonHasher struct {
	params yeahapi.ArgonParams
}

type HighwayHasher struct {
	key string
}

func NewArgonHasher(params yeahapi.ArgonParams) *ArgonHasher {
	return &ArgonHasher{
		params: params,
	}
}

func NewHighwayHasher(key string) *HighwayHasher {
	return &HighwayHasher{
		key: key,
	}
}

func (h *HighwayHasher) Hash(b []byte) (string, error) {
	const op yeahapi.Op = "highwayHasher.Hash"
	key, err := hex.DecodeString(h.key)
	if err != nil {
		return "", yeahapi.E(op, err)
	}

	hash, err := highwayhash.New(key)

	if err != nil {
		return "", yeahapi.E(op, err)
	}

	hash.Write(b)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (h *ArgonHasher) Hash(b []byte) (string, error) {
	salt, err := genRandBytes(h.params.SaltLen)
	if err != nil {
		return "", err
	}

	idKey := argon2.IDKey(b, salt, h.params.Time, h.params.Memory, h.params.Threads, h.params.KeyLen)
	b64salt := base64.RawStdEncoding.EncodeToString(salt)
	b64hash := base64.RawStdEncoding.EncodeToString(idKey)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, h.params.Memory, h.params.Time, h.params.Threads, b64salt, b64hash)

	return encodedHash, nil
}

func (h *ArgonHasher) Verify(s, encoded string) error {
	const op yeahapi.Op = "argonHasher.Verify"
	p, salt, hash, err := h.Decode(encoded)
	if err != nil {
		return yeahapi.E(op, err)
	}

	otherHash := argon2.IDKey([]byte(s), salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	if subtle.ConstantTimeCompare(hash, otherHash) == 0 {
		return yeahapi.E(op, "invalid hash")
	}
	return nil
}

func (h *ArgonHasher) Decode(encoded string) (p *yeahapi.ArgonParams, salt, hash []byte, err error) {
	const op yeahapi.Op = "argonHasher.Decode"
	vals := strings.Split(encoded, "$")

	if len(vals) != 6 {
		return nil, nil, nil, yeahapi.E(op, yeahapi.EInvalid, "invalid hash length")
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)

	if err != nil {
		return nil, nil, nil, yeahapi.E(op, err)
	}

	if version != argon2.Version {
		return nil, nil, nil, yeahapi.E(op, yeahapi.EInvalid, "invalid version")
	}

	p = &yeahapi.ArgonParams{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Threads)

	if err != nil {
		return nil, nil, nil, yeahapi.E(op, err)
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}

	p.SaltLen = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])

	if err != nil {
		return nil, nil, nil, yeahapi.E(op, err)
	}

	p.KeyLen = uint32(len(hash))

	return p, salt, hash, nil
}

func genRandBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
