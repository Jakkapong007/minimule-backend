package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/service"
)

// ── OrderResolver ─────────────────────────────────────────────────────────────

type OrderResolver struct {
	o    *model.Order
	root *RootResolver
}

func (r *OrderResolver) ID() graphql.ID             { return graphql.ID(r.o.ID) }
func (r *OrderResolver) OrderNumber() string        { return r.o.OrderNumber }
func (r *OrderResolver) Status() string             { return string(r.o.Status) }
func (r *OrderResolver) Subtotal() float64          { return r.o.Subtotal }
func (r *OrderResolver) DiscountAmount() float64    { return r.o.DiscountAmount }
func (r *OrderResolver) ShippingFee() float64       { return r.o.ShippingFee }
func (r *OrderResolver) Total() float64             { return r.o.Total }
func (r *OrderResolver) PromotionCode() *string     { return r.o.PromotionCode }
func (r *OrderResolver) CreatedAt() graphql.Time    { return graphql.Time{Time: r.o.CreatedAt} }

func (r *OrderResolver) ShippingMethod(ctx context.Context) (*ShippingMethodResolver, error) {
	if r.o.ShippingMethodID == nil {
		return nil, nil
	}
	m, err := r.root.ShippingSvc.GetMethodByID(ctx, *r.o.ShippingMethodID)
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ShippingMethodResolver{s: m}, nil
}

func (r *OrderResolver) Shipment(ctx context.Context) (*ShipmentResolver, error) {
	s, err := r.root.ShippingSvc.GetShipmentByOrderID(ctx, r.o.ID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, nil
	}
	return &ShipmentResolver{s: s}, nil
}

func (r *OrderResolver) Items(ctx context.Context) ([]*OrderItemResolver, error) {
	if r.o.Items != nil {
		resolvers := make([]*OrderItemResolver, len(r.o.Items))
		for i, item := range r.o.Items {
			resolvers[i] = &OrderItemResolver{item: item, root: r.root}
		}
		return resolvers, nil
	}
	items, err := r.root.OrderSvc.GetOrderItems(ctx, r.o.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*OrderItemResolver, len(items))
	for i, item := range items {
		resolvers[i] = &OrderItemResolver{item: item, root: r.root}
	}
	return resolvers, nil
}

// ── OrderItemResolver ─────────────────────────────────────────────────────────

type OrderItemResolver struct {
	item *model.OrderItem
	root *RootResolver
}

func (r *OrderItemResolver) ID() graphql.ID      { return graphql.ID(r.item.ID) }
func (r *OrderItemResolver) Quantity() int32     { return int32(r.item.Quantity) }
func (r *OrderItemResolver) UnitPrice() float64  { return r.item.UnitPrice }

func (r *OrderItemResolver) Product(ctx context.Context) (*ProductResolver, error) {
	if r.item.Product != nil {
		return &ProductResolver{p: r.item.Product, root: r.root}, nil
	}
	p, err := r.root.ProductSvc.GetProduct(ctx, r.item.ProductID)
	if err != nil {
		// Return nil product with error — preserve order data
		return nil, err
	}
	return &ProductResolver{p: p, root: r.root}, nil
}

