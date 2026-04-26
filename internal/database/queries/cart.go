package queries

import (
	"context"
	"time"

	"github.com/jakka/minimule-backend/graph/model"
)

func (q *Queries) GetActiveCartByUserID(ctx context.Context, userID string) (*model.Cart, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, user_id, status, updated_at
		FROM carts WHERE user_id = $1 AND status = 'active'
		LIMIT 1
	`, userID)
	return scanCart(row)
}

func (q *Queries) CreateCart(ctx context.Context, userID string) (*model.Cart, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO carts (user_id, status)
		VALUES ($1, 'active')
		RETURNING id, user_id, status, updated_at
	`, userID)
	return scanCart(row)
}

func (q *Queries) GetCartItems(ctx context.Context, cartID string) ([]*model.CartItem, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, cart_id, product_id, variant_id, quantity, unit_price
		FROM cart_items WHERE cart_id = $1
		ORDER BY id
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.CartItem
	for rows.Next() {
		var ci model.CartItem
		if err := rows.Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.VariantID, &ci.Quantity, &ci.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, &ci)
	}
	return items, rows.Err()
}

func (q *Queries) UpsertCartItem(ctx context.Context, cartID, productID string, variantID *string, quantity int, unitPrice float64) (*model.CartItem, error) {
	// Check for existing item first (avoids pgx type-inference issues with functional ON CONFLICT).
	var existingID string
	var matchSQL string
	var matchArgs []any
	if variantID == nil {
		matchSQL = `SELECT id FROM cart_items WHERE cart_id=$1 AND product_id=$2 AND variant_id IS NULL LIMIT 1`
		matchArgs = []any{cartID, productID}
	} else {
		matchSQL = `SELECT id FROM cart_items WHERE cart_id=$1 AND product_id=$2 AND variant_id=$3 LIMIT 1`
		matchArgs = []any{cartID, productID, *variantID}
	}
	_ = q.pool.QueryRow(ctx, matchSQL, matchArgs...).Scan(&existingID)

	var ci model.CartItem
	if existingID != "" {
		err := q.pool.QueryRow(ctx, `
			UPDATE cart_items SET quantity=quantity+$1, unit_price=$2 WHERE id=$3
			RETURNING id, cart_id, product_id, variant_id, quantity, unit_price
		`, quantity, unitPrice, existingID).Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.VariantID, &ci.Quantity, &ci.UnitPrice)
		return &ci, err
	}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO cart_items (cart_id, product_id, variant_id, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, cart_id, product_id, variant_id, quantity, unit_price
	`, cartID, productID, variantID, quantity, unitPrice).Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.VariantID, &ci.Quantity, &ci.UnitPrice)
	return &ci, err
}

func (q *Queries) UpdateCartItemQty(ctx context.Context, cartItemID string, quantity int) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE cart_items SET quantity = $1 WHERE id = $2
	`, quantity, cartItemID)
	return err
}

func (q *Queries) DeleteCartItem(ctx context.Context, cartItemID string) error {
	_, err := q.pool.Exec(ctx, `DELETE FROM cart_items WHERE id = $1`, cartItemID)
	return err
}

func (q *Queries) GetCartItemByID(ctx context.Context, id string) (*model.CartItem, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, cart_id, product_id, variant_id, quantity, unit_price
		FROM cart_items WHERE id = $1
	`, id)
	var ci model.CartItem
	err := row.Scan(&ci.ID, &ci.CartID, &ci.ProductID, &ci.VariantID, &ci.Quantity, &ci.UnitPrice)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return &ci, err
}

func (q *Queries) CheckoutCart(ctx context.Context, cartID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE carts SET status = 'checked_out' WHERE id = $1
	`, cartID)
	return err
}

func (q *Queries) TouchCart(ctx context.Context, cartID string) error {
	_, err := q.pool.Exec(ctx, `
		UPDATE carts SET updated_at = NOW() WHERE id = $1
	`, cartID)
	return err
}

func scanCart(row interface{ Scan(...any) error }) (*model.Cart, error) {
	var c model.Cart
	var userID *string
	var statusStr string
	var updatedAt time.Time
	err := row.Scan(&c.ID, &userID, &statusStr, &updatedAt)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if userID != nil {
		c.UserID = userID
	}
	c.Status = model.CartStatus(statusStr)
	c.UpdatedAt = updatedAt
	return &c, nil
}
