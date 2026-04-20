package model

import "time"

// ── Category ──────────────────────────

type Category struct {
	ID           string
	ParentID     *string
	Name         string
	Slug         *string
	IconURL      *string
	DisplayOrder int
	IsActive     bool

	CreatedAt time.Time
}

// ── Product ───────────────────────────

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"
	ProductStatusActive   ProductStatus = "active"
	ProductStatusArchived ProductStatus = "archived"
)

type Product struct {
	ID          string
	Name        string
	Description *string

	BasePrice float64
	StockQty  int

	IsCustomizable  bool
	IsFeatured      bool
	PopularityScore float64
	AvgRating       float64
	ReviewCount     int
	Status          ProductStatus

	CategoryID string
	SellerID   string

	CreatedAt time.Time
	UpdatedAt *time.Time

	Category *Category
	Images   []*ProductImage
	Variants []*ProductVariant
}

// ── Product Image ─────────────────────

type ProductImage struct {
	ID        string
	ProductID string
	ImageURL  string
	AltText   *string
	SortOrder int
	IsPrimary bool
}

// ── Variant ───────────────────────────

type ProductVariant struct {
	ID        string
	ProductID string
	SKU       *string
	Size      *string
	Color     *string
	Material  *string

	PriceModifier float64
	StockQty      int
}
