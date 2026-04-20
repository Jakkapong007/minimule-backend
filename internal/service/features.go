package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

// ── Review Service ────────────────────────────────────────────────────────────

type ReviewService struct{ q *queries.Queries }

func NewReviewService(q *queries.Queries) *ReviewService { return &ReviewService{q: q} }

func (s *ReviewService) GetProductReviews(ctx context.Context, productID string, limit, offset int) ([]*model.ProductReview, error) {
	return s.q.GetProductReviews(ctx, productID, limit, offset)
}

func (s *ReviewService) AddReview(ctx context.Context, userID string, input model.AddReviewInput) (*model.ProductReview, error) {
	if input.Rating < 1 || input.Rating > 5 {
		return nil, fmt.Errorf("%w: rating must be 1–5", ErrBadRequest)
	}
	r, err := s.q.AddReview(ctx, userID, input)
	if errors.Is(err, queries.ErrDuplicate) {
		return nil, fmt.Errorf("%w: you already reviewed this product", ErrBadRequest)
	}
	return r, err
}

func (s *ReviewService) DeleteReview(ctx context.Context, id, userID string) error {
	err := s.q.DeleteReview(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

// ── Payment Method Service ────────────────────────────────────────────────────

type PaymentService struct{ q *queries.Queries }

func NewPaymentService(q *queries.Queries) *PaymentService { return &PaymentService{q: q} }

func (s *PaymentService) GetMethods(ctx context.Context, userID string) ([]*model.PaymentMethod, error) {
	return s.q.GetPaymentMethods(ctx, userID)
}

func (s *PaymentService) Add(ctx context.Context, userID string, input model.AddPaymentMethodInput) (*model.PaymentMethod, error) {
	return s.q.AddPaymentMethod(ctx, userID, input)
}

func (s *PaymentService) Remove(ctx context.Context, id, userID string) error {
	err := s.q.RemovePaymentMethod(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PaymentService) SetDefault(ctx context.Context, id, userID string) (*model.PaymentMethod, error) {
	m, err := s.q.SetDefaultPaymentMethod(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return m, err
}

// ── Shipping Service ──────────────────────────────────────────────────────────

type ShippingService struct{ q *queries.Queries }

func NewShippingService(q *queries.Queries) *ShippingService { return &ShippingService{q: q} }

func (s *ShippingService) ListMethods(ctx context.Context) ([]*model.ShippingMethod, error) {
	return s.q.ListShippingMethods(ctx)
}

func (s *ShippingService) GetShipmentByOrderID(ctx context.Context, orderID string) (*model.Shipment, error) {
	return s.q.GetShipmentByOrderID(ctx, orderID)
}

func (s *ShippingService) GetMethodByID(ctx context.Context, id string) (*model.ShippingMethod, error) {
	m, err := s.q.GetShippingMethodByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return m, err
}

// ── Promotion Service ─────────────────────────────────────────────────────────

type PromotionService struct{ q *queries.Queries }

func NewPromotionService(q *queries.Queries) *PromotionService { return &PromotionService{q: q} }

type PromoCheckResult struct {
	Valid          bool
	Message        string
	DiscountAmount float64
	Promotion      *model.Promotion
}

func (s *PromotionService) CheckCode(ctx context.Context, code string, orderTotal float64) *PromoCheckResult {
	promo, err := s.q.GetPromotionByCode(ctx, code)
	if errors.Is(err, queries.ErrNotFound) || err != nil {
		return &PromoCheckResult{Valid: false, Message: "Promotion code not found or expired"}
	}
	if promo.MaxUses != nil && promo.UsedCount >= *promo.MaxUses {
		return &PromoCheckResult{Valid: false, Message: "Promotion code has reached its usage limit"}
	}
	if orderTotal < promo.MinOrderAmount {
		return &PromoCheckResult{Valid: false, Message: fmt.Sprintf("Minimum order amount is %.2f", promo.MinOrderAmount)}
	}

	var discount float64
	if promo.DiscountType == "percentage" {
		discount = math.Round(orderTotal*promo.DiscountValue/100*100) / 100
	} else {
		discount = promo.DiscountValue
		if discount > orderTotal {
			discount = orderTotal
		}
	}
	return &PromoCheckResult{Valid: true, Message: "Promotion applied!", DiscountAmount: discount, Promotion: promo}
}

func (s *PromotionService) GetByCode(ctx context.Context, code string) (*model.Promotion, error) {
	p, err := s.q.GetPromotionByCode(ctx, code)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

// ── Notification Service ──────────────────────────────────────────────────────

type NotificationService struct{ q *queries.Queries }

func NewNotificationService(q *queries.Queries) *NotificationService {
	return &NotificationService{q: q}
}

type NotificationPage struct {
	Items       []*model.Notification
	UnreadCount int
}

func (s *NotificationService) GetPage(ctx context.Context, userID string, limit, offset int) (*NotificationPage, error) {
	items, err := s.q.GetNotifications(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	count, err := s.q.GetUnreadNotificationCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &NotificationPage{Items: items, UnreadCount: count}, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) (*model.Notification, error) {
	n, err := s.q.MarkNotificationRead(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return n, err
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.q.MarkAllNotificationsRead(ctx, userID)
}

// ── Search Service ────────────────────────────────────────────────────────────

type SearchService struct{ q *queries.Queries }

func NewSearchService(q *queries.Queries) *SearchService { return &SearchService{q: q} }

type SearchResult struct {
	Products []*model.Product
	Total    int
}

func (s *SearchService) SearchProducts(ctx context.Context, userID, query string, limit, offset int) (*SearchResult, error) {
	if userID != "" {
		_ = s.q.AddSearchHistory(ctx, userID, query)
	}
	products, total, err := s.q.SearchProducts(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return &SearchResult{Products: products, Total: total}, nil
}

func (s *SearchService) GetHistory(ctx context.Context, userID string) ([]string, error) {
	return s.q.GetSearchHistory(ctx, userID)
}

func (s *SearchService) ClearHistory(ctx context.Context, userID string) error {
	return s.q.ClearSearchHistory(ctx, userID)
}

// ── Profile/Address extensions on AuthService ─────────────────────────────────

func (s *AuthService) UpdateProfile(ctx context.Context, userID string, input model.UpdateProfileInput) (*model.User, error) {
	return s.q.UpdateUserProfile(ctx, userID, input)
}

func (s *AuthService) AcceptPDPA(ctx context.Context, userID, version string) (*model.User, error) {
	return s.q.AcceptPDPA(ctx, userID, version)
}

func (s *AuthService) AddAddress(ctx context.Context, userID string, input model.AddAddressInput) (*model.UserAddress, error) {
	return s.q.AddAddress(ctx, userID, input)
}

func (s *AuthService) RemoveAddress(ctx context.Context, id, userID string) error {
	err := s.q.RemoveAddress(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *AuthService) SetDefaultAddress(ctx context.Context, id, userID string) (*model.UserAddress, error) {
	a, err := s.q.SetDefaultAddress(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return a, err
}
