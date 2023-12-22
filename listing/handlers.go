package listing

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/yeahuz/yeah-api/auth"
	c "github.com/yeahuz/yeah-api/common"
)

func HandleCreateListing() c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var listingData createListingData
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&listingData); err != nil {
			return err
		}

		if err := listingData.validate(); err != nil {
			return err
		}

		session := r.Context().Value("session").(*auth.Session)

		listing, err := newListing(session.UserID, listingData.Title, listingData.CategoryID)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := listing.save(ctx); err != nil {
			return err
		}

		return c.JSON(w, http.StatusOK, listing)
	}
}
