package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
)

// ── ProductResolver ───────────────────────────────────────────────────────────

type ProductResolver struct {
	p    *model.Product
	root *RootResolver
}

func (r *ProductResolver) ID() graphql.ID          { return graphql.ID(r.p.ID) }
func (r *ProductResolver) Name() string            { return r.p.Name }
func (r *ProductResolver) Description() *string    { return r.p.Description }
func (r *ProductResolver) BasePrice() float64      { return r.p.BasePrice }
func (r *ProductResolver) StockQty() int32         { return int32(r.p.StockQty) }
func (r *ProductResolver) Status() string          { return string(r.p.Status) }
func (r *ProductResolver) IsFeatured() bool        { return r.p.IsFeatured }
func (r *ProductResolver) AvgRating() float64      { return r.p.AvgRating }
func (r *ProductResolver) ReviewCount() int32      { return int32(r.p.ReviewCount) }
func (r *ProductResolver) CreatedAt() graphql.Time { return graphql.Time{Time: r.p.CreatedAt} }

func (r *ProductResolver) Category(ctx context.Context) (*CategoryResolver, error) {
	if r.p.Category != nil {
		return &CategoryResolver{c: r.p.Category}, nil
	}
	cat, err := r.root.ProductSvc.GetCategory(ctx, r.p.CategoryID)
	if err != nil {
		return nil, err
	}
	return &CategoryResolver{c: cat}, nil
}

func (r *ProductResolver) Images(ctx context.Context) (*[]*ProductImageResolver, error) {
	if r.p.Images != nil {
		resolvers := make([]*ProductImageResolver, len(r.p.Images))
		for i, img := range r.p.Images {
			resolvers[i] = &ProductImageResolver{img: img}
		}
		return &resolvers, nil
	}
	imgs, err := r.root.ProductSvc.GetProductImages(ctx, r.p.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*ProductImageResolver, len(imgs))
	for i, img := range imgs {
		resolvers[i] = &ProductImageResolver{img: img}
	}
	return &resolvers, nil
}

func (r *ProductResolver) Variants(ctx context.Context) (*[]*ProductVariantResolver, error) {
	if r.p.Variants != nil {
		resolvers := make([]*ProductVariantResolver, len(r.p.Variants))
		for i, v := range r.p.Variants {
			resolvers[i] = &ProductVariantResolver{v: v}
		}
		return &resolvers, nil
	}
	variants, err := r.root.ProductSvc.GetProductVariants(ctx, r.p.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*ProductVariantResolver, len(variants))
	for i, v := range variants {
		resolvers[i] = &ProductVariantResolver{v: v}
	}
	return &resolvers, nil
}

// ── CategoryResolver ──────────────────────────────────────────────────────────

type CategoryResolver struct {
	c *model.Category
}

func (r *CategoryResolver) ID() graphql.ID    { return graphql.ID(r.c.ID) }
func (r *CategoryResolver) Name() string      { return r.c.Name }
func (r *CategoryResolver) Slug() *string     { return r.c.Slug }
func (r *CategoryResolver) IconUrl() *string  { return r.c.IconURL }
func (r *CategoryResolver) IsActive() bool    { return r.c.IsActive }

// ── ProductImageResolver ──────────────────────────────────────────────────────

type ProductImageResolver struct {
	img *model.ProductImage
}

func (r *ProductImageResolver) ID() graphql.ID    { return graphql.ID(r.img.ID) }
func (r *ProductImageResolver) ImageUrl() string  { return r.img.ImageURL }
func (r *ProductImageResolver) IsPrimary() bool   { return r.img.IsPrimary }

// ── ProductVariantResolver ────────────────────────────────────────────────────

type ProductVariantResolver struct {
	v *model.ProductVariant
}

func (r *ProductVariantResolver) ID() graphql.ID          { return graphql.ID(r.v.ID) }
func (r *ProductVariantResolver) Sku() *string            { return r.v.SKU }
func (r *ProductVariantResolver) Size() *string           { return r.v.Size }
func (r *ProductVariantResolver) Color() *string          { return r.v.Color }
func (r *ProductVariantResolver) Material() *string       { return r.v.Material }
func (r *ProductVariantResolver) PriceModifier() float64  { return r.v.PriceModifier }
func (r *ProductVariantResolver) StockQty() int32         { return int32(r.v.StockQty) }
