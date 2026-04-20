package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/service"
)

// ── PaymentMethod resolver ────────────────────────────────────────────────────

type PaymentMethodResolver struct{ m *model.PaymentMethod }

func (r *PaymentMethodResolver) ID() graphql.ID    { return graphql.ID(r.m.ID) }
func (r *PaymentMethodResolver) Type() string      { return r.m.Type }
func (r *PaymentMethodResolver) Label() string     { return r.m.Label }
func (r *PaymentMethodResolver) LastFour() *string { return r.m.LastFour }
func (r *PaymentMethodResolver) Brand() *string    { return r.m.Brand }
func (r *PaymentMethodResolver) IsDefault() bool   { return r.m.IsDefault }

// ── ShippingMethod resolver ───────────────────────────────────────────────────

type ShippingMethodResolver struct{ s *model.ShippingMethod }

func (r *ShippingMethodResolver) ID() graphql.ID        { return graphql.ID(r.s.ID) }
func (r *ShippingMethodResolver) Name() string          { return r.s.Name }
func (r *ShippingMethodResolver) Description() *string  { return r.s.Description }
func (r *ShippingMethodResolver) Carrier() string       { return r.s.Carrier }
func (r *ShippingMethodResolver) EstimatedDaysMin() int32 {
	return int32(r.s.EstimatedDaysMin)
}
func (r *ShippingMethodResolver) EstimatedDaysMax() int32 {
	return int32(r.s.EstimatedDaysMax)
}
func (r *ShippingMethodResolver) BaseFee() float64 { return r.s.BaseFee }

// ── Shipment resolver ─────────────────────────────────────────────────────────

type ShipmentResolver struct{ s *model.Shipment }

func (r *ShipmentResolver) ID() graphql.ID              { return graphql.ID(r.s.ID) }
func (r *ShipmentResolver) Status() string              { return r.s.Status }
func (r *ShipmentResolver) TrackingNumber() *string     { return r.s.TrackingNumber }
func (r *ShipmentResolver) Carrier() *string            { return r.s.Carrier }
func (r *ShipmentResolver) EstimatedDelivery() *graphql.Time {
	if r.s.EstimatedDelivery == nil {
		return nil
	}
	t := graphql.Time{Time: *r.s.EstimatedDelivery}
	return &t
}
func (r *ShipmentResolver) ShippedAt() *graphql.Time {
	if r.s.ShippedAt == nil {
		return nil
	}
	t := graphql.Time{Time: *r.s.ShippedAt}
	return &t
}
func (r *ShipmentResolver) DeliveredAt() *graphql.Time {
	if r.s.DeliveredAt == nil {
		return nil
	}
	t := graphql.Time{Time: *r.s.DeliveredAt}
	return &t
}

// ── ProductReview resolver ────────────────────────────────────────────────────

type ProductReviewResolver struct {
	r    *model.ProductReview
	root *RootResolver
}

func (r *ProductReviewResolver) ID() graphql.ID { return graphql.ID(r.r.ID) }
func (r *ProductReviewResolver) Product(ctx context.Context) (*ProductResolver, error) {
	p, err := r.root.ProductSvc.GetProduct(ctx, r.r.ProductID)
	if err != nil {
		return nil, err
	}
	return &ProductResolver{p: p, root: r.root}, nil
}
func (r *ProductReviewResolver) User(ctx context.Context) (*UserResolver, error) {
	u, err := r.root.Auth.GetUser(ctx, r.r.UserID)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: u, root: r.root}, nil
}
func (r *ProductReviewResolver) Rating() int32  { return int32(r.r.Rating) }
func (r *ProductReviewResolver) Body() *string  { return r.r.Body }
func (r *ProductReviewResolver) CreatedAt() graphql.Time { return graphql.Time{Time: r.r.CreatedAt} }

// ── Promotion resolver ────────────────────────────────────────────────────────

type PromotionResolver struct{ p *model.Promotion }

func (r *PromotionResolver) ID() graphql.ID          { return graphql.ID(r.p.ID) }
func (r *PromotionResolver) Code() string            { return r.p.Code }
func (r *PromotionResolver) Description() *string    { return r.p.Description }
func (r *PromotionResolver) DiscountType() string    { return r.p.DiscountType }
func (r *PromotionResolver) DiscountValue() float64  { return r.p.DiscountValue }
func (r *PromotionResolver) MinOrderAmount() float64 { return r.p.MinOrderAmount }
func (r *PromotionResolver) ExpiresAt() *graphql.Time {
	if r.p.ExpiresAt == nil {
		return nil
	}
	t := graphql.Time{Time: *r.p.ExpiresAt}
	return &t
}

