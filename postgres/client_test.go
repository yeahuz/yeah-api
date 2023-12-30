package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/postgres"
)

func MustCreateClient(tb testing.TB, ctx context.Context, pool *pgxpool.Pool, client *yeahapi.Client) (*yeahapi.Client, context.Context) {
	tb.Helper()
	if _, err := postgres.NewClientService(pool).CreateClient(ctx, client); err != nil {
		tb.Fatal(err)
	}

	return client, yeahapi.NewContextWithClient(ctx, client)
}
