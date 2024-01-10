package postgres_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/postgres"
)

func TestListingService_CreateListing(t *testing.T) {
	s := postgres.NewListingService(pool)
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		listing := MustCreateListing(t, ctx, pool)
		if other, err := s.Listing(ctx, listing.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(other, listing) {
			t.Fatalf("mismatch: %#v != %#v", other, listing)
		}
	})
}

func TestListingService_DeleteListing(t *testing.T) {
	s := postgres.NewListingService(pool)
	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		listing := MustCreateListing(t, ctx, pool)

		if err := s.DeleteListing(ctx, listing.ID); err != nil {
			t.Fatal(err)
		}

		if _, err := s.Listing(ctx, listing.ID); err != nil && !yeahapi.EIs(yeahapi.ENotFound, err) {
			t.Fatalf("unexpected error: %#v", err)
		}
	})
}

func TestListingService_CreateSku(t *testing.T) {
	s := postgres.NewListingService(pool)

	t.Run("OK", func(t *testing.T) {
		ctx := context.Background()
		listing := MustCreateListing(t, ctx, pool)

		sku, err := s.CreateSku(ctx, &yeahapi.ListingSku{
			ListingID:     listing.ID,
			Price:         299,
			PriceCurrency: yeahapi.CurrencyUSD,
			Attrs: yeahapi.ListingAttrs{
				"ram":   "8 GB",
				"model": "Iphone 14 Pro Max",
			},
		})

		if err != nil {
			t.Fatal(err)
		}

		if other, err := s.Sku(ctx, sku.ID); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(other, sku) {
			t.Fatalf("mismatch: %#v != %#v", other, sku)
		}
	})
}

func MustCreateListing(tb testing.TB, ctx context.Context, pool *pgxpool.Pool) *yeahapi.Listing {
	tb.Helper()
	user := MustCreateUser(tb, ctx, pool, &yeahapi.User{
		Email:     randEmail(),
		FirstName: "John",
		LastName:  "Doe",
	})

	category := MustCreateCategory(tb, ctx, pool, &yeahapi.Category{})

	listing, err := postgres.NewListingService(pool).CreateListing(ctx, &yeahapi.Listing{
		Title:      "Hello world, listing",
		OwnerID:    user.ID,
		CategoryID: category.ID,
		Status:     yeahapi.ListingStatusDraft,
	})

	if err != nil {
		tb.Fatal(err)
	}

	return listing
}
