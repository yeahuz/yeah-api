package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerCategoryRoutes() {
	s.mux.Handle("/categories.getCategories", get(s.clientOnly(s.handleGetCategories())))
	s.mux.Handle("/categories.getAttributes", get(s.clientOnly(s.handleGetAttributes())))
}

func (s *Server) handleGetCategories() Handler {
	const op yeahapi.Op = "http/category.handleGetCategories"
	type response struct {
		T          string             `json:"_"`
		Categories []yeahapi.Category `json:"categories"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		categories, err := s.CategoryService.Categories(ctx, "en")
		if err != nil {
			fmt.Println(err)
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, response{"listing.categories", categories})
	}
}

func (s *Server) handleGetAttributes() Handler {
	const op yeahapi.Op = "http/category.handleGetAttributes"
	type response struct {
		T          string                       `json:"_"`
		Attributes []*yeahapi.CategoryAttribute `json:"attributes"`
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		attributes, err := s.CategoryService.Attributes(ctx, "1", "en")
		if err != nil {
			fmt.Println(err)
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, response{"listing.categoryAttributes", attributes})
	}
}
