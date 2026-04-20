package service

import (
	"context"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

type CategoryService struct {
	q *queries.Queries
}

func NewCategoryService(q *queries.Queries) *CategoryService {
	return &CategoryService{q: q}
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]*model.Category, error) {
	return s.q.ListCategories(ctx)
}
