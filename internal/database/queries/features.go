package queries

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jakka/minimule-backend/graph/model"
)

// ── Payment Methods ───────────────────────────────────────────────────────────

func (q *Queries) GetPaymentMethods(ctx context.Context, userID string) ([]*model.PaymentMethod, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, type, label, last_four, brand, token, is_default, created_at
		FROM user_payment_methods WHERE user_id = $1 ORDER BY is_default DESC, created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.PaymentMethod
	for rows.Next() {
		m := &model.PaymentMethod{}
		if err := rows.Scan(&m.ID, &m.UserID, &m.Type, &m.Label, &m.LastFour, &m.Brand, &m.Token, &m.IsDefault, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (q *Queries) AddPaymentMethod(ctx context.Context, userID string, input model.AddPaymentMethodInput) (*model.PaymentMethod, error) {
	label := input.Type
	if input.Label != nil && *input.Label != "" {
		label = *input.Label
	} else if input.Brand != nil {
		label = *input.Brand
	}
	token := ""
	if input.Token != nil {
		token = *input.Token
	}
	m := &model.PaymentMethod{}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO user_payment_methods (user_id, type, label, last_four, brand, token)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, type, label, last_four, brand, token, is_default, created_at
	`, userID, input.Type, label, input.LastFour, input.Brand, token).
		Scan(&m.ID, &m.UserID, &m.Type, &m.Label, &m.LastFour, &m.Brand, &m.Token, &m.IsDefault, &m.CreatedAt)
	return m, err
}

func (q *Queries) RemovePaymentMethod(ctx context.Context, id, userID string) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM user_payment_methods WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (q *Queries) SetDefaultPaymentMethod(ctx context.Context, id, userID string) (*model.PaymentMethod, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE user_payment_methods SET is_default = FALSE WHERE user_id = $1`, userID); err != nil {
		return nil, err
	}
	m := &model.PaymentMethod{}
	if err := tx.QueryRow(ctx, `
		UPDATE user_payment_methods SET is_default = TRUE WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, type, label, last_four, brand, token, is_default, created_at
	`, id, userID).Scan(&m.ID, &m.UserID, &m.Type, &m.Label, &m.LastFour, &m.Brand, &m.Token, &m.IsDefault, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m, tx.Commit(ctx)
}

// ── Shipping ──────────────────────────────────────────────────────────────────

func (q *Queries) ListShippingMethods(ctx context.Context) ([]*model.ShippingMethod, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, name, description, carrier, estimated_days_min, estimated_days_max, base_fee
		FROM shipping_methods WHERE is_active = TRUE ORDER BY base_fee
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ShippingMethod
	for rows.Next() {
		s := &model.ShippingMethod{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Carrier, &s.EstimatedDaysMin, &s.EstimatedDaysMax, &s.BaseFee); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (q *Queries) GetShippingMethodByID(ctx context.Context, id string) (*model.ShippingMethod, error) {
	s := &model.ShippingMethod{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, name, description, carrier, estimated_days_min, estimated_days_max, base_fee
		FROM shipping_methods WHERE id = $1
	`, id).Scan(&s.ID, &s.Name, &s.Description, &s.Carrier, &s.EstimatedDaysMin, &s.EstimatedDaysMax, &s.BaseFee)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return s, err
}

func (q *Queries) GetShipmentByOrderID(ctx context.Context, orderID string) (*model.Shipment, error) {
	s := &model.Shipment{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, order_id, shipping_method_id, tracking_number, carrier, status,
		       estimated_delivery, shipped_at, delivered_at, created_at, updated_at
		FROM shipments WHERE order_id = $1 LIMIT 1
	`, orderID).Scan(&s.ID, &s.OrderID, &s.ShippingMethodID, &s.TrackingNumber, &s.Carrier,
		&s.Status, &s.EstimatedDelivery, &s.ShippedAt, &s.DeliveredAt, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return s, err
}

// ── Product Reviews ───────────────────────────────────────────────────────────

func (q *Queries) GetProductReviews(ctx context.Context, productID string, limit, offset int) ([]*model.ProductReview, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, product_id, user_id, order_id, rating, body, created_at, updated_at
		FROM product_reviews WHERE product_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, productID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.ProductReview
	for rows.Next() {
		r := &model.ProductReview{}
		if err := rows.Scan(&r.ID, &r.ProductID, &r.UserID, &r.OrderID, &r.Rating, &r.Body, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (q *Queries) AddReview(ctx context.Context, userID string, input model.AddReviewInput) (*model.ProductReview, error) {
	r := &model.ProductReview{}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO product_reviews (product_id, user_id, order_id, rating, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, product_id, user_id, order_id, rating, body, created_at, updated_at
	`, input.ProductID, userID, input.OrderID, input.Rating, input.Body).
		Scan(&r.ID, &r.ProductID, &r.UserID, &r.OrderID, &r.Rating, &r.Body, &r.CreatedAt, &r.UpdatedAt)
	if err != nil && strings.Contains(err.Error(), "unique") {
		return nil, ErrDuplicate
	}
	return r, err
}

func (q *Queries) DeleteReview(ctx context.Context, id, userID string) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM product_reviews WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Promotions ────────────────────────────────────────────────────────────────

func (q *Queries) GetPromotionByCode(ctx context.Context, code string) (*model.Promotion, error) {
	p := &model.Promotion{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, code, description, discount_type, discount_value, min_order_amount,
		       max_uses, used_count, starts_at, expires_at, is_active, created_at
		FROM promotions
		WHERE code = $1 AND is_active = TRUE
		  AND starts_at <= now()
		  AND (expires_at IS NULL OR expires_at > now())
	`, strings.ToUpper(code)).Scan(
		&p.ID, &p.Code, &p.Description, &p.DiscountType, &p.DiscountValue, &p.MinOrderAmount,
		&p.MaxUses, &p.UsedCount, &p.StartsAt, &p.ExpiresAt, &p.IsActive, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (q *Queries) IncrementPromotionUsed(ctx context.Context, id string) error {
	_, err := q.pool.Exec(ctx, `UPDATE promotions SET used_count = used_count + 1 WHERE id = $1`, id)
	return err
}

// ── Notifications ─────────────────────────────────────────────────────────────

func (q *Queries) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*model.Notification, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, type, title, body, is_read, created_at
		FROM notifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.Notification
	for rows.Next() {
		n := &model.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (q *Queries) GetUnreadNotificationCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := q.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`, userID).Scan(&count)
	return count, err
}

func (q *Queries) MarkNotificationRead(ctx context.Context, id, userID string) (*model.Notification, error) {
	n := &model.Notification{}
	err := q.pool.QueryRow(ctx, `
		UPDATE notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, type, title, body, is_read, created_at
	`, id, userID).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return n, err
}

func (q *Queries) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	_, err := q.pool.Exec(ctx, `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`, userID)
	return err
}

func (q *Queries) CreateNotification(ctx context.Context, userID, nType, title, body string) (*model.Notification, error) {
	n := &model.Notification{}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, type, title, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, type, title, body, is_read, created_at
	`, userID, nType, title, body).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.CreatedAt)
	return n, err
}

// ── Search ────────────────────────────────────────────────────────────────────

func (q *Queries) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*model.Product, int, error) {
	pattern := "%" + strings.ToLower(query) + "%"
	rows, err := q.pool.Query(ctx, `
		SELECT id, name, description, base_price, stock_qty, status, is_featured,
		       avg_rating, review_count, category_id, seller_id, created_at, updated_at
		FROM products
		WHERE status = 'active'
		  AND (LOWER(name) LIKE $1 OR LOWER(description) LIKE $1)
		ORDER BY popularity_score DESC
		LIMIT $2 OFFSET $3
	`, pattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p, err := scanProductRowFull(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	q.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM products WHERE status = 'active'
		AND (LOWER(name) LIKE $1 OR LOWER(description) LIKE $1)
	`, pattern).Scan(&total)

	return products, total, nil
}

func (q *Queries) GetSearchHistory(ctx context.Context, userID string) ([]string, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT query FROM (
		  SELECT DISTINCT ON (query) query, created_at
		  FROM search_history WHERE user_id = $1
		  ORDER BY query, created_at DESC
		) t ORDER BY created_at DESC LIMIT 10
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (q *Queries) ClearSearchHistory(ctx context.Context, userID string) error {
	_, err := q.pool.Exec(ctx, `DELETE FROM search_history WHERE user_id = $1`, userID)
	return err
}

func (q *Queries) AddSearchHistory(ctx context.Context, userID, query string) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO search_history (user_id, query) VALUES ($1, $2)
	`, userID, query)
	return err
}

// ── Post Votes ────────────────────────────────────────────────────────────────

func (q *Queries) VotePost(ctx context.Context, postID, userID string) error {
	_, err := q.pool.Exec(ctx, `INSERT INTO post_votes (post_id, user_id) VALUES ($1, $2)`, postID, userID)
	if err != nil && strings.Contains(err.Error(), "unique") {
		return ErrDuplicate
	}
	return err
}

func (q *Queries) UnvotePost(ctx context.Context, postID, userID string) error {
	_, err := q.pool.Exec(ctx, `DELETE FROM post_votes WHERE post_id = $1 AND user_id = $2`, postID, userID)
	return err
}

func (q *Queries) IsPostVotedByUser(ctx context.Context, postID, userID string) (bool, error) {
	var exists bool
	err := q.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM post_votes WHERE post_id = $1 AND user_id = $2)`, postID, userID).Scan(&exists)
	return exists, err
}

func (q *Queries) GetShowcase(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, caption, image_url, is_sticker_design,
		       COALESCE(visibility, 'public'), like_count, comment_count, vote_count, category_id, created_at, updated_at
		FROM posts
		WHERE visibility = 'public'
		ORDER BY vote_count DESC, like_count DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsFull(rows)
}

// ── Extended Product Queries ──────────────────────────────────────────────────

func (q *Queries) ListProductsFiltered(ctx context.Context, limit, offset int, categoryID *string, minPrice, maxPrice *float64, isFeatured *bool, sort string) ([]*model.Product, error) {
	conds := []string{"p.status = 'active'"}
	args := []interface{}{}
	n := 1

	if categoryID != nil {
		conds = append(conds, fmt.Sprintf("p.category_id = $%d", n))
		args = append(args, *categoryID)
		n++
	}
	if minPrice != nil {
		conds = append(conds, fmt.Sprintf("p.base_price >= $%d", n))
		args = append(args, *minPrice)
		n++
	}
	if maxPrice != nil {
		conds = append(conds, fmt.Sprintf("p.base_price <= $%d", n))
		args = append(args, *maxPrice)
		n++
	}
	if isFeatured != nil {
		conds = append(conds, fmt.Sprintf("p.is_featured = $%d", n))
		args = append(args, *isFeatured)
		n++
	}

	orderBy := "p.popularity_score DESC, p.created_at DESC"
	switch sort {
	case "newest":
		orderBy = "p.created_at DESC"
	case "price_asc":
		orderBy = "p.base_price ASC"
	case "price_desc":
		orderBy = "p.base_price DESC"
	case "rating":
		orderBy = "p.avg_rating DESC, p.review_count DESC"
	}

	args = append(args, limit, offset)
	sql := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.base_price, p.stock_qty, p.status, p.is_featured,
		       p.avg_rating, p.review_count, p.category_id, p.seller_id, p.created_at, p.updated_at
		FROM products p
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(conds, " AND "), orderBy, n, n+1)

	rows, err := q.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p, err := scanProductRowFull(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func scanProductRowFull(rows pgx.Rows) (*model.Product, error) {
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

// ── Address Management ────────────────────────────────────────────────────────

func (q *Queries) AddAddress(ctx context.Context, userID string, input model.AddAddressInput) (*model.UserAddress, error) {
	if input.IsDefault != nil && *input.IsDefault {
		q.pool.Exec(ctx, `UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1`, userID)
	}
	a := &model.UserAddress{}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO user_addresses (user_id, label, recipient_name, phone, address_line1, address_line2, subdistrict, district, province, postal_code, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, label, recipient_name, phone, address_line1, address_line2, subdistrict, district, province, postal_code, is_default
	`, userID, input.Label, input.RecipientName, input.Phone, input.AddressLine1, input.AddressLine2,
		input.Subdistrict, input.District, input.Province, input.PostalCode,
		input.IsDefault != nil && *input.IsDefault).
		Scan(&a.ID, &a.Label, &a.RecipientName, &a.Phone, &a.AddressLine1, &a.AddressLine2,
			&a.Subdistrict, &a.District, &a.Province, &a.PostalCode, &a.IsDefault)
	return a, err
}

func (q *Queries) RemoveAddress(ctx context.Context, id, userID string) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM user_addresses WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (q *Queries) SetDefaultAddress(ctx context.Context, id, userID string) (*model.UserAddress, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	tx.Exec(ctx, `UPDATE user_addresses SET is_default = FALSE WHERE user_id = $1`, userID)
	a := &model.UserAddress{}
	if err := tx.QueryRow(ctx, `
		UPDATE user_addresses SET is_default = TRUE WHERE id = $1 AND user_id = $2
		RETURNING id, label, recipient_name, phone, address_line1, address_line2, subdistrict, district, province, postal_code, is_default
	`, id, userID).Scan(&a.ID, &a.Label, &a.RecipientName, &a.Phone, &a.AddressLine1, &a.AddressLine2, &a.Subdistrict, &a.District, &a.Province, &a.PostalCode, &a.IsDefault); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, tx.Commit(ctx)
}

// ── Profile Update ────────────────────────────────────────────────────────────

func (q *Queries) UpdateUserProfile(ctx context.Context, userID string, input model.UpdateProfileInput) (*model.User, error) {
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if input.FullName != nil || input.Phone != nil || input.AvatarUrl != nil {
		sets := []string{}
		args := []interface{}{}
		n := 1
		if input.FullName != nil {
			sets = append(sets, fmt.Sprintf("full_name = $%d", n))
			args = append(args, *input.FullName)
			n++
		}
		if input.Phone != nil {
			sets = append(sets, fmt.Sprintf("phone = $%d", n))
			args = append(args, *input.Phone)
			n++
		}
		if input.AvatarUrl != nil {
			sets = append(sets, fmt.Sprintf("avatar_url = $%d", n))
			args = append(args, *input.AvatarUrl)
			n++
		}
		if len(sets) > 0 {
			args = append(args, userID)
			tx.Exec(ctx, fmt.Sprintf("UPDATE users SET %s, updated_at = now() WHERE id = $%d", strings.Join(sets, ", "), n), args...)
		}
	}

	if input.Bio != nil || input.PreferredLanguage != nil {
		sets := []string{}
		args := []interface{}{}
		n := 1
		if input.Bio != nil {
			sets = append(sets, fmt.Sprintf("bio = $%d", n))
			args = append(args, *input.Bio)
			n++
		}
		if input.PreferredLanguage != nil {
			sets = append(sets, fmt.Sprintf("preferred_language = $%d", n))
			args = append(args, *input.PreferredLanguage)
			n++
		}
		if len(sets) > 0 {
			args = append(args, userID)
			tx.Exec(ctx, fmt.Sprintf(`
				INSERT INTO user_profiles (user_id, %s)
				VALUES ($%d, %s)
				ON CONFLICT (user_id) DO UPDATE SET %s
			`, strings.Join(sets, ", "), n,
				strings.Join(func() []string {
					ph := make([]string, n-1)
					for i := range ph {
						ph[i] = fmt.Sprintf("$%d", i+1)
					}
					return ph
				}(), ", "),
				strings.Join(sets, ", ")), args...)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return q.GetUserByID(ctx, userID)
}

func (q *Queries) AcceptPDPA(ctx context.Context, userID, version string) (*model.User, error) {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO user_profiles (user_id, pdpa_consent, pdpa_consent_at, pdpa_version)
		VALUES ($2, TRUE, now(), $1)
		ON CONFLICT (user_id) DO UPDATE
		SET pdpa_consent = TRUE, pdpa_consent_at = now(), pdpa_version = $1
	`, version, userID)
	if err != nil {
		return nil, err
	}
	return q.GetUserByID(ctx, userID)
}
