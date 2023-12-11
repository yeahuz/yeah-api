package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var (
	emailRegex = regexp.MustCompile(`(?i)^(([^<>()[\].,;:\s@"]+(\.[^<>()[\].,;:\s@"]+)*)|(".+"))@(([^<>()[\].,;:\s@"]+\.)+[^<>()[\].,;:\s@"]{2,})$`)
	l          = localizer.GetDefault()
)

func newLoginToken() (*LoginToken, error) {
	b := make([]byte, 16)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Second * 30)
	payload := make([]byte, 24)
	copy(payload, b)
	binary.BigEndian.PutUint64(payload[16:], uint64(expiresAt.Unix()))
	h := hmac.New(sha256.New, []byte(config.Config.JwtSecret))
	h.Write(payload)
	sig := h.Sum(nil)

	loginToken := &LoginToken{
		Token:     fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(payload), base64.RawURLEncoding.EncodeToString(sig)),
		ExpiresAt: expiresAt,
	}

	return loginToken, nil
}

func parseLoginToken(tok string) (*LoginToken, error) {
	parts := strings.Split(tok, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	token := &LoginToken{
		payload: payload,
		sig:     sig,
	}

	return token, nil
}

func (t *LoginToken) verify() bool {
	h := hmac.New(sha256.New, []byte(config.Config.JwtSecret))
	h.Write(t.payload)
	checksum := h.Sum(nil)

	var expiresAt int64
	binary.Read(bytes.NewReader(t.payload[16:]), binary.BigEndian, &expiresAt)
	if time.Now().After(time.Unix(expiresAt, 0)) {
		return false
	}

	return bytes.Equal(checksum, t.sig)
}

func (pcd PhoneCodeData) validate() error {
	if len(pcd.PhoneNumber) == 0 {
		return errors.NewBadRequest(l.T("Phone number is required"))
	}

	if len(pcd.PhoneNumber) != 13 {
		return errors.NewBadRequest(l.T("Phone number is invalid"))
	}

	return nil
}

func (ecd EmailCodeData) validate() error {
	if len(ecd.Email) == 0 {
		return errors.NewBadRequest(l.T("Email is required"))
	}

	if !emailRegex.MatchString(ecd.Email) {
		return errors.NewBadRequest(l.T("Email is invalid"))
	}

	return nil
}

func (sipd SignInPhoneData) validate() error {
	errs := make(map[string]string)

	if len(sipd.PhoneNumber) == 0 {
		errs["phone_number"] = l.T("Phone number is required")
	}

	if len(sipd.PhoneNumber) != 13 {
		errs["phone_number"] = l.T("Phone number is invalid")
	}

	if len(sipd.Code) == 0 {
		errs["code"] = l.T("Phone code is required")
	}

	if len(sipd.Hash) == 0 {
		errs["hash"] = l.T("Hash is required")
	}

	if len(errs) > 0 {
		return errors.NewValidation(errs)
	}

	return nil
}

func (sied SignInEmailData) validate() error {
	errs := make(map[string]string)

	if !emailRegex.MatchString(sied.Email) {
		errs["email"] = l.T("Email is invalid")
	}

	if len(sied.Email) == 0 {
		errs["email"] = l.T("Email is required")
	}

	if len(sied.Code) == 0 {
		errs["code"] = l.T("Email code is required")
	}

	if len(sied.Hash) == 0 {
		errs["hash"] = l.T("Hash is required")
	}

	if len(errs) > 0 {
		return errors.NewValidation(errs)
	}

	return nil
}

func (sued SignUpEmailData) validate() error {
	errs := make(map[string]string)

	if !emailRegex.MatchString(sued.Email) {
		errs["email"] = l.T("Email is invalid")
	}

	if len(sued.Email) == 0 {
		errs["email"] = l.T("Email is required")
	}

	if len(sued.Code) == 0 {
		errs["code"] = l.T("Email code is required")
	}

	if len(sued.Hash) == 0 {
		errs["hash"] = l.T("Hash is required")
	}

	if len(sued.FirstName) == 0 {
		errs["first_name"] = l.T("First name is required")
	}

	if len(sued.LastName) == 0 {
		errs["last_name"] = l.T("Last name is required")
	}

	if len(errs) > 0 {
		return errors.NewValidation(errs)
	}

	return nil
}

func (supd SignUpPhoneData) validate() error {
	errs := make(map[string]string)

	if len(supd.PhoneNumber) == 0 {
		errs["phone_number"] = l.T("Phone number is required")
	}

	if len(supd.PhoneNumber) != 13 {
		errs["phone_number"] = l.T("Phone number is invalid")
	}

	if len(supd.Code) == 0 {
		errs["code"] = l.T("Phone code is required")
	}

	if len(supd.Hash) == 0 {
		errs["hash"] = l.T("Hash is required")
	}

	if len(supd.FirstName) == 0 {
		errs["first_name"] = l.T("First name is required")
	}

	if len(supd.LastName) == 0 {
		errs["last_name"] = l.T("Last name is required")
	}

	if len(errs) > 0 {
		return errors.NewValidation(errs)
	}

	return nil
}
