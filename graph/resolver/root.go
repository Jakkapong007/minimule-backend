// Package resolver implements graph-gophers/graphql-go resolvers for miniMule.
// No code generation required — resolver methods are matched by reflection.
package resolver

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/middleware"
	"github.com/jakka/minimule-backend/internal/service"
)

// RootResolver is the top-level resolver passed to graphql.MustParseSchema.
type RootResolver struct {
	Auth         *service.AuthService
	ProductSvc   *service.ProductService
	CategorySvc  *service.CategoryService
	CartSvc      *service.CartService
	OrderSvc     *service.OrderService
	PostSvc      *service.PostService
	ReviewSvc    *service.ReviewService
	PaymentSvc   *service.PaymentService
	ShippingSvc  *service.ShippingService
	PromotionSvc *service.PromotionService
	NotifSvc     *service.NotificationService
	SearchSvc    *service.SearchService
}

// ── Auth helpers ───────────────────────────────────────────────────────────────

func requireClaims(ctx context.Context) (*middleware.AuthClaims, error) {
	c, ok := middleware.ClaimsFromCtx(ctx)
	if !ok {
		return nil, errors.New("authentication required")
	}
	return c, nil
}

func requireAdmin(ctx context.Context) (*middleware.AuthClaims, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if c.Role != model.UserRoleAdmin {
		return nil, errors.New("admin access required")
	}
	return c, nil
}

// ── Queries ───────────────────────────────────────────────────────────────────

func (r *RootResolver) Me(ctx context.Context) (*UserResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, nil // nullable field — return nil, not error
	}
	user, err := r.Auth.GetUser(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: user, root: r}, nil
}

func (r *RootResolver) User(ctx context.Context, args struct{ ID graphql.ID }) (*UserResolver, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	user, err := r.Auth.GetUser(ctx, string(args.ID))
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	return &UserResolver{u: user, root: r}, err
}

func (r *RootResolver) Products(ctx context.Context, args struct {
	Limit      *int32
	Offset     *int32
	CategoryId *graphql.ID
	MinPrice   *float64
	MaxPrice   *float64
	IsFeatured *bool
	Sort       *string
}) ([]*ProductResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	var catID *string
	if args.CategoryId != nil {
		s := string(*args.CategoryId)
		catID = &s
	}
	sort := ""
	if args.Sort != nil {
		sort = *args.Sort
	}
	products, err := r.ProductSvc.ListProductsFiltered(ctx, limit, offset, catID, args.MinPrice, args.MaxPrice, args.IsFeatured, sort)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*ProductResolver, len(products))
	for i, p := range products {
		resolvers[i] = &ProductResolver{p: p, root: r}
	}
	return resolvers, nil
}

func (r *RootResolver) Product(ctx context.Context, args struct{ ID graphql.ID }) (*ProductResolver, error) {
	p, err := r.ProductSvc.GetProduct(ctx, string(args.ID))
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ProductResolver{p: p, root: r}, nil
}

func (r *RootResolver) Categories(ctx context.Context) ([]*CategoryResolver, error) {
	cats, err := r.CategorySvc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*CategoryResolver, len(cats))
	for i, c := range cats {
		resolvers[i] = &CategoryResolver{c: c}
	}
	return resolvers, nil
}

func (r *RootResolver) MyCart(ctx context.Context) (*CartResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, nil // nullable
	}
	cart, err := r.CartSvc.GetOrCreateCart(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	return &CartResolver{cart: cart, root: r}, nil
}

func (r *RootResolver) MyOrders(ctx context.Context) ([]*OrderResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	orders, err := r.OrderSvc.ListUserOrders(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*OrderResolver, len(orders))
	for i, o := range orders {
		resolvers[i] = &OrderResolver{o: o, root: r}
	}
	return resolvers, nil
}

func (r *RootResolver) Order(ctx context.Context, args struct{ ID graphql.ID }) (*OrderResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	o, err := r.OrderSvc.GetOrder(ctx, string(args.ID), c.UserID)
	if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrForbidden) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &OrderResolver{o: o, root: r}, nil
}

// ── Feed Queries ──────────────────────────────────────────────────────────────

func (r *RootResolver) Feed(ctx context.Context, args struct {
	Limit  *int32
	Offset *int32
}) ([]*PostResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	posts, err := r.PostSvc.GetFeed(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return wrapPosts(posts, r), nil
}

func (r *RootResolver) Post(ctx context.Context, args struct{ ID graphql.ID }) (*PostResolver, error) {
	p, err := r.PostSvc.GetPost(ctx, string(args.ID))
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: p, root: r}, nil
}

func (r *RootResolver) UserPosts(ctx context.Context, args struct {
	UserID graphql.ID
	Limit  *int32
	Offset *int32
}) ([]*PostResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	posts, err := r.PostSvc.GetPostsByUser(ctx, string(args.UserID), limit, offset)
	if err != nil {
		return nil, err
	}
	return wrapPosts(posts, r), nil
}

