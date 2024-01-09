package postgres_test

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/inmem"
	"github.com/yeahuz/yeah-api/postgres"
)

func TestClientService_CreateClient(t *testing.T) {
	argonHasher := inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	s := postgres.NewClientService(pool, argonHasher)

	t.Run("OK", func(t *testing.T) {
		clients := [2]yeahapi.Client{
			{
				Name:   "confidential-client",
				Secret: "client",
				Type:   yeahapi.ClientConfidential,
			},
			{
				Name: "public-client",
				Type: yeahapi.ClientPublic,
			},
		}

		ctx := context.Background()

		for _, c := range clients {
			cc, err := s.CreateClient(ctx, &c)
			if err != nil {
				t.Fatal(err)
			}

			if other, err := s.Client(ctx, cc.ID); err != nil {
				t.Fatal(err)
			} else if other.Type != cc.Type {
				t.Fatalf("mismatch: %#v != %#v", cc.Type, other.Type)
			}
		}
	})
}

func TestClientService_Client(t *testing.T) {
	argonHasher := inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})
	s := postgres.NewClientService(pool, argonHasher)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		client, _ := MustCreateClient(t, ctx, pool, &yeahapi.Client{
			Name:   "Client",
			Secret: "whatever",
			Type:   yeahapi.ClientConfidential,
		})

		if _, err := s.Client(ctx, client.ID); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrClientNotFound", func(t *testing.T) {
		id, _ := uuid.NewV7()
		_, err := s.Client(context.Background(), yeahapi.ClientID{id})
		if !yeahapi.EIs(yeahapi.ENotFound, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestClientService_VerifySecret(t *testing.T) {
	argonHasher := inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})
	s := postgres.NewClientService(pool, argonHasher)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		client, _ := MustCreateClient(t, ctx, pool, &yeahapi.Client{
			Name:   "client",
			Secret: "hello",
			Type:   yeahapi.ClientInternal,
		})

		client, err := s.Client(ctx, client.ID)
		if err != nil {
			t.Fatal(err)
		}

		if err := s.VerifySecret(client, "hello"); err != nil {
			t.Fatal(err)
		}
	})
}

func MustCreateClient(tb testing.TB, ctx context.Context, pool *pgxpool.Pool, client *yeahapi.Client) (*yeahapi.Client, context.Context) {
	tb.Helper()
	argonHasher := inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})
	if _, err := postgres.NewClientService(pool, argonHasher).CreateClient(ctx, client); err != nil {
		tb.Fatal(err)
	}

	return client, yeahapi.NewContextWithClient(ctx, client)
}
