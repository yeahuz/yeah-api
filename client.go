package yeahapi

import "context"

type ClientID string

type clientType string

var (
	clientInternal     clientType = "internal"
	clientConfidential clientType = "confidential"
	clientPublic       clientType = "public"
)

type Client struct {
	ID     ClientID `json:"id"`
	Name   string   `json:"name"`
	secret string
	Type   clientType `json:"type"`
	Active bool       `json:"active"`
}

type ClientService interface {
	Client(ctx context.Context, id ClientID) (*Client, error)
}