type PromoCheckResultResolver struct {
	valid          bool
	message        string
	discountAmount float64
	promotion      *model.Promotion
}

func (r *PromoCheckResultResolver) Valid() bool           { return r.valid }
func (r *PromoCheckResultResolver) Message() string       { return r.message }
func (r *PromoCheckResultResolver) DiscountAmount() float64 { return r.discountAmount }
func (r *PromoCheckResultResolver) Promotion() *PromotionResolver {
	if r.promotion == nil {
		return nil
	}
	return &PromotionResolver{p: r.promotion}
}

// ── Notification resolver ─────────────────────────────────────────────────────

type NotificationResolver struct{ n *model.Notification }

func (r *NotificationResolver) ID() graphql.ID     { return graphql.ID(r.n.ID) }
func (r *NotificationResolver) Type() string       { return r.n.Type }
func (r *NotificationResolver) Title() string      { return r.n.Title }
func (r *NotificationResolver) Body() string       { return r.n.Body }
func (r *NotificationResolver) IsRead() bool       { return r.n.IsRead }
func (r *NotificationResolver) CreatedAt() graphql.Time { return graphql.Time{Time: r.n.CreatedAt} }

type NotificationPageResolver struct {
	items       []*model.Notification
	unreadCount int
}

func (r *NotificationPageResolver) Items() []*NotificationResolver {
	out := make([]*NotificationResolver, len(r.items))
	for i, n := range r.items {
		out[i] = &NotificationResolver{n: n}
	}
	return out
}
func (r *NotificationPageResolver) UnreadCount() int32 { return int32(r.unreadCount) }

// ── SearchResult resolver ─────────────────────────────────────────────────────

type SearchResultResolver struct {
	products []*model.Product
	total    int
	root     *RootResolver
}

func (r *SearchResultResolver) Products() []*ProductResolver {
	out := make([]*ProductResolver, len(r.products))
	for i, p := range r.products {
		out[i] = &ProductResolver{p: p, root: r.root}
	}
	return out
}
func (r *SearchResultResolver) Total() int32 { return int32(r.total) }

// ── Root queries ──────────────────────────────────────────────────────────────

