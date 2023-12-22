package listing

import "github.com/gofrs/uuid"

type Listing struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	CategoryID string    `json:"category_id"`
	OwnerID    uuid.UUID `json:"owner_id"`
}

type createListingData struct {
	Title      string `json:"title"`
	CategoryID string `json:"category_id"`
}
