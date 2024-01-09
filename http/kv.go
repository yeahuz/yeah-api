package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	yeahapi "github.com/yeahuz/yeah-api"
)

func (s *Server) registerKVRoutes() {
	s.mux.Handle("/kv.set", post(s.clientOnly(s.handleKVSet())))
	s.mux.Handle("/kv.get", post(s.clientOnly(s.handleKVGet())))
	s.mux.Handle("/kv.remove", post(s.clientOnly(s.handleKVRemove())))
}

type keyData struct {
	Key string `json:"key"`
}

type kvData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (d keyData) Ok() error {
	if d.Key == "" {
		return yeahapi.E(yeahapi.EInvalid, "Key is required")
	}
	return nil
}

func (d kvData) Ok() error {
	if d.Value == "" {
		return yeahapi.E(yeahapi.EInvalid, "Value is required")
	}
	return nil
}

func (s *Server) handleKVSet() Handler {
	const op yeahapi.Op = "http/kv.handleKVSet"
	type response struct {
		T string `json:"_"`
		*yeahapi.KVItem
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var req kvData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		client := yeahapi.ClientFromContext(ctx)
		item, err := s.KVService.Set(ctx, &yeahapi.KVItem{
			ClientID: client.ID,
			Key:      req.Key,
			Value:    req.Value,
		})

		if err != nil {
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		return JSON(w, r, http.StatusOK, response{"kv.item", item})
	}
}

func (s *Server) handleKVGet() Handler {
	const op yeahapi.Op = "http/kv.handleKVGet"
	type response struct {
		T string `json:"_"`
		*yeahapi.KVItem
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		var req keyData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		client := yeahapi.ClientFromContext(ctx)
		item, err := s.KVService.Get(ctx, client.ID, req.Key)
		if err != nil {
			if yeahapi.EIs(yeahapi.ENotFound, err) {
				return yeahapi.E(op, err, fmt.Sprintf("Value with key %s not found", req.Key))
			}
			return yeahapi.E(op, err, "Something went wrong on our end. Please, try again later")
		}

		return JSON(w, r, http.StatusOK, response{"kv.item", item})
	}
}

func (s *Server) handleKVRemove() Handler {
	const op yeahapi.Op = "http/kv.handleKVRemove"
	return func(w http.ResponseWriter, r *http.Request) error {
		var req keyData
		defer r.Body.Close()
		if err := decode(r, &req); err != nil {
			return yeahapi.E(op, err)
		}

		ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
		defer cancel()

		client := yeahapi.ClientFromContext(ctx)

		if err := s.KVService.Remove(ctx, client.ID, req.Key); err != nil {
			return yeahapi.E(op, err)
		}

		return JSON(w, r, http.StatusOK, nil)
	}
}
