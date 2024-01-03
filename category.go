package yeahapi

import "context"

type Category struct {
	ID          string
	ParentID    string
	Title       string
	Description string
}

type CategoryReference struct {
	TableName  string
	CategoryID string
	Columns    []string
}

type CategoryAttribute struct {
	ID                   string
	Required             bool
	EnabledForVariations bool
	Key                  string
	CategoryID           string
	Options              []CategoryAttributeOption
}

type CategoryAttributeOption struct {
	ID          string
	AttributeID string
	Value       string
	Unit        string
}

type CategoryService interface {
	Categories(ctx context.Context, lang string) ([]Category, error)
	References(ctx context.Context) ([]CategoryReference, error)
	Attributes(ctx context.Context, categoryID string) ([]CategoryAttribute, error)
}
