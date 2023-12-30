package postgres_test

import (
	"context"
	"reflect"
	"testing"

	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/postgres"
)

func TestKVService_Set(t *testing.T) {
	ctx := context.Background()
	client, _ := MustCreateClient(t, ctx, pool, &yeahapi.Client{
		Name:   "Client",
		Secret: "whatever",
		Type:   yeahapi.ClientConfidential,
	})

	s := postgres.NewKVService(pool)

	t.Run("OK", func(t *testing.T) {
		updates := [3]yeahapi.KVItem{
			{
				Key:      "hello",
				Value:    "world",
				ClientID: client.ID,
			},
			{
				Key:      "hello",
				Value:    "not-world",
				ClientID: client.ID,
			},
			{
				Key:      "hello",
				Value:    "hello-hello",
				ClientID: client.ID,
			},
		}

		for _, v := range updates {
			item, err := s.Set(ctx, &v)
			if err != nil {
				t.Fatal(err)
			}

			if other, err := s.Get(ctx, client.ID, v.Key); err != nil {
				t.Fatal(err)
			} else if !reflect.DeepEqual(item, other) {
				t.Fatalf("mismatch: %#v != %#v", item, other)
			}
		}
	})
}

func TestKVService_Get(t *testing.T) {
	ctx := context.Background()
	client, _ := MustCreateClient(t, ctx, pool, &yeahapi.Client{
		Name:   "Client",
		Secret: "whatever",
		Type:   yeahapi.ClientConfidential,
	})

	s := postgres.NewKVService(pool)
	t.Run("OK", func(t *testing.T) {
		item := &yeahapi.KVItem{
			Key:      "hello",
			Value:    "world",
			ClientID: client.ID,
		}
		item, err := s.Set(ctx, item)
		if err != nil {
			t.Fatal(err)
		}

		if other, err := s.Get(ctx, client.ID, item.Key); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(item, other) {
			t.Fatalf("mismatch: %#v != %#v", item, other)
		}
	})

	t.Run("ErrKeyNotFound", func(t *testing.T) {
		_, err := s.Get(ctx, client.ID, "non-existing-key")
		if !yeahapi.EIs(yeahapi.ENotExist, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestKVService_Remove(t *testing.T) {
	ctx := context.Background()
	client, _ := MustCreateClient(t, ctx, pool, &yeahapi.Client{
		Name:   "Client",
		Secret: "whatever",
		Type:   yeahapi.ClientConfidential,
	})

	s := postgres.NewKVService(pool)
	t.Run("OK", func(t *testing.T) {
		item := &yeahapi.KVItem{
			Key:      "hola",
			Value:    "hola, test",
			ClientID: client.ID,
		}

		item, err := s.Set(ctx, item)
		if err != nil {
			t.Fatal(err)
		}

		if err := s.Remove(ctx, client.ID, item.Key); err != nil {
			t.Fatal(err)
		}

		if _, err := s.Get(ctx, client.ID, item.Key); err != nil && !yeahapi.EIs(yeahapi.ENotExist, err) {
			t.Fatalf("unxpected error: %#v", err)
		}
	})
}
