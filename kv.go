package yeahapi

import "context"

type KVItem struct {
	Key      string
	Value    string
	ClientID ClientID
}

type KVService interface {
	Set(ctx context.Context, item *KVItem) (*KVItem, error)
	Get(ctx context.Context, clientID ClientID, key string) (*KVItem, error)
	Remove(ctx context.Context, clientID ClientID, key string) error
}

func (i *KVItem) Ok() error {
	if i.ClientID.IsNil() {
		return E(EInvalid, "Client id is required")
	} else if i.Value == "" {
		return E(EInvalid, "Value is required")
	}
	return nil
}
