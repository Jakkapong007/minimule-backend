package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
)

// ── CartResolver ──────────────────────────────────────────────────────────────

type CartResolver struct {
	cart *model.Cart
	root *RootResolver
}

func (r *CartResolver) ID() graphql.ID          { return graphql.ID(r.cart.ID) }
func (r *CartResolver) Status() string          { return string(r.cart.Status) }
func (r *CartResolver) UpdatedAt() graphql.Time { return graphql.Time{Time: r.cart.UpdatedAt} }
func (r *CartResolver) Subtotal() float64 {
	var total float64
	for _, item := range r.cart.Items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total
}

func (r *CartResolver) Items(ctx context.Context) ([]*CartItemResolver, error) {
	resolvers := make([]*CartItemResolver, len(r.cart.Items))
	for i, item := range r.cart.Items {
		resolvers[i] = &CartItemResolver{item: item, root: r.root}
	}
	return resolvers, nil
}

// ── CartItemResolver ──────────────────────────────────────────────────────────

type CartItemResolver struct {
	item *model.CartItem
	root *RootResolver
}

func (r *CartItemResolver) ID() graphql.ID       { return graphql.ID(r.item.ID) }
func (r *CartItemResolver) Quantity() int32      { return int32(r.item.Quantity) }
func (r *CartItemResolver) UnitPrice() float64   { return r.item.UnitPrice }

func (r *CartItemResolver) Product(ctx context.Context) (*ProductResolver, error) {
	if r.item.Product != nil {
		return &ProductResolver{p: r.item.Product, root: r.root}, nil
	}
	p, err := r.root.ProductSvc.GetProduct(ctx, r.item.ProductID)
	if err != nil {
		return nil, err
	}
	return &ProductResolver{p: p, root: r.root}, nil
}
