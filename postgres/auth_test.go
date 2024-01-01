package postgres_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/inmem"
	"github.com/yeahuz/yeah-api/postgres"
)

const highwayHashKey = "7dc06c4157760bcae3f24c3aa3d63c9dd74ad8ea714000675ef2c1eebb5ad4ad"

func TestAuthService_CreateOtp(t *testing.T) {
	var argonHasher = inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	var highwayHasher = inmem.NewHighwayHasher(highwayHashKey)
	var s = postgres.NewAuthService(pool, argonHasher, highwayHasher)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		otp := &yeahapi.Otp{
			Identifier: randEmail(),
			ExpiresAt:  time.Now().Add(time.Minute * 15),
		}

		if other, err := s.CreateOtp(ctx, otp); err != nil {
			t.Fatal(err)
		} else if other.Code == "" {
			t.Fatal("otp code not generated")
		} else if other.Hash == "" {
			t.Fatal("otp hash not generated")
		} else if !reflect.DeepEqual(other, otp) {
			t.Fatalf("mismatch: %#v != %#v", other, otp)
		}
	})
}

func TestAuthService_VerifyOtp(t *testing.T) {
	var argonHasher = inmem.NewArgonHasher(yeahapi.ArgonParams{
		SaltLen: 15,
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})

	var highwayHasher = inmem.NewHighwayHasher(highwayHashKey)
	var s = postgres.NewAuthService(pool, argonHasher, highwayHasher)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		otp, err := s.CreateOtp(ctx, &yeahapi.Otp{
			Identifier: randEmail(),
			ExpiresAt:  time.Now().Add(time.Minute * 15),
		})

		if err != nil {
			t.Fatal(err)
		}

		if err := s.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       otp.Hash,
			Code:       otp.Code,
			Identifier: otp.Identifier,
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ErrOtpExpired", func(t *testing.T) {
		ctx := context.Background()

		otp, err := s.CreateOtp(ctx, &yeahapi.Otp{
			Identifier: randEmail(),
			ExpiresAt:  time.Now().Add(time.Second * 2),
		})

		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Second * 3)

		if err := s.VerifyOtp(ctx, otp); !yeahapi.EIs(yeahapi.EOtpCodeExpired, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})

	t.Run("ErrHashNotMatched", func(t *testing.T) {
		ctx := context.Background()
		otp, err := s.CreateOtp(ctx, &yeahapi.Otp{
			Identifier: randEmail(),
			ExpiresAt:  time.Now().Add(time.Second * 10),
		})

		if err != nil {
			t.Fatal(err)
		}

		if err := s.VerifyOtp(ctx, &yeahapi.Otp{
			Hash:       otp.Hash,
			Code:       otp.Code,
			Identifier: randEmail(),
		}); !yeahapi.EIs(yeahapi.EOtpHashNotMatched, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestAuthService_Otp(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
	})
}

func TestAuthService_CreateAuth(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
	})
}

func TestAuthService_DeleteAuth(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
	})
}

func TestAuthService_Session(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
	})
}
