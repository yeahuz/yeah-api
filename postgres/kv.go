package postgres

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type KVService struct {
	pool *pgxpool.Pool
}

func NewKVService(pool *pgxpool.Pool) *KVService {
	return &KVService{
		pool: pool,
	}
}

func (kv *KVService) Set(ctx context.Context, item *yeahapi.KVItem) (*yeahapi.KVItem, error) {
	const op yeahapi.Op = "postgres/KVService.Set"
	if item.Key == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return nil, yeahapi.E(op, err)
		}
		item.Key = id.String()
	}
	_, err := kv.pool.Exec(ctx,
		"insert into kv_store (key, client_id, value) values ($1, $2, $3) on conflict (client_id, key) do update set value = $3",
		item.Key, item.ClientID, item.Value,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return item, nil
}

func (kv *KVService) Get(ctx context.Context, clientID yeahapi.ClientID, key string) (*yeahapi.KVItem, error) {
	const op yeahapi.Op = "postgres/KVService.Get"
	var item yeahapi.KVItem
	err := kv.pool.QueryRow(ctx,
		"select key, value, client_id from kv_store where key = $1 and client_id = $2",
		key, clientID).Scan(&item.Key, &item.Value, &item.ClientID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}
		return nil, yeahapi.E(op, err)
	}

	return &item, nil
}

func (kv *KVService) Remove(ctx context.Context, clientID yeahapi.ClientID, key string) error {
	const op yeahapi.Op = "postgres/KVService.Remove"

	if _, err := kv.pool.Exec(ctx, "delete from kv_store where key = $1 and client_id = $2", key, clientID); err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}
