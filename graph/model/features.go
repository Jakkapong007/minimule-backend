package model

import "time"

// ── Payment Methods ───────────────────────────────────────────────────────────

type PaymentMethod struct {
	ID        string
	UserID    string
	Type      string
	Label     string
	LastFour  *string
	Brand     *string
	Token     string
	IsDefault bool
	CreatedAt time.Time
}

type AddPaymentMethodInput struct {
	Type     string
	Label    *string
	LastFour *string
	Brand    *string
	Token    *string
}

// ── Shipping ──────────────────────────────────────────────────────────────────

type ShippingMethod struct {
	ID               string
	Name             string
	Description      *string
	Carrier          string
	EstimatedDaysMin int
	EstimatedDaysMax int
	BaseFee          float64
	IsActive         bool
}

type Shipment struct {
	ID               string
	OrderID          string
	ShippingMethodID *string
	TrackingNumber   *string
	Carrier          *string
	Status           string
	EstimatedDelivery *time.Time
	ShippedAt        *time.Time
	DeliveredAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ── Reviews ───────────────────────────────────────────────────────────────────

type ProductReview struct {
	ID        string
	ProductID string
	UserID    string
	OrderID   *string
	Rating    int
	Body      *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AddReviewInput struct {
	ProductID string
	Rating    int32
	Body      *string
	OrderID   *string
}

// ── Promotions ────────────────────────────────────────────────────────────────

type Promotion struct {
	ID               string
	Code             string
	Description      *string
	DiscountType     string
	DiscountValue    float64
	MinOrderAmount   float64
	MaxUses          *int
	UsedCount        int
	StartsAt         time.Time
	ExpiresAt        *time.Time
	IsActive         bool
	CreatedAt        time.Time
}

// ── Notifications ─────────────────────────────────────────────────────────────

type Notification struct {
	ID        string
	UserID    string
	Type      string
	Title     string
	Body      string
	IsRead    bool
	CreatedAt time.Time
}

// ── Search ────────────────────────────────────────────────────────────────────

type SearchHistory struct {
	ID        string
	UserID    string
	Query     string
	CreatedAt time.Time
}

// ── Extended inputs ───────────────────────────────────────────────────────────

type UpdateProfileInput struct {
	FullName          *string
	Phone             *string
	Bio               *string
	AvatarUrl         *string
	PreferredLanguage *string
}

type AddAddressInput struct {
	Label         *string
	RecipientName string
	Phone         string
	AddressLine1  string
	AddressLine2  *string
	Subdistrict   *string
	District      *string
	Province      *string
	PostalCode    *string
	IsDefault     *bool
}

type CreateOrderInput struct {
	AddressID        string
	ShippingMethodID *string
	PromotionCode    *string
}
