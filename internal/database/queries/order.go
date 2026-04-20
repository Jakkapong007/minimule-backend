package queries

import (
	"context"
	"fmt"
	"time"

	"github.com/jakka/minimule-backend/graph/model"
)

func (q *Queries) CreateOrderFull(ctx context.Context, userID, addressID string, shippingMethodID, promoID, promoCode *string, items []*model.CartItem, subtotal, discountAmount, shippingFee, total float64) (*model.Order, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('order_number_seq')`).Scan(&seq); err != nil {
		return nil, err
	}
	orderNumber := fmt.Sprintf("MM-%08d", seq)

	var order model.Order
	var statusStr string
	var updatedAt *time.Time
	if err := tx.QueryRow(ctx, `
		INSERT INTO orders (order_number, user_id, address_id, shipping_method_id, promotion_id, promotion_code, subtotal, discount_amount, shipping_fee, total, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending')
		RETURNING id, order_number, user_id, address_id, subtotal, discount_amount, shipping_fee, total, status, created_at, updated_at
	`, orderNumber, userID, addressID, shippingMethodID, promoID, promoCode, subtotal, discountAmount, shippingFee, total).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID,
		&order.Subtotal, &order.DiscountAmount, &order.ShippingFee, &order.Total,
		&statusStr, &order.CreatedAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	order.Status = model.OrderStatus(statusStr)
	order.UpdatedAt = updatedAt

	for _, ci := range items {
		var oi model.OrderItem
		subtotalItem := float64(ci.Quantity) * ci.UnitPrice
		if err := tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, product_id, variant_id, quantity, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, order_id, product_id, variant_id, quantity, unit_price, subtotal
		`, order.ID, ci.ProductID, ci.VariantID, ci.Quantity, ci.UnitPrice, subtotalItem).Scan(
			&oi.ID, &oi.OrderID, &oi.ProductID, &oi.VariantID, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal,
		); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, &oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (q *Queries) CreateOrder(ctx context.Context, userID, addressID string, items []*model.CartItem, subtotal, total float64) (*model.Order, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Generate a human-readable order number
	var seq int64
	if err := tx.QueryRow(ctx, `SELECT nextval('order_number_seq')`).Scan(&seq); err != nil {
		return nil, err
	}
	orderNumber := fmt.Sprintf("MM-%08d", seq)

	var order model.Order
	var statusStr string
	var updatedAt *time.Time
	if err := tx.QueryRow(ctx, `
		INSERT INTO orders (order_number, user_id, address_id, subtotal, discount_amount, shipping_fee, total, status)
		VALUES ($1, $2, $3, $4, 0, 0, $5, 'pending')
		RETURNING id, order_number, user_id, address_id, subtotal, discount_amount, shipping_fee, total, status, created_at, updated_at
	`, orderNumber, userID, addressID, subtotal, total).Scan(
		&order.ID, &order.OrderNumber, &order.UserID, &order.AddressID,
		&order.Subtotal, &order.DiscountAmount, &order.ShippingFee, &order.Total,
		&statusStr, &order.CreatedAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	order.Status = model.OrderStatus(statusStr)
	order.UpdatedAt = updatedAt

	for _, ci := range items {
		var oi model.OrderItem
		subtotalItem := float64(ci.Quantity) * ci.UnitPrice
		if err := tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, product_id, variant_id, quantity, unit_price, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, order_id, product_id, variant_id, quantity, unit_price, subtotal
		`, order.ID, ci.ProductID, ci.VariantID, ci.Quantity, ci.UnitPrice, subtotalItem).Scan(
			&oi.ID, &oi.OrderID, &oi.ProductID, &oi.VariantID, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal,
		); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, &oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &order, nil
}

func (q *Queries) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, order_number, user_id, address_id, subtotal, discount_amount, shipping_fee, total, status, created_at, updated_at, promotion_code, shipping_method_id
		FROM orders WHERE id = $1
	`, id)
	o, err := scanOrder(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return o, err
}

func (q *Queries) ListUserOrders(ctx context.Context, userID string) ([]*model.Order, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, order_number, user_id, address_id, subtotal, discount_amount, shipping_fee, total, status, created_at, updated_at, promotion_code, shipping_method_id
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		o, err := scanOrderRow(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (q *Queries) GetOrderItems(ctx context.Context, orderID string) ([]*model.OrderItem, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, order_id, product_id, variant_id, quantity, unit_price, subtotal
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.OrderItem
	for rows.Next() {
		var oi model.OrderItem
		if err := rows.Scan(&oi.ID, &oi.OrderID, &oi.ProductID, &oi.VariantID, &oi.Quantity, &oi.UnitPrice, &oi.Subtotal); err != nil {
			return nil, err
		}
		items = append(items, &oi)
	}
	return items, rows.Err()
}

func scanOrder(row interface{ Scan(...any) error }) (*model.Order, error) {
	var o model.Order
	var statusStr string
	var updatedAt *time.Time
	err := row.Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.AddressID,
		&o.Subtotal, &o.DiscountAmount, &o.ShippingFee, &o.Total,
		&statusStr, &o.CreatedAt, &updatedAt, &o.PromotionCode, &o.ShippingMethodID,
	)
	if err != nil {
		return nil, err
	}
	o.Status = model.OrderStatus(statusStr)
	o.UpdatedAt = updatedAt
	return &o, nil
}

func scanOrderRow(row interface{ Scan(...any) error }) (*model.Order, error) {
	return scanOrder(row)
}
