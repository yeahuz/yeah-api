package http

import (
	"context"
	"net/http"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerCategoryRoutes() {
	s.mux.Handle("/categories.getCategories", post(s.clientOnly(s.handleGetCategories())))
	s.mux.Handle("/categories.getAttributes", post(s.clientOnly(s.handleGetAttributes())))
}

func (s *Server) handleGetCategories() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return JSON(w, r, http.StatusOK, nil)
	}
}

func (s *Server) handleGetAttributes() Handler {
	const op yeahapi.Op = "http/category.handleGetAttributes"
	return func(w http.ResponseWriter, r *http.Request) error {

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()
		attributes, err := s.CategoryService.Attributes(ctx, "1", "en")
		if err != nil {
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, attributes)
	}
}
