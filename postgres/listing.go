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

func (s *ListingService) Listing(ctx context.Context, id uuid.UUID) (*yeahapi.Listing, error) {
	const op yeahapi.Op = "postgres/ListingService.Listing"
	var listing yeahapi.Listing
	err := s.pool.QueryRow(ctx,
		"select id, title, owner_id, category_id, status from listings where id = $1", id).Scan(&listing.ID, &listing.Title, &listing.OwnerID, &listing.CategoryID, &listing.Status)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, yeahapi.E(op, yeahapi.ENotFound)
		}
		return nil, yeahapi.E(op, err)
	}

	return &listing, nil
}

func (s *ListingService) CreateListing(ctx context.Context, listing *yeahapi.Listing) (*yeahapi.Listing, error) {
	const op yeahapi.Op = "postgres/ListingService.CreateListing"

	if err := listing.Ok(); err != nil {
		return nil, yeahapi.E(op, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	listing.ID = id
	_, err = s.pool.Exec(ctx,
		"insert into listings (id, title, owner_id, category_id, status) values ($1, $2, $3, $4, $5)",
		listing.ID, listing.Title, listing.OwnerID, listing.CategoryID, listing.Status)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return listing, nil
}

func (s *ListingService) DeleteListing(ctx context.Context, id uuid.UUID) error {
	const op yeahapi.Op = "postgres/ListingService.DeleteListing"
	_, err := s.pool.Exec(ctx, "delete from listings where id = $1", id)
	if err != nil {
		return yeahapi.E(op, err)
	}
	return nil
}

func (s *ListingService) CreateSku(ctx context.Context, sku *yeahapi.ListingSku) (*yeahapi.ListingSku, error) {
	const op yeahapi.Op = "postgres/ListingService.CreateSku"
	id, err := uuid.NewV7()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	sku.ID = id

	_, err = s.pool.Exec(ctx, "insert into listing_skus (id, custom_sku, listing_id, attrs) values ($1, $2, $3, $4)",
		sku.ID, sku.CustomSku, sku.ListingID, sku.Attrs,
	)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	return sku, nil
}

func (s *ListingService) DeleteSku(ctx context.Context, id uuid.UUID) error {
	const op yeahapi.Op = "postgres/ListingService.DeleteSku"

	_, err := s.pool.Exec(ctx, "delete from listing_skus where id = $1", id)
	if err != nil {
		return yeahapi.E(op, err)
	}

	return nil
}

func (s *ListingService) Skus(ctx context.Context, listingID uuid.UUID) ([]yeahapi.ListingSku, error) {
	const op yeahapi.Op = "postgres/ListingService.Skus"
	skus := make([]yeahapi.ListingSku, 0)

	rows, err := s.pool.Query(ctx,
		`select ls.id, ls.custom_sku, ls.listing_id, ls.attrs, lsp.amount, lsp.currency, lsp.start_date from listing_skus ls
		left join listing_sku_prices lsp on lsp.sku_id = (select id from listing_sku_prices where sku_id = ls.id order by start_date desc limit 1)`)

	defer rows.Close()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	for rows.Next() {
		var s yeahapi.ListingSku
		if err := rows.Scan(s.ID, s.CustomSku, s.ListingID, s.Attrs, s.Price.Amount, s.Price.Currency, s.Price.StartDate); err != nil {
			return nil, yeahapi.E(op, err)
		}

		skus = append(skus, s)
	}

	if err := rows.Err(); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return skus, nil
}
