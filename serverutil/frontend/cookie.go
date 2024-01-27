package frontend

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

const flashKey = "flash"

type CookieService interface {
	SetCookie(w http.ResponseWriter, cookie *http.Cookie) error
	ReadCookie(r *http.Request, name string) (string, error)
	SetFlash(w http.ResponseWriter, name string, value []byte)
	GetFlash(w http.ResponseWriter, r *http.Request, name string) ([]byte, error)
}

type Cookies struct {
	secret string
}

func NewCookieService(secret string) *Cookies {
	return &Cookies{
		secret: secret,
	}
}

func (c *Cookies) SetCookie(w http.ResponseWriter, cookie *http.Cookie) error {
	block, err := aes.NewCipher([]byte(c.secret))
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return err
	}

	plaintext := fmt.Sprintf("%s:%s", cookie.Name, cookie.Value)
	encryptedValue := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	cookie.Value = string(encryptedValue)
	http.SetCookie(w, cookie)
	return nil
}

func (c *Cookies) ReadCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(c.secret))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()

	nonce := cookie.Value[:nonceSize]
	ciphertext := cookie.Value[nonceSize:]
	plaintext, err := aesGCM.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return "", yeahapi.E(yeahapi.EInvalid, "invalid cookie value")
	}

	expectedName, value, ok := strings.Cut(string(plaintext), ":")
	if !ok {
		return "", yeahapi.E(yeahapi.EInvalid, "invalid cookie value")
	}

	if expectedName != name {
		return "", yeahapi.E(yeahapi.EInvalid, "invalid cookie value")
	}

	return value, nil
}

func (c *Cookies) SetFlash(w http.ResponseWriter, name string, value []byte) {
	cookie := &http.Cookie{Name: flashKey + name, Value: base64.URLEncoding.EncodeToString(value)}
	http.SetCookie(w, cookie)
}

func (c *Cookies) GetFlash(w http.ResponseWriter, r *http.Request, name string) ([]byte, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	value, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil, err
	}

	http.SetCookie(w, &http.Cookie{Name: flashKey + name, MaxAge: -1, Expires: time.Unix(1, 0)})
	return value, err
}
