package yeahapi

import "context"

type contextKey int

const (
	sessionContextKey = contextKey(iota + 1)
	clientContextKey
	localizerContextKey
	flashContextKey
)

func NewContextWithLocalizer(ctx context.Context, localizer *Localizer) context.Context {
	return context.WithValue(ctx, localizerContextKey, localizer)
}

func NewContextWithClient(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, clientContextKey, client)
}

func NewContextWithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

func NewContextWithFlash(ctx context.Context, flash Flash) context.Context {
	return context.WithValue(ctx, flashContextKey, flash)
}

func LocalizerFromContext(ctx context.Context) *Localizer {
	localizer, _ := ctx.Value(localizerContextKey).(*Localizer)
	return localizer
}

func SessionFromContext(ctx context.Context) *Session {
	session, _ := ctx.Value(sessionContextKey).(*Session)
	return session
}

func ClientFromContext(ctx context.Context) *Client {
	client, _ := ctx.Value(clientContextKey).(*Client)
	return client
}

func FlashFromContext(ctx context.Context) Flash {
	message, _ := ctx.Value(flashContextKey).(Flash)
	return message
}
