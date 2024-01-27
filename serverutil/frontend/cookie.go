package frontend

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strings"

	yeahapi "github.com/yeahuz/yeah-api"
)

type CookieService interface {
	SetCookie(w http.ResponseWriter, cookie *http.Cookie) error
	ReadCookie(r *http.Request, name string) (string, error)
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

	// Wrap the cipher block in Galois Counter Mode.
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
