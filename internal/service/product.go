package service

import (
	"context"
	"errors"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

type ProductService struct {
	q *queries.Queries
}

func NewProductService(q *queries.Queries) *ProductService {
	return &ProductService{q: q}
}

func (s *ProductService) CreateProduct(ctx context.Context, input model.CreateProductInput, sellerID string) (*model.Product, error) {
	return s.q.CreateProduct(ctx, input, sellerID)
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*model.Product, error) {
	p, err := s.q.GetProductByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.q.ListProducts(ctx, limit, offset)
}

func (s *ProductService) ListProductsFiltered(ctx context.Context, limit, offset int, categoryID *string, minPrice, maxPrice *float64, isFeatured *bool, sort string) ([]*model.Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.q.ListProductsFiltered(ctx, limit, offset, categoryID, minPrice, maxPrice, isFeatured, sort)
}

func (s *ProductService) GetProductImages(ctx context.Context, productID string) ([]*model.ProductImage, error) {
	return s.q.GetProductImages(ctx, productID)
}

func (s *ProductService) GetProductVariants(ctx context.Context, productID string) ([]*model.ProductVariant, error) {
	return s.q.GetProductVariants(ctx, productID)
}

func (s *ProductService) GetCategory(ctx context.Context, id string) (*model.Category, error) {
	c, err := s.q.GetCategoryByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return c, err
}
