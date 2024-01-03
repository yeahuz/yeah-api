package yeahapi

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
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

type ListingSkuPrice struct {
	SkuID     string
	Amount    int
	Currency  Currency
	StartDate time.Time
}

// type ElectronicAttrs struct {
// 	Brand string
// }

// type PhoneAttrs struct {
// 	ElectronicAttrs
// 	Model           string
// 	StorageCapacity string
// 	Color           string
// 	Ram             string
// }

type ListingAttrs map[string]interface{}

type ListingSku struct {
	ID        uuid.UUID
	CustomSku string
	ListingID uuid.UUID
	Price     ListingSkuPrice
	ListingAttrs
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
	CreateSku(ctx context.Context, sku *ListingSku) (*ListingSku, error)
	DeleteSku(ctx context.Context, id uuid.UUID) error
	Skus(ctx context.Context, listingID uuid.UUID) ([]ListingSku, error)
}

func (a ListingAttrs) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *ListingAttrs) Scan(value interface{}) error {
	b, ok := value.([]byte)

	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
