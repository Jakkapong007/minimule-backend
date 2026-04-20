package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

type CartService struct {
	q *queries.Queries
}

func NewCartService(q *queries.Queries) *CartService {
	return &CartService{q: q}
}

// GetOrCreateCart returns the user's active cart, creating one if needed.
func (s *CartService) GetOrCreateCart(ctx context.Context, userID string) (*model.Cart, error) {
	cart, err := s.q.GetActiveCartByUserID(ctx, userID)
	if errors.Is(err, queries.ErrNotFound) {
		cart, err = s.q.CreateCart(ctx, userID)
	}
	if err != nil {
		return nil, err
	}

	items, err := s.q.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items
	return cart, nil
}

// AddToCart adds a product to the cart (or increments quantity if already present).
func (s *CartService) AddToCart(ctx context.Context, userID, productID string, quantity int) (*model.Cart, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be positive", ErrBadRequest)
	}

	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Unit price comes from the product — look it up via cart items query
	// (simplified: caller provides unit price via product service integration)
	_, err = s.q.UpsertCartItem(ctx, cart.ID, productID, nil, quantity, 0)
	if err != nil {
		return nil, err
	}
	_ = s.q.TouchCart(ctx, cart.ID)

	return s.GetOrCreateCart(ctx, userID)
}

// AddToCartWithPrice adds a product with a known price (used by resolver which fetches the product first).
func (s *CartService) AddToCartWithPrice(ctx context.Context, userID, productID string, quantity int, unitPrice float64) (*model.Cart, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("%w: quantity must be positive", ErrBadRequest)
	}

	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	_, err = s.q.UpsertCartItem(ctx, cart.ID, productID, nil, quantity, unitPrice)
	if err != nil {
		return nil, err
	}
	_ = s.q.TouchCart(ctx, cart.ID)

	return s.GetOrCreateCart(ctx, userID)
}

// UpdateCartItem changes quantity for an existing cart item. Quantity 0 removes the item.
func (s *CartService) UpdateCartItem(ctx context.Context, userID, cartItemID string, quantity int) (*model.Cart, error) {
	item, err := s.q.GetCartItemByID(ctx, cartItemID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify item belongs to user's cart
	cart, err := s.q.GetActiveCartByUserID(ctx, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if item.CartID != cart.ID {
		return nil, ErrForbidden
	}

	if quantity <= 0 {
		err = s.q.DeleteCartItem(ctx, cartItemID)
	} else {
		err = s.q.UpdateCartItemQty(ctx, cartItemID, quantity)
	}
	if err != nil {
		return nil, err
	}
	_ = s.q.TouchCart(ctx, cart.ID)

	return s.GetOrCreateCart(ctx, userID)
}

// GetCartForOrder returns the active cart with items for order creation.
func (s *CartService) GetCartForOrder(ctx context.Context, userID string) (*model.Cart, error) {
	cart, err := s.q.GetActiveCartByUserID(ctx, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, fmt.Errorf("%w: cart is empty", ErrBadRequest)
	}
	if err != nil {
		return nil, err
	}
	items, err := s.q.GetCartItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: cart is empty", ErrBadRequest)
	}
	cart.Items = items
	return cart, nil
}

// CheckoutCart marks the cart as checked out after order creation.
func (s *CartService) CheckoutCart(ctx context.Context, cartID string) error {
	return s.q.CheckoutCart(ctx, cartID)
}
