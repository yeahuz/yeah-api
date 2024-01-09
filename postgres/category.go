package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	yeahapi "github.com/yeahuz/yeah-api"
)

type CategoryService struct {
	pool *pgxpool.Pool
}

func NewCategoryService(pool *pgxpool.Pool) *CategoryService {
	return &CategoryService{
		pool: pool,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, category *yeahapi.Category) (*yeahapi.Category, error) {
	const op yeahapi.Op = "postgres/CategoryService.CreateCategory"

	if _, err := s.pool.Exec(ctx, "insert into category"); err != nil {
		return nil, yeahapi.E(op, err)
	}
	return nil, nil
}

func (s *CategoryService) Categories(ctx context.Context, lang string) ([]yeahapi.Category, error) {
	const op yeahapi.Op = "postgres/CategoryService.Categories"
	categories := make([]yeahapi.Category, 0)

	rows, err := s.pool.Query(ctx, "select c.id, coalesce(c.parent_id, 0), ct.title, ct.description from categories c left join categories_tr ct on ct.category_id = c.id and ct.lang_code = $1", lang)

	defer rows.Close()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	for rows.Next() {
		var c yeahapi.Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Title, &c.Description); err != nil {
			return nil, yeahapi.E(op, err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return categories, nil
}

func (s *CategoryService) References(ctx context.Context) ([]yeahapi.CategoryReference, error) {
	const op yeahapi.Op = "postgres/CategoryService.References"
	references := make([]yeahapi.CategoryReference, 0)

	rows, err := s.pool.Query(ctx, "select table_name, category_id, columns from category_reference")

	defer rows.Close()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	for rows.Next() {
		var r yeahapi.CategoryReference
		if err := rows.Scan(&r.TableName, &r.CategoryID, &r.Columns); err != nil {
			return nil, yeahapi.E(op, err)
		}

		references = append(references, r)
	}

	if err := rows.Err(); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return references, nil
}

func (s *CategoryService) Attributes(ctx context.Context, categoryID string, lang string) ([]*yeahapi.CategoryAttribute, error) {
	const op yeahapi.Op = "postgres/CategoryService.Attributes"

	rows, err := s.pool.Query(ctx,
		`select a.id, a.required, a.enabled_for_variations, a.key, a.category_id, at.name,
		ao.id as option_id, coalesce(aot.name, ao.value) as option_name, ao.value as option_value, ao.unit as option_unit, ao.attribute_id as option_attribute_id
		from attributes a
		left join attributes_tr at on at.attribute_id = a.id and at.lang_code = $1
		left join attribute_options ao on ao.attribute_id = a.id
		left join attribute_options_tr aot on aot.attribute_option_id = ao.id and aot.lang_code = $1
		where a.category_id = $2
		`,
		lang, categoryID)

	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	defer rows.Close()
	attributes := make([]*yeahapi.CategoryAttribute, 0)
	var currentAtr *yeahapi.CategoryAttribute

	for rows.Next() {
		var id, name, categoryID, key string
		var required, enabledForVariations bool
		var opt yeahapi.CategoryAttributeOption

		err := rows.Scan(&id, &required, &enabledForVariations, &key, &categoryID, &name, &opt.ID, &opt.Name, &opt.Value, &opt.Unit, &opt.AttributeID)
		if err != nil {
			return nil, yeahapi.E(op, err)
		}

		if currentAtr == nil || currentAtr.ID != id {
			currentAtr = &yeahapi.CategoryAttribute{
				ID:                   id,
				Name:                 name,
				CategoryID:           categoryID,
				Key:                  key,
				Required:             required,
				EnabledForVariations: enabledForVariations,
			}

			attributes = append(attributes, currentAtr)
		}

		currentAtr.Options = append(currentAtr.Options, opt)
	}

	return attributes, nil
}
