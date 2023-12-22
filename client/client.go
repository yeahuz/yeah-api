package client

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	e "errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yeahuz/yeah-api/auth/argon"
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
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
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

	if err := argon.Verify(secret, c.secret); err != nil {
		return errors.NewBadRequest(l.T("Missing or invalid secret for confidential client"))
	}

	return nil
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

	_, err = db.Pool.Exec(ctx,
		"insert into clients (name, type, secret) values ($1, $2, $3) returning id",
		c.ID, c.Name, c.Type, hash,
	)

	return err
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
		return nil, err
	}

	return &client, nil
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
