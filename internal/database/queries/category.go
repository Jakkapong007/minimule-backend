package queries

import (
	"context"

	"github.com/jakka/minimule-backend/graph/model"
)

func (q *Queries) GetCategoryByID(ctx context.Context, id string) (*model.Category, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, parent_id, name, slug, icon_url, display_order, is_active, created_at
		FROM categories WHERE id = $1
	`, id)
	return scanCategory(row)
}

func (q *Queries) ListCategories(ctx context.Context) ([]*model.Category, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, parent_id, name, slug, icon_url, display_order, is_active, created_at
		FROM categories
		WHERE is_active = true
		ORDER BY display_order, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []*model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(
			&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.IconURL, &c.DisplayOrder, &c.IsActive, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		cats = append(cats, &c)
	}
	return cats, rows.Err()
}

func scanCategory(row interface{ Scan(...any) error }) (*model.Category, error) {
	var c model.Category
	err := row.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.IconURL, &c.DisplayOrder, &c.IsActive, &c.CreatedAt)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return &c, err
}
