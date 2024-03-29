package yeahapi

import "context"

type Category struct {
	ID          int    `json:"id"`
	ParentID    *int   `json:"parent_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CategoryReference struct {
	TableName  string
	CategoryID int
	Columns    []string
}

type CategoryAttribute struct {
	ID                   int                       `json:"id"`
	Required             bool                      `json:"required"`
	EnabledForVariations bool                      `json:"enabled_for_variations"`
	Key                  string                    `json:"key"`
	Name                 string                    `json:"name"`
	CategoryID           int                       `json:"category_id"`
	Options              []CategoryAttributeOption `json:"options"`
}

type CategoryAttributeOption struct {
	ID          int    `json:"id"`
	AttributeID int    `json:"attribute_id"`
	Value       string `json:"value"`
	Unit        string `json:"unit"`
	Name        string `json:"name"`
}

type CategoryService interface {
	Categories(ctx context.Context, lang string) ([]Category, error)
	References(ctx context.Context) ([]CategoryReference, error)
	Attributes(ctx context.Context, categoryID string, lang string) ([]*CategoryAttribute, error)
	CreateCategory(ctx context.Context, category *Category) (*Category, error)
}
