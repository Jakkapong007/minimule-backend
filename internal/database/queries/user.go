package queries

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jakka/minimule-backend/graph/model"
)

// ── Users ─────────────────────────────────────────────────────────────────────

func (q *Queries) CreateUser(
	ctx context.Context,
	email, passwordHash string,
	fullName, phone *string,
	role model.UserRole,
) (*model.User, error) {
	row := q.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, phone, role, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, email, password_hash, phone, full_name, avatar_url, role, is_active, created_at, updated_at
	`, email, passwordHash, fullName, phone, string(role))
	u, err := scanUser(row)
	if err != nil && strings.Contains(err.Error(), "unique") {
		return nil, ErrDuplicate
	}
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, phone, full_name, avatar_url, role, is_active, created_at, updated_at
		FROM users WHERE email = $1
	`, email)
	u, err := scanUser(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return u, err
}

func (q *Queries) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, phone, full_name, avatar_url, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id)
	u, err := scanUser(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return u, err
}

func scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	var roleStr string
	var updatedAt *time.Time
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Phone, &u.FullName,
		&u.AvatarURL, &roleStr, &u.IsActive, &u.CreatedAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.Role = model.UserRole(roleStr)
	u.UpdatedAt = updatedAt
	return &u, nil
}

// ── User Profiles ─────────────────────────────────────────────────────────────

func (q *Queries) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, user_id, bio, preferred_language,
		       COALESCE(pdpa_consent, FALSE), pdpa_consent_at, pdpa_version
		FROM user_profiles WHERE user_id = $1
	`, userID)
	var p model.UserProfile
	err := row.Scan(&p.ID, &p.UserID, &p.Bio, &p.PreferredLanguage,
		&p.PDPAConsent, &p.PDPAConsentAt, &p.PDPAVersion)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return &p, err
}

// ── User Addresses ────────────────────────────────────────────────────────────

func (q *Queries) GetUserAddresses(ctx context.Context, userID string) ([]*model.UserAddress, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, label, recipient_name, phone,
		       address_line1, address_line2, subdistrict, district, province, postal_code, is_default
		FROM user_addresses WHERE user_id = $1
		ORDER BY is_default DESC, id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []*model.UserAddress
	for rows.Next() {
		var a model.UserAddress
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
			&a.AddressLine1, &a.AddressLine2, &a.Subdistrict, &a.District, &a.Province, &a.PostalCode, &a.IsDefault,
		); err != nil {
			return nil, err
		}
		addrs = append(addrs, &a)
	}
	return addrs, rows.Err()
}

func (q *Queries) GetUserAddress(ctx context.Context, id string) (*model.UserAddress, error) {
	row := q.pool.QueryRow(ctx, `
		SELECT id, user_id, label, recipient_name, phone,
		       address_line1, address_line2, subdistrict, district, province, postal_code, is_default
		FROM user_addresses WHERE id = $1
	`, id)
	var a model.UserAddress
	err := row.Scan(
		&a.ID, &a.UserID, &a.Label, &a.RecipientName, &a.Phone,
		&a.AddressLine1, &a.AddressLine2, &a.Subdistrict, &a.District, &a.Province, &a.PostalCode, &a.IsDefault,
	)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return &a, err
}
