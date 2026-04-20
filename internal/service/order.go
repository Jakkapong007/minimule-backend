package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

type OrderService struct {
	q    *queries.Queries
	cart *CartService
}

func NewOrderService(q *queries.Queries, cart *CartService) *OrderService {
	return &OrderService{q: q, cart: cart}
}

// CreateOrder converts the user's active cart into an order.
func (s *OrderService) CreateOrder(ctx context.Context, userID string, input model.CreateOrderInput) (*model.Order, error) {
	addr, err := s.q.GetUserAddress(ctx, input.AddressID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, fmt.Errorf("%w: address not found", ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	if addr.UserID != userID {
		return nil, ErrForbidden
	}

	cart, err := s.cart.GetCartForOrder(ctx, userID)
	if err != nil {
		return nil, err
	}

	var subtotal float64
	for _, item := range cart.Items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	var shippingFee float64
	if input.ShippingMethodID != nil {
		methods, _ := s.q.ListShippingMethods(ctx)
		for _, m := range methods {
			if m.ID == *input.ShippingMethodID {
				shippingFee = m.BaseFee
				break
			}
		}
	}

	discountAmount := 0.0
	var promoID *string
	if input.PromotionCode != nil && *input.PromotionCode != "" {
		promo, err := s.q.GetPromotionByCode(ctx, *input.PromotionCode)
		if err == nil {
			if promo.DiscountType == "percentage" {
				discountAmount = subtotal * promo.DiscountValue / 100
			} else {
				discountAmount = promo.DiscountValue
			}
			_ = s.q.IncrementPromotionUsed(ctx, promo.ID)
			promoID = &promo.ID
		}
	}

	total := subtotal + shippingFee - discountAmount
	if total < 0 {
		total = 0
	}

	order, err := s.q.CreateOrderFull(ctx, userID, input.AddressID, input.ShippingMethodID, promoID, input.PromotionCode, cart.Items, subtotal, discountAmount, shippingFee, total)
	if err != nil {
		return nil, err
	}

	if err := s.cart.CheckoutCart(ctx, cart.ID); err != nil {
		_ = err
	}

	return order, nil
}

// GetOrder fetches an order by ID, verifying it belongs to the user.
func (s *OrderService) GetOrder(ctx context.Context, id, userID string) (*model.Order, error) {
	order, err := s.q.GetOrderByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrForbidden
	}

	items, err := s.q.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.Items = items
	return order, nil
}

// ListUserOrders returns all orders for a user.
func (s *OrderService) ListUserOrders(ctx context.Context, userID string) ([]*model.Order, error) {
	return s.q.ListUserOrders(ctx, userID)
}

// GetOrderItems fetches order items (lazy-loaded by resolver).
func (s *OrderService) GetOrderItems(ctx context.Context, orderID string) ([]*model.OrderItem, error) {
	return s.q.GetOrderItems(ctx, orderID)
}
