package postgres

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type ClientService struct {
	pool        *pgxpool.Pool
	ArgonHasher yeahapi.ArgonHasher
}

func NewClientService(pool *pgxpool.Pool, hasher yeahapi.ArgonHasher) *ClientService {
	return &ClientService{
		pool:        pool,
		ArgonHasher: hasher,
	}
}

func (c *ClientService) Client(ctx context.Context, id yeahapi.ClientID) (*yeahapi.Client, error) {
	const op yeahapi.Op = "postgres/ClientService.Client"
	var client yeahapi.Client
	err := c.pool.QueryRow(ctx,
		"select id, name, secret, type from clients where id = $1", id,
	).Scan(&client.ID, &client.Name, &client.Secret, &client.Type)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}

		return nil, yeahapi.E(op, err)
	}

	return &client, nil
}

func (c *ClientService) CreateClient(ctx context.Context, client *yeahapi.Client) (*yeahapi.Client, error) {
	const op yeahapi.Op = "postgres/ClientService.Client"
	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	if client.Type != yeahapi.ClientPublic && client.Secret == "" {
		return nil, yeahapi.E(op, yeahapi.EInvalid, "Client secret is required for non-public clients")
	}

	hash, err := c.ArgonHasher.Hash([]byte(client.Secret))
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	client.ID = yeahapi.ClientID{id}
	_, err = c.pool.Exec(ctx,
		"insert into clients (id, name, secret, type) values ($1, $2, $3, $4)",
		client.ID, client.Name, hash, client.Type,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return client, nil
}

func (c *ClientService) VerifySecret(client *yeahapi.Client, secret string) error {
	const op yeahapi.Op = "postgres/ClientService.VerifySecret"
	if client.Type == yeahapi.ClientPublic {
		return nil
	}

	if err := c.ArgonHasher.Verify(secret, client.Secret); err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}
