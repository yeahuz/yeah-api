package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	e "errors"

	"github.com/jackc/pgx/v5"
	"github.com/yeahuz/yeah-api/auth/argon"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

type clientType string

var (
	clientInternal     clientType = "internal"
	clientConfidential clientType = "confidential"
	clientPublic       clientType = "public"
)

type Client struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	secret string
	Type   clientType `json:"type"`
	Active bool       `json:"active"`
}

func New(name string, clientType clientType) *Client {
	return &Client{
		Name: name,
		Type: clientType,
	}
}

func (c *Client) Verify(secret string) error {
	if c.Type == clientPublic {
		return nil
	}

	return argon.Verify(secret, c.secret)
}

func (c *Client) Save(ctx context.Context) error {
	secret, err := generateSecret()
	if err != nil {
		return err
	}

	hash, err := argon.Hash(secret)
	if err != nil {
		return err
	}

	return db.Pool.QueryRow(ctx,
		"insert into clients (name, type, secret) values ($1, $2, $3) returning id",
		c.Name, c.Type, hash,
	).Scan(&c.ID)
}

func GetById(ctx context.Context, id string) (*Client, error) {
	var client Client
	err := db.Pool.QueryRow(ctx,
		"select id, active, name, secret, type from clients where id = $1",
		id).Scan(&client.ID, &client.Active, &client.Name, &client.secret, &client.Type)

	if err != nil {
		if e.Is(err, pgx.ErrNoRows) {
			return nil, errors.NewNotFound(l.T("Client with id %s not found", id))
		}
		return nil, errors.Internal
	}

	return &client, nil
}

func Middleware(next http.Handler) http.Handler {
	return c.HandleError(func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		clientId := r.Header.Get("X-Client-Id")
		clientSecret := r.Header.Get("X-Client-Secret")
		client, err := GetById(ctx, clientId)
		if err != nil {
			return err
		}

		if err := client.Verify(clientSecret); err != nil {
			return err
		}

		ctx = context.WithValue(r.Context(), "client", client)
		next.ServeHTTP(w, r.WithContext(ctx))
		return nil
	})
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
