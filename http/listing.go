package http

import (
	"context"
	"net/http"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerListingRoutes() {
	s.mux.Handle("/listings.createListing", post(s.userOnly(s.handleCreateListing())))
}

type createListingData struct {
	Title      string `json:"title"`
	CategoryID string `json:"category_id"`
}

func (d createListingData) Ok() error {
	if d.Title == "" {
		return yeahapi.E(yeahapi.EInvalid, "Title is required")
	}
	if d.CategoryID == "" {
		return yeahapi.E(yeahapi.EInvalid, "Category is required")
	}
	return nil
}

func (s *Server) handleCreateListing() Handler {
	const op yeahapi.Op = "http/listings.handleCreateListing"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req createListingData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		session := yeahapi.SessionFromContext(ctx)
		listing, err := s.ListingService.CreateListing(ctx, &yeahapi.Listing{
			CategoryID: req.CategoryID,
			Title:      req.Title,
			OwnerID:    session.UserID,
		})
		if err != nil {
			return yeahapi.E(op, err, "Couldn't create listing. Please, try again")
		}

		return JSON(w, r, http.StatusOK, listing)
	}
}
