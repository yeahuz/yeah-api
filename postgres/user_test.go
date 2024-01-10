package postgres_test

import (
	"context"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/postgres"
)

func randStr(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

func randEmail() string {
	return randStr(10) + "@" + randStr(6)
}

func TestUserService_User(t *testing.T) {
	s := postgres.NewUserService(pool)
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()

		u := MustCreateUser(t, ctx, pool, &yeahapi.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     randEmail(),
		})

		if other, err := s.User(ctx, u.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(u, other) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
	})

	t.Run("ErrUserNotFound", func(t *testing.T) {
		id, _ := uuid.NewV7()
		_, err := s.User(context.Background(), yeahapi.UserID{id})
		if !yeahapi.EIs(yeahapi.ENotFound, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestUserService_ByEmail(t *testing.T) {
	s := postgres.NewUserService(pool)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		u := MustCreateUser(t, ctx, pool, &yeahapi.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     randEmail(),
		})

		if other, err := s.ByEmail(ctx, u.Email); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(other, u) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
	})
}

func TestUserService_ByPhone(t *testing.T) {
	s := postgres.NewUserService(pool)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		u := MustCreateUser(t, ctx, pool, &yeahapi.User{
			FirstName:   "John",
			LastName:    "Doe",
			Email:       randEmail(),
			PhoneNumber: strconv.Itoa(9999999 + rand.Intn(9999999)),
		})

		if other, err := s.ByPhone(ctx, u.PhoneNumber); err != nil {
			t.Fatal(err)
		} else if other.PhoneNumber != u.PhoneNumber {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
	})

	t.Run("ErrUserNotFoundByPhone", func(t *testing.T) {
		_, err := s.ByPhone(context.Background(), "00000000000")
		if !yeahapi.EIs(yeahapi.ENotFound, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestUserService_CreateUser(t *testing.T) {
	s := postgres.NewUserService(pool)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		u, err := s.CreateUser(ctx, &yeahapi.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     randEmail(),
		})

		if err != nil {
			t.Fatal(err)
		}

		if other, err := s.User(ctx, u.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(other, u) {
			t.Fatalf("mismatch: %#v != %#v", u, other)
		}
	})
}

func TestUserService_LinkAccount(t *testing.T) {
	s := postgres.NewUserService(pool)
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		user := MustCreateUser(t, ctx, pool, &yeahapi.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     randEmail(),
		})

		if err := s.LinkAccount(ctx, &yeahapi.Account{
			UserID:            user.ID,
			Provider:          "google",
			ProviderAccountID: randStr(20),
		}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestUserService_Account(t *testing.T) {
	s := postgres.NewUserService(pool)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		user := MustCreateUser(t, ctx, pool, &yeahapi.User{
			FirstName: "John",
			LastName:  "Doe",
			Email:     randEmail(),
		})

		account := &yeahapi.Account{
			UserID:            user.ID,
			Provider:          "google",
			ProviderAccountID: randStr(20),
		}

		if err := s.LinkAccount(ctx, account); err != nil {
			t.Fatal(err)
		}

		if other, err := s.Account(ctx, account.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(other, account) {
			t.Fatalf("mismatch: %#v != %#v", account, other)
		}
	})

	t.Run("ErrAccountNotFound", func(t *testing.T) {
		id, _ := uuid.NewV7()
		_, err := s.Account(context.Background(), id)
		if !yeahapi.EIs(yeahapi.ENotFound, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func MustCreateUser(tb testing.TB, ctx context.Context, pool *pgxpool.Pool, user *yeahapi.User) *yeahapi.User {
	tb.Helper()
	if _, err := postgres.NewUserService(pool).CreateUser(ctx, user); err != nil {
		tb.Fatal(err)
	}

	return user
}
