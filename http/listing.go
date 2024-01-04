package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerListingRoutes() {
	s.mux.Handle("/listings.createListing", post(s.userOnly(s.handleCreateListing())))
	s.mux.Handle("/listings.getListing", post(s.userOnly(s.handleGetListing())))
	s.mux.Handle("/listings.deleteListing", post(s.userOnly(s.handleDeleteListing())))
	s.mux.Handle("/listings.createSku", post(s.userOnly(s.handleCreateSku())))
	s.mux.Handle("/listings.deleteSku", post(s.userOnly(s.handleDeleteSku())))
	s.mux.Handle("/listings.getSkus", post(s.userOnly(s.handleGetSkus())))
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

func (s *Server) handleGetListing() Handler {
	const op yeahapi.Op = "http/listings.handleGetListing"
	type request struct {
		ID uuid.UUID `json:"listing_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		listing, err := s.ListingService.Listing(ctx, req.ID)
		if err != nil {
			return yeahapi.E(op, err)
		}
		return JSON(w, r, http.StatusOK, listing)
	}
}

func (s *Server) handleDeleteListing() Handler {
	const op yeahapi.Op = "http/listings.handleDeleteListing"
	type request struct {
		ID uuid.UUID `json:"listing_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		if err := s.ListingService.DeleteListing(ctx, req.ID); err != nil {
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, nil)
	}
}

type createSkuData struct {
	ListingID uuid.UUID            `json:"listing_id"`
	UnitPrice int                  `json:"unit_price"`
	Currency  yeahapi.Currency     `json:"currency"`
	CustomSku string               `json:"custom_sku"`
	Attrs     yeahapi.ListingAttrs `json:"attrs"`
}

func (d createSkuData) Ok() error {
	if d.Currency == "" {
		return yeahapi.E(yeahapi.EInvalid, "Currency is required")
	}

	if d.Currency != yeahapi.CurrencyUSD {
		return yeahapi.E(yeahapi.EInvalid, "Invalid currency. Use USD instead")
	}
	return nil
}

func (s *Server) handleCreateSku() Handler {
	const op yeahapi.Op = "http/listings.handleCreateSku"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req createSkuData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		sku, err := s.ListingService.CreateSku(ctx, &yeahapi.ListingSku{
			ListingID: req.ListingID,
			Price: yeahapi.ListingSkuPrice{
				Amount:    req.UnitPrice,
				Currency:  req.Currency,
				StartDate: time.Now(),
			},
			CustomSku:    req.CustomSku,
			ListingAttrs: req.Attrs,
		})

		if err != nil {
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, sku)
	}
}

func (s *Server) handleDeleteSku() Handler {
	const op yeahapi.Op = "http/listings.handleDeleteSku"
	type request struct {
		ID uuid.UUID `json:"sku_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		if err := s.ListingService.DeleteSku(ctx, req.ID); err != nil {
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, nil)
	}
}

func (s *Server) handleGetSkus() Handler {
	const op yeahapi.Op = "http/listings.handleGetSkus"
	type request struct {
		ID uuid.UUID `json:"listing_id"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var req request
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		skus, err := s.ListingService.Skus(ctx, req.ID)
		if err != nil {
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, skus)
	}
}
