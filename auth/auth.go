package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	e "errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/config"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var (
	emailRegex = regexp.MustCompile(`(?i)^(([^<>()[\].,;:\s@"]+(\.[^<>()[\].,;:\s@"]+)*)|(".+"))@(([^<>()[\].,;:\s@"]+\.)+[^<>()[\].,;:\s@"]{2,})$`)
	l          = localizer.GetDefault()
)

func newOAuthFlow(data oAuthFlowData) oAuthFlow {
	flow := oAuthFlow{}
	switch data.Provider {
	case providerGoogle:
		{
			//TODO: pass redirect url
			flow.URL = config.Config.GoogleOAuthConf.AuthCodeURL(data.State)
			break
		}
	}

	return flow
}

func newSession(userID, clientID, userAgent string) *session {
	return &session{
		UserID:    userID,
		ClientID:  clientID,
		UserAgent: userAgent,
	}
}

func (s *session) save(ctx context.Context) error {
	return db.Pool.QueryRow(ctx,
		"insert into sessions (user_id, client_id, user_agent, ip) values ($1, $2, $3, $4) returning id",
		s.UserID, s.ClientID, s.UserAgent, s.IP,
	).Scan(&s.ID)
}

func (s *session) remove(ctx context.Context) error {
	if _, err := db.Pool.Exec(ctx, "delete from sessions where id = $1", s.ID); err != nil {
		return err
	}
	return nil
}

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

func (d oAuthFlowData) validate() error {
	errs := make(map[string]string)

	if len(d.Provider) == 0 {
		errs["provider"] = l.T("OAuth provider is required")
	}

	if d.Provider != providerGoogle {
		errs["provider"] = l.T("Unsupported OAuth provider")
	}

	if len(d.State) == 0 {
		errs["state"] = l.T("OAuth state is required")
	}

	if len(errs) > 0 {
		return errors.NewValidation(errs)
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

func (d signInGoogleData) validate() error {
	if len(d.Code) == 0 {
		return errors.NewBadRequest(l.T("Response code is required"))
	}

	return nil
}

func (sipd signInPhoneData) validate() error {
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

func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}

func getSessionById(ctx context.Context, id string) (*session, error) {
	var session session
	err := db.Pool.QueryRow(ctx,
		"select id, user_id, active from sessions where id = $1",
		id).Scan(&session.ID, &session.UserID, &session.Active)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("Session with id %s not found", id))
		}
		return nil, errors.Internal
	}

	return &session, nil
}

func getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) == 0 {
		ip = r.Header.Get("X-Real-Ip")
	}
	if len(ip) == 0 {
		ip = r.RemoteAddr
	}

	return ip
}

func Middleware(next http.Handler) http.Handler {
	return c.HandleError(func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		sessionId := r.Header.Get("X-Session-Id")
		if sessionId == "" {
			return errors.Unauthorized
		}

		if !isValidUUID(sessionId) {
			return errors.NewBadRequest(l.T("Missing valid session id"))
		}

		session, err := getSessionById(ctx, sessionId)
		if err != nil {
			return err
		}

		if !session.Active {
			return errors.Unauthorized
		}

		ctx = context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}
