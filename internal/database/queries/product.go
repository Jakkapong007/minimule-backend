package queries

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/jakka/minimule-backend/graph/model"
)

// ── Products ──────────────────────────────────────────────────────────────────

func (q *Queries) CreateProduct(ctx context.Context, input model.CreateProductInput, sellerID string) (*model.Product, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO products (name, description, base_price, stock_qty, category_id, seller_id, status, is_featured)
		VALUES ($1, $2, $3, $4, $5, $6, 'draft', false)
		RETURNING id, name, description, base_price, stock_qty, status, is_featured,
		          COALESCE(avg_rating,0), COALESCE(review_count,0), category_id, seller_id, created_at, updated_at
	`, input.Name, input.Description, input.BasePrice, input.StockQty, input.CategoryID, sellerID)
	return scanProduct(row)
}

func (q *Queries) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, name, description, base_price, stock_qty, status, is_featured,
		       COALESCE(avg_rating,0), COALESCE(review_count,0), category_id, seller_id, created_at, updated_at
		FROM products WHERE id = $1
	`, id)
	p, err := scanProduct(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return p, err
}

func (q *Queries) ListProducts(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, name, description, base_price, stock_qty, status, is_featured,
		       COALESCE(avg_rating,0), COALESCE(review_count,0), category_id, seller_id, created_at, updated_at
		FROM products
		WHERE status = 'active'
		ORDER BY popularity_score DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p, err := scanProductRow(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func scanProduct(row pgx.Row) (*model.Product, error) {
	var p model.Product
	var statusStr string
	err := row.Scan(
		&p.ID, &p.Name, &p.Description, &p.BasePrice, &p.StockQty,
		&statusStr, &p.IsFeatured, &p.AvgRating, &p.ReviewCount,
		&p.CategoryID, &p.SellerID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Status = model.ProductStatus(statusStr)
	return &p, nil
}

func scanProductRow(rows pgx.Rows) (*model.Product, error) {
	var p model.Product
	var statusStr string
	err := rows.Scan(
		&p.ID, &p.Name, &p.Description, &p.BasePrice, &p.StockQty,
		&statusStr, &p.IsFeatured, &p.AvgRating, &p.ReviewCount,
		&p.CategoryID, &p.SellerID, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Status = model.ProductStatus(statusStr)
	return &p, nil
}

// ── Product Images ────────────────────────────────────────────────────────────

func (q *Queries) GetProductImages(ctx context.Context, productID string) ([]*model.ProductImage, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, product_id, image_url, is_primary
		FROM product_images
		WHERE product_id = $1
		ORDER BY is_primary DESC, sort_order
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imgs []*model.ProductImage
	for rows.Next() {
		var img model.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.IsPrimary); err != nil {
			return nil, err
		}
		imgs = append(imgs, &img)
	}
	return imgs, rows.Err()
}

// ── Product Variants ──────────────────────────────────────────────────────────

func (q *Queries) GetProductVariants(ctx context.Context, productID string) ([]*model.ProductVariant, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, product_id, sku, size, color, material, price_modifier, stock_qty
		FROM product_variants
		WHERE product_id = $1
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []*model.ProductVariant
	for rows.Next() {
		var v model.ProductVariant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Size, &v.Color, &v.Material, &v.PriceModifier, &v.StockQty); err != nil {
			return nil, err
		}
		variants = append(variants, &v)
	}
	return variants, rows.Err()
}
