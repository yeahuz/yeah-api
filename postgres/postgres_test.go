package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	pgContainer, err := postgres.RunContainer(context.Background(),
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithInitScripts("migrations/20231122101049_initial.up.sql"),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second),
		),
	)

	if err != nil {
		os.Exit(1)
	}

	connStr, err := pgContainer.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		os.Exit(1)
	}

	if pool, err = pgxpool.New(context.Background(), connStr); err != nil {
		os.Exit(1)
	}

	os.Exit(m.Run())
}
