package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

type PostService struct {
	q    *queries.Queries
	auth *AuthService
}

func NewPostService(q *queries.Queries, auth *AuthService) *PostService {
	return &PostService{q: q, auth: auth}
}

func (s *PostService) GetFeed(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.q.GetFeed(ctx, limit, offset)
}

func (s *PostService) GetPost(ctx context.Context, id string) (*model.Post, error) {
	p, err := s.q.GetPostByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *PostService) GetPostsByUser(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.q.GetPostsByUser(ctx, userID, limit, offset)
}

func (s *PostService) GetStickerDesigns(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.q.GetStickerDesigns(ctx, limit, offset)
}

func (s *PostService) CreatePost(ctx context.Context, userID string, input model.CreatePostInput) (*model.Post, error) {
	if input.ImageURL == "" {
		return nil, fmt.Errorf("%w: imageUrl is required", ErrBadRequest)
	}
	isStickerDesign := false
	if input.IsStickerDesign != nil {
		isStickerDesign = *input.IsStickerDesign
	}
	visibility := "public"
	if input.Visibility != nil {
		visibility = *input.Visibility
	}
	return s.q.CreatePost(ctx, userID, input.ImageURL, input.Caption, isStickerDesign, visibility, input.CategoryID)
}

func (s *PostService) DeletePost(ctx context.Context, id, userID string) error {
	err := s.q.DeletePost(ctx, id, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostService) LikePost(ctx context.Context, postID, userID string) (*model.Post, error) {
	err := s.q.LikePost(ctx, postID, userID)
	if err != nil && !errors.Is(err, queries.ErrDuplicate) {
		return nil, err
	}
	return s.GetPost(ctx, postID)
}

func (s *PostService) UnlikePost(ctx context.Context, postID, userID string) (*model.Post, error) {
	if err := s.q.UnlikePost(ctx, postID, userID); err != nil {
		return nil, err
	}
	return s.GetPost(ctx, postID)
}

func (s *PostService) IsLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	return s.q.IsPostLikedByUser(ctx, postID, userID)
}

func (s *PostService) GetComments(ctx context.Context, postID string, limit, offset int) ([]*model.PostComment, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.q.GetCommentsByPost(ctx, postID, limit, offset)
}

func (s *PostService) AddComment(ctx context.Context, postID, userID, body string) (*model.PostComment, error) {
	if body == "" {
		return nil, fmt.Errorf("%w: comment body is required", ErrBadRequest)
	}
	return s.q.AddComment(ctx, postID, userID, body)
}

func (s *PostService) DeleteComment(ctx context.Context, commentID, userID string) error {
	err := s.q.DeleteComment(ctx, commentID, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostService) GetCommentByID(ctx context.Context, id string) (*model.PostComment, error) {
	c, err := s.q.GetCommentByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return c, err
}

func (s *PostService) GetShowcase(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.q.GetShowcase(ctx, limit, offset)
}

func (s *PostService) VotePost(ctx context.Context, postID, userID string) (*model.Post, error) {
	err := s.q.VotePost(ctx, postID, userID)
	if err != nil && !errors.Is(err, queries.ErrDuplicate) {
		return nil, err
	}
	return s.GetPost(ctx, postID)
}

func (s *PostService) UnvotePost(ctx context.Context, postID, userID string) (*model.Post, error) {
	if err := s.q.UnvotePost(ctx, postID, userID); err != nil {
		return nil, err
	}
	return s.GetPost(ctx, postID)
}

func (s *PostService) IsVotedByUser(ctx context.Context, postID, userID string) (bool, error) {
	return s.q.IsPostVotedByUser(ctx, postID, userID)
}
