package credential

import (
	"context"

	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
)

type Credential struct {
	ID           int
	CredentialID string
	Title        string
	PubKey       string
	Counter      int
	UserID       int
	Transports   []string
}

func (c *Credential) Save(ctx context.Context) error {
	err := db.Pool.QueryRow(ctx,
		"insert into credentials (title, credential_id, pubkey, counter, user_id, transports) values ($1, $2, $3, $4, $5, $6) returning id",
		c.Title, c.CredentialID, c.PubKey, c.Counter, c.UserID, c.Transports,
	).Scan(&c.ID)

	if err != nil {
		// TODO: handle errors properly
		return errors.Internal
	}
	return nil
}

func GetById(ctx context.Context, credId string) (*Credential, error) {
	var credential Credential
	err := db.Pool.QueryRow(ctx,
		"select id, credential_id, title, transports, user_id from credentials where credential_id = $1", credId).Scan(
		&credential.ID, &credential.Title, &credential.Transports, &credential.UserID)

	if err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	return &credential, nil
}

func GetAll(ctx context.Context) ([]Credential, error) {
	credentials := make([]Credential, 0)

	rows, err := db.Pool.Query(ctx, "select id, credential_id, title, transports, user_id from credentials order by id desc")

	defer rows.Close()
	if err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	for rows.Next() {
		var crd Credential
		if err := rows.Scan(&crd.ID, &crd.CredentialID, &crd.Title, &crd.Transports, &crd.UserID); err != nil {
			// TODO: handle errors properly
			return nil, errors.Internal
		}
		credentials = append(credentials, crd)
	}

	if err := rows.Err(); err != nil {
		// TODO: handle errors properly
		return nil, errors.Internal
	}

	return credentials, nil
}
