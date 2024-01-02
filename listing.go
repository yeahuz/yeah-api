package yeahapi

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

type ListingStatus string

const (
	ListingStatusActive     ListingStatus = "ACTIVE"
	ListingStatusModeration ListingStatus = "MODERATION"
	ListingStatusIndexing   ListingStatus = "INDEXING"
	ListingStatusArchived   ListingStatus = "ARCHIVED"
	ListingStatusDraft      ListingStatus = "DRAFT"
	ListingStatusDeleted    ListingStatus = "DELETED"
)

type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyUZS Currency = "UZS"
)

type Category struct {
	ID          string
	ParentID    string
	Title       string
	Description string
}

type ListingSkuPrice struct {
	SkuID     string
	Amount    int
	Currency  Currency
	StartDate time.Time
}

type ListingSku struct {
	ID        uuid.UUID
	CustomSku string
	ListingID uuid.UUID
	Price     ListingSkuPrice
}

type Listing struct {
	ID         uuid.UUID     `json:"id"`
	Title      string        `json:"title"`
	CategoryID string        `json:"category_id"`
	OwnerID    UserID        `json:"owner_id"`
	Status     ListingStatus `json:"status"`
	Skus       []ListingSku
}

type ListingService interface {
	CreateListing(ctx context.Context, listing *Listing) (*Listing, error)
	Listing(ctx context.Context, id uuid.UUID) (*Listing, error)
	DeleteListing(ctx context.Context, id uuid.UUID) error
}

type CategoryService interface {
	Categories(ctx context.Context, lang string) []Category
}
