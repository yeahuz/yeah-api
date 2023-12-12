package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var (
	emailRegex = regexp.MustCompile(`(?i)^(([^<>()[\].,;:\s@"]+(\.[^<>()[\].,;:\s@"]+)*)|(".+"))@(([^<>()[\].,;:\s@"]+\.)+[^<>()[\].,;:\s@"]{2,})$`)
	l          = localizer.GetDefault()
)

func newLoginToken() (*loginToken, error) {
	b := make([]byte, 16)

	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Second * 30)
	payload := make([]byte, 24)
	copy(payload, b)
	binary.BigEndian.PutUint64(payload[16:], uint64(expiresAt.Unix()))
	h := hmac.New(sha256.New, []byte(config.Config.SigningSecret))
	h.Write(payload)
	sig := h.Sum(nil)

	loginToken := &loginToken{
		Token:     fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(payload), base64.RawURLEncoding.EncodeToString(sig)),
		ExpiresAt: expiresAt,
	}

	return loginToken, nil
}

func parseLoginToken(tok string) (*loginToken, error) {
	parts := strings.Split(tok, ".")
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
	token := &loginToken{
		payload:   payload,
		sig:       sig,
		ExpiresAt: time.Unix(expiresAt, 0),
	}

	return token, nil
}

func (t *loginToken) verify() error {
	h := hmac.New(sha256.New, []byte(config.Config.SigningSecret))
	h.Write(t.payload)
	checksum := h.Sum(nil)

	if time.Now().After(t.ExpiresAt) {
		return errors.NewBadRequest(l.T("Login token expired"))
	}

	if !bytes.Equal(checksum, t.sig) {
		return errors.NewBadRequest(l.T("Login token is invalid"))
	}

	return nil
}

func (pcd phoneCodeData) validate() error {
	if len(pcd.PhoneNumber) == 0 {
		return errors.NewBadRequest(l.T("Phone number is required"))
	}

	if len(pcd.PhoneNumber) != 13 {
		return errors.NewBadRequest(l.T("Phone number is invalid"))
	}

	return nil
}

func (ecd emailCodeData) validate() error {
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

func (sied signInEmailData) validate() error {
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

func (sued signUpEmailData) validate() error {
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

func (supd signUpPhoneData) validate() error {
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

func Middleware(next http.Handler) http.Handler {
	return c.HandleError(func(w http.ResponseWriter, r *http.Request) error {
		fmt.Printf("TODO: auth.middleware() not implemented yet!\n")
		next.ServeHTTP(w, r)
		return nil
	})
}
