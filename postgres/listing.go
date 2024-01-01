package postgres

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type ListingService struct {
	pool *pgxpool.Pool
}

func NewListingService(pool *pgxpool.Pool) *ListingService {
	return &ListingService{
		pool: pool,
	}
}

func (l *ListingService) Listing(ctx context.Context, id uuid.UUID) (*yeahapi.Listing, error) {
	const op yeahapi.Op = "postgres/ListingService.CreateListing"
	var listing yeahapi.Listing
	err := l.pool.QueryRow(ctx,
		"select id, title, owner_id, category_id from listings where id = $1", id).Scan(&listing.ID, &listing.Title, &listing.OwnerID, &listing.CategoryID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}
		return nil, yeahapi.E(op, err)
	}

	return &listing, nil
}

func (l *ListingService) CreateListing(ctx context.Context, listing *yeahapi.Listing) (*yeahapi.Listing, error) {
	const op yeahapi.Op = "postgres/ListingService.CreateListing"
	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	listing.ID = id
	_, err = l.pool.Exec(ctx,
		"insert into listings (id, title, owner_id, category_id) values ($1, $2, $3, $4)",
		listing.ID, listing.Title, listing.OwnerID, listing.CategoryID)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return listing, nil
}

func (l *ListingService) DeleteListing(ctx context.Context, id uuid.UUID) error {
	const op yeahapi.Op = "postgres/ListingService.DeleteListing"
	return yeahapi.E(op, yeahapi.ENotImplemented)
}