func (r *RootResolver) StickerDesigns(ctx context.Context, args struct {
	Limit  *int32
	Offset *int32
}) ([]*PostResolver, error) {
	limit, offset := paginationArgs(args.Limit, args.Offset, 20)
	posts, err := r.PostSvc.GetStickerDesigns(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return wrapPosts(posts, r), nil
}

// ── Mutations ─────────────────────────────────────────────────────────────────

func (r *RootResolver) Register(ctx context.Context, args struct {
	Email    string
	Password string
	Name     string
}) (*UserResolver, error) {
	user, err := r.Auth.Register(ctx, args.Email, args.Password, args.Name)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: user, root: r}, nil
}

func (r *RootResolver) Login(ctx context.Context, args struct {
	Email    string
	Password string
}) (string, error) {
	return r.Auth.Login(ctx, args.Email, args.Password)
}

func (r *RootResolver) CreateProduct(ctx context.Context, args struct {
	Input model.CreateProductInput
}) (*ProductResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	if c.Role != model.UserRoleArtist && c.Role != model.UserRoleAdmin {
		return nil, errors.New("artist or admin access required")
	}
	p, err := r.ProductSvc.CreateProduct(ctx, args.Input, c.UserID)
	if err != nil {
		return nil, err
	}
	return &ProductResolver{p: p, root: r}, nil
}

func (r *RootResolver) AddToCart(ctx context.Context, args struct {
	ProductId graphql.ID
	Quantity  int32
}) (*CartResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	product, err := r.ProductSvc.GetProduct(ctx, string(args.ProductId))
	if errors.Is(err, service.ErrNotFound) {
		return nil, errors.New("product not found")
	}
	if err != nil {
		return nil, err
	}
	cart, err := r.CartSvc.AddToCartWithPrice(ctx, c.UserID, string(args.ProductId), int(args.Quantity), product.BasePrice)
	if err != nil {
		return nil, err
	}
	return &CartResolver{cart: cart, root: r}, nil
}

func (r *RootResolver) UpdateCartItem(ctx context.Context, args struct {
	CartItemId graphql.ID
	Quantity   int32
}) (*CartResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	cart, err := r.CartSvc.UpdateCartItem(ctx, c.UserID, string(args.CartItemId), int(args.Quantity))
	if err != nil {
		return nil, err
	}
	return &CartResolver{cart: cart, root: r}, nil
}

func (r *RootResolver) CreateOrder(ctx context.Context, args struct {
	Input model.CreateOrderInput
}) (*OrderResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	o, err := r.OrderSvc.CreateOrder(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &OrderResolver{o: o, root: r}, nil
}

// ── Feed Mutations ────────────────────────────────────────────────────────────

func (r *RootResolver) CreatePost(ctx context.Context, args struct {
	Input model.CreatePostInput
}) (*PostResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.PostSvc.CreatePost(ctx, c.UserID, args.Input)
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: p, root: r}, nil
}

func (r *RootResolver) DeletePost(ctx context.Context, args struct{ ID graphql.ID }) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	if err := r.PostSvc.DeletePost(ctx, string(args.ID), c.UserID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *RootResolver) LikePost(ctx context.Context, args struct{ PostId graphql.ID }) (*PostResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.PostSvc.LikePost(ctx, string(args.PostId), c.UserID)
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: p, root: r}, nil
}

func (r *RootResolver) UnlikePost(ctx context.Context, args struct{ PostId graphql.ID }) (*PostResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.PostSvc.UnlikePost(ctx, string(args.PostId), c.UserID)
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: p, root: r}, nil
}

func (r *RootResolver) AddComment(ctx context.Context, args struct {
	PostId graphql.ID
	Body   string
}) (*PostCommentResolver, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return nil, err
	}
	comment, err := r.PostSvc.AddComment(ctx, string(args.PostId), c.UserID, args.Body)
	if err != nil {
		return nil, err
	}
	return &PostCommentResolver{c: comment, root: r}, nil
}

func (r *RootResolver) DeleteComment(ctx context.Context, args struct{ CommentId graphql.ID }) (bool, error) {
	c, err := requireClaims(ctx)
	if err != nil {
		return false, err
	}
	if err := r.PostSvc.DeleteComment(ctx, string(args.CommentId), c.UserID); err != nil {
		return false, err
	}
	return true, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func paginationArgs(limit, offset *int32, defaultLimit int) (int, int) {
	l, o := defaultLimit, 0
	if limit != nil {
		l = int(*limit)
	}
	if offset != nil {
		o = int(*offset)
	}
	return l, o
}

func wrapPosts(posts []*model.Post, r *RootResolver) []*PostResolver {
	resolvers := make([]*PostResolver, len(posts))
	for i, p := range posts {
		resolvers[i] = &PostResolver{p: p, root: r}
	}
	return resolvers
}