func (r *RootResolver) MyPaymentMethods(ctx context.Context) ([]*PaymentMethodResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	methods, err := r.PaymentSvc.GetMethods(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	out := make([]*PaymentMethodResolver, len(methods))
	for i, m := range methods {
		out[i] = &PaymentMethodResolver{m: m}
	}
	return out, nil
}

func (r *RootResolver) ShippingMethods(ctx context.Context) ([]*ShippingMethodResolver, error) {
	methods, err := r.ShippingSvc.ListMethods(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*ShippingMethodResolver, len(methods))
	for i, s := range methods {
		out[i] = &ShippingMethodResolver{s: s}
	}
	return out, nil
}

func (r *RootResolver) ProductReviews(ctx context.Context, args struct {
	ProductID graphql.ID
	Limit     *int32
	Offset    *int32
}) ([]*ProductReviewResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	reviews, err := r.ReviewSvc.GetProductReviews(ctx, string(args.ProductID), limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*ProductReviewResolver, len(reviews))
	for i, rv := range reviews {
		out[i] = &ProductReviewResolver{r: rv, root: r}
	}
	return out, nil
}

func (r *RootResolver) CheckPromoCode(ctx context.Context, args struct {
	Code       string
	OrderTotal float64
}) (*PromoCheckResultResolver, error) {
	res := r.PromotionSvc.CheckCode(ctx, args.Code, args.OrderTotal)
	return &PromoCheckResultResolver{
		valid:          res.Valid,
		message:        res.Message,
		discountAmount: res.DiscountAmount,
		promotion:      res.Promotion,
	}, nil
}

func (r *RootResolver) MyNotifications(ctx context.Context, args struct {
	Limit  *int32
	Offset *int32
}) (*NotificationPageResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	page, err := r.NotifSvc.GetPage(ctx, c.UserID, limit, offset)
	if err != nil {
		return nil, err
	}
	return &NotificationPageResolver{items: page.Items, unreadCount: page.UnreadCount}, nil
}

func (r *RootResolver) SearchProducts(ctx context.Context, args struct {
	Query  string
	Limit  *int32
	Offset *int32
}) (*SearchResultResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	userID := ""
	if c, err := requireClaims(ctx); err == nil {
		userID = c.UserID
	}
	res, err := r.SearchSvc.SearchProducts(ctx, userID, args.Query, limit, offset)
	if err != nil {
		return nil, err
	}
	return &SearchResultResolver{products: res.Products, total: res.Total, root: r}, nil
}

func (r *RootResolver) MySearchHistory(ctx context.Context) ([]string, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	return r.SearchSvc.GetHistory(ctx, c.UserID)
}

func (r *RootResolver) Showcase(ctx context.Context, args struct {
	Limit  *int32
	Offset *int32
}) ([]*PostResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	posts, err := r.PostSvc.GetShowcase(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return wrapPosts(posts, r), nil
}

// ── Root mutations ────────────────────────────────────────────────────────────

func (r *RootResolver) UpdateProfile(ctx context.Context, args struct {
	Input model.UpdateProfileInput
}) (*UserResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	u, err := r.Auth.UpdateProfile(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: u, root: r}, nil
}

func (r *RootResolver) AcceptPDPA(ctx context.Context, args struct {
	Version string
}) (*UserResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	u, err := r.Auth.AcceptPDPA(ctx, c.UserID, args.Version)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: u, root: r}, nil
}

func (r *RootResolver) AddAddress(ctx context.Context, args struct {
	Input model.AddAddressInput
}) (*UserAddressResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	a, err := r.Auth.AddAddress(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &UserAddressResolver{a: a}, nil
}

func (r *RootResolver) RemoveAddress(ctx context.Context, args struct{ ID graphql.ID }) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	if err := r.Auth.RemoveAddress(ctx, string(args.ID), c.UserID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RootResolver) SetDefaultAddress(ctx context.Context, args struct{ ID graphql.ID }) (*UserAddressResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	a, err := r.Auth.SetDefaultAddress(ctx, string(args.ID), c.UserID)
	if err != nil {
		return nil, err
	}
	return &UserAddressResolver{a: a}, nil
}

func (r *RootResolver) UpdateAddress(ctx context.Context, args struct {
	ID    graphql.ID
	Input model.AddAddressInput
}) (*UserAddressResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	_ = c
	// For now delegate to add (upsert-style); a real impl would UPDATE WHERE id=$1 AND user_id=$2
	a, err := r.Auth.AddAddress(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &UserAddressResolver{a: a}, nil
}

func (r *RootResolver) AddPaymentMethod(ctx context.Context, args struct {
	Input model.AddPaymentMethodInput
}) (*PaymentMethodResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	m, err := r.PaymentSvc.Add(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &PaymentMethodResolver{m: m}, nil
}

func (r *RootResolver) RemovePaymentMethod(ctx context.Context, args struct{ ID graphql.ID }) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	if err := r.PaymentSvc.Remove(ctx, string(args.ID), c.UserID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RootResolver) SetDefaultPaymentMethod(ctx context.Context, args struct{ ID graphql.ID }) (*PaymentMethodResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	m, err := r.PaymentSvc.SetDefault(ctx, string(args.ID), c.UserID)
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &PaymentMethodResolver{m: m}, nil
}

func (r *RootResolver) AddReview(ctx context.Context, args struct {
	Input model.AddReviewInput
}) (*ProductReviewResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	rv, err := r.ReviewSvc.AddReview(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &ProductReviewResolver{r: rv, root: r}, nil
}

func (r *RootResolver) DeleteReview(ctx context.Context, args struct{ ID graphql.ID }) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	if err := r.ReviewSvc.DeleteReview(ctx, string(args.ID), c.UserID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RootResolver) RemoveFromCart(ctx context.Context, args struct{ CartItemId graphql.ID }) (*CartResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	cart, err := r.CartSvc.UpdateCartItem(ctx, c.UserID, string(args.CartItemId), 0)
	if err != nil {
		return nil, err
	}
	return &CartResolver{cart: cart, root: r}, nil
}

func (r *RootResolver) MarkNotificationRead(ctx context.Context, args struct{ ID graphql.ID }) (*NotificationResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	n, err := r.NotifSvc.MarkRead(ctx, string(args.ID), c.UserID)
	if err != nil {
		return nil, err
	}
	return &NotificationResolver{n: n}, nil
}

func (r *RootResolver) MarkAllNotificationsRead(ctx context.Context) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	return true, r.NotifSvc.MarkAllRead(ctx, c.UserID)
}

func (r *RootResolver) ClearSearchHistory(ctx context.Context) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	return true, r.SearchSvc.ClearHistory(ctx, c.UserID)
}

func (r *RootResolver) VotePost(ctx context.Context, args struct{ PostId graphql.ID }) (*PostResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.PostSvc.VotePost(ctx, string(args.PostId), c.UserID)
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: p, root: r}, nil
}

func (r *RootResolver) UnvotePost(ctx context.Context, args struct{ PostId graphql.ID }) (*PostResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.PostSvc.UnvotePost(ctx, string(args.PostId), c.UserID)
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: p, root: r}, nil
}

