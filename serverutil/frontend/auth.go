package frontend

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/yeahuz/yeah-api/serverutil/frontend/templ/auth"
)

func Loading() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<div id='loading'>Loading</div>")
		w.(http.Flusher).Flush()
		return err
	})
}

func Header() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<div>This is the header</div>")
		w.(http.Flusher).Flush()
		return err
	})
}

func Footer() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<div>This is the footer</div>")
		w.(http.Flusher).Flush()
		return err
	})
}

func Content(ch chan struct{}) templ.Component {
	go func() {
		time.Sleep(time.Second * 2)
		ch <- struct{}{}
	}()

	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<div id='content'>Content</div>
    <script>
      let content = document.getElementById('content');
      let loading = document.getElementById('loading');
      loading.replaceWith(content);
    </script>`)
		return err
	})
}

type SuspendibleComponentFunc func(ch chan struct{}) templ.Component

type suspense struct {
	ch       chan struct{}
	fallback templ.Component
	content  templ.Component
}

func (s suspense) Suspend() <-chan struct{} {
	return s.ch
}

func Suspense(fallback templ.Component, content SuspendibleComponentFunc) *suspense {
	ch := make(chan struct{})
	return &suspense{
		ch:       ch,
		fallback: fallback,
		content:  content(ch),
	}
}

func Page() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if err := Header().Render(ctx, w); err != nil {
			return err
		}

		sus := Suspense(Loading(), Content)
		if err := sus.fallback.Render(ctx, w); err != nil {
			return err
		}

		if err := Footer().Render(ctx, w); err != nil {
			return err
		}

		<-sus.Suspend()
		return sus.content.Render(ctx, w)
	})
}

func (s *Server) registerAuthRoutes() {
	s.mux.Handle("/auth/login", routes(map[string]Handler{
		http.MethodGet:  s.handleGetLogin(),
		http.MethodPost: s.handleLogin(),
	}))
}

func (s *Server) handleGetLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		method := r.URL.Query().Get("method")
		return auth.Login(auth.LoginProps{Method: method}).Render(r.Context(), w)
	}
}

func (s *Server) handleLogin() Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		return nil
	}
}
