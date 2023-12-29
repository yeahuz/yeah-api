package yeahapi

import "context"

type ClientID string

type clientType string

var (
	ClientInternal     clientType = "internal"
	ClientConfidential clientType = "confidential"
	ClientPublic       clientType = "public"
)

type Client struct {
	ID     ClientID `json:"id"`
	Name   string   `json:"name"`
	Secret string
	Type   clientType `json:"type"`
	Active bool       `json:"active"`
}

type ClientService interface {
	Client(ctx context.Context, id ClientID) (*Client, error)
	VerifySecret(client *Client, secret string) error
}
