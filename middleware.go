package main

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/yeahuz/yeah-api/auth"
	"github.com/yeahuz/yeah-api/client"
	c "github.com/yeahuz/yeah-api/common"
	"github.com/yeahuz/yeah-api/internal/errors"
	"github.com/yeahuz/yeah-api/internal/localizer"
)

func clientOnly(next c.ApiFunc) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		clientId := r.Header.Get("X-Client-Id")

		if clientId == "" {
			return errors.Unauthorized
		}

		clientSecret := r.Header.Get("X-Client-Secret")
		client, err := client.GetById(ctx, clientId)
		if err != nil {
			return err
		}

		if err := client.Verify(clientSecret); err != nil {
			return err
		}

		ctx = context.WithValue(r.Context(), "client", client)
		return next(w, r.WithContext(ctx))
	}
}

func localized(next c.ApiFunc) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		lang := r.Header.Get("Accept-Language")
		l := localizer.Get(lang)
		ctx := context.WithValue(r.Context(), "localizer", l)
		return next(w, r.WithContext(ctx))
	}
}

func userOnly(next c.ApiFunc) c.ApiFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		sessionId := r.Header.Get("X-Session-Id")
		if sessionId == "" {
			return errors.Unauthorized
		}

		l := r.Context().Value("localizer").(localizer.Localizer)
		if !isValidUUID(sessionId) {
			return errors.NewBadRequest(l.T("Missing valid session id"))
		}

		session, err := auth.GetSessionById(ctx, sessionId)
		if err != nil {
			return err
		}

		if !session.Active {
			return errors.Unauthorized
		}

		ctx = context.WithValue(r.Context(), "session", session)
		return next(w, r.WithContext(ctx))
	}
}

func isValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}
