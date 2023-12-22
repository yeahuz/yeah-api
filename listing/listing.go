package listing

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/yeahuz/yeah-api/db"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

var l = localizer.GetDefault()

func newListing(ownerID uuid.UUID, title string, categoryID string) (*Listing, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	return &Listing{
		ID:         id,
		Title:      title,
		OwnerID:    ownerID,
		CategoryID: categoryID,
	}, nil
}

func (l *Listing) save(ctx context.Context) error {
	_, err := db.Pool.Exec(ctx,
		"insert into listings (id, title, owner_id, category_id) values ($1, $2, $3, $4)",
		l.ID, l.Title, l.OwnerID, l.CategoryID,
	)
	return err
}

func (d createListingData) validate() error {
	errs := make(map[string]string)

	if len(d.Title) == 0 {
		errs["title"] = l.T("Listing title is required")
	}

	if len(d.CategoryID) == 0 {
		errs["category_id"] = l.T("Listing category is required")
	}

	if len(errs) > 0 {
		return errors.NewValidation(errs)
	}

	return nil
}
