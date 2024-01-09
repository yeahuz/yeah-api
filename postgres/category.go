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

func (s *CategoryService) Categories(ctx context.Context, lang string) ([]yeahapi.Category, error) {
	const op yeahapi.Op = "postgres/CategoryService.Categories"
	categories := make([]yeahapi.Category, 0)

	rows, err := s.pool.Query(ctx, "select c.id, c.parent_id, ct.title, ct.description from categories c left join categories_tr ct on ct.category_id = c.id and ct.lang_code = $1", lang)

	defer rows.Close()
	if err != nil {
		return nil, yeahapi.E(op, err)
	}

	for rows.Next() {
		var c yeahapi.Category
		if err := rows.Scan(&c.ID, c.ParentID, c.Title, c.Description); err != nil {
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
		if err := rows.Scan(r.TableName, r.CategoryID, r.Columns); err != nil {
			return nil, yeahapi.E(op, err)
		}

		references = append(references, r)
	}

	if err := rows.Err(); err != nil {
		return nil, yeahapi.E(op, err)
	}

	return references, nil
}

func (s *CategoryService) Attributes(ctx context.Context, categoryID string, lang string) ([]yeahapi.CategoryAttribute, error) {
	const op yeahapi.Op = "postgres/CategoryService.Attributes"
	attributes := make([]yeahapi.CategoryAttribute, 0)
	attributesMap := map[string]yeahapi.CategoryAttribute{}

	rows, err := s.pool.Query(ctx,
		`select a.id, a.required, a.enabled_for_variations, a.key, a.category_id, at.name,
		ao.id as option_id, coalesce(aot.name, ao.value) as option_name, ao.value as option_value, ao.unit as option_unit
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
	for rows.Next() {
		var atr yeahapi.CategoryAttribute
		var opt yeahapi.CategoryAttributeOption

		err := rows.Scan(atr.ID, atr.Required, atr.EnabledForVariations, atr.Key, atr.CategoryID, atr.Name, opt.ID, opt.Name, opt.Value, opt.Unit)
		if err != nil {
			return nil, yeahapi.E(op, err)
		}

		existingAtr, ok := attributesMap[atr.ID]
		if !ok {
			attributesMap[atr.ID] = atr
		}

		existingAtr.Options = append(existingAtr.Options, opt)
	}

	return attributes, nil
}
