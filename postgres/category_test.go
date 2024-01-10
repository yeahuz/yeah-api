package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
	"github.com/yeahuz/yeah-api/postgres"
)

func TestCategoryService_CreateCategory(t *testing.T) {
	t.Run("OK", func(t *testing.T) {})
}

func TestCategoryService_Categories(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
	})
}

func TestCategoryService_Attributes(t *testing.T) {
	t.Run("OK", func(t *testing.T) {})
}

func MustCreateCategory(tb testing.TB, ctx context.Context, pool *pgxpool.Pool, category *yeahapi.Category) *yeahapi.Category {
	tb.Helper()
	if _, err := postgres.NewCategoryService(pool).CreateCategory(ctx, category); err != nil {
		tb.Fatal(err)
	}

	return category
}
