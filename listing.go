package yeahapi

import (
	"context"

	"github.com/gofrs/uuid"
)

type Listing struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	CategoryID string    `json:"category_id"`
	OwnerID    UserID    `json:"owner_id"`
}

type ListingService interface {
	CreateListing(ctx context.Context, listing *Listing) (*Listing, error)
	Listing(ctx context.Context, id uuid.UUID) (*Listing, error)
	DeleteListing(ctx context.Context, id uuid.UUID) error
}
