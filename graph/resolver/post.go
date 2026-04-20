package resolver

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/middleware"
)

// ── PostResolver ──────────────────────────────────────────────────────────────

type PostResolver struct {
	p    *model.Post
	root *RootResolver
}

func (r *PostResolver) ID() graphql.ID           { return graphql.ID(r.p.ID) }
func (r *PostResolver) Caption() *string         { return r.p.Caption }
func (r *PostResolver) ImageUrl() string         { return r.p.ImageURL }
func (r *PostResolver) IsStickerDesign() bool    { return r.p.IsStickerDesign }
func (r *PostResolver) Visibility() string       { return r.p.Visibility }
func (r *PostResolver) LikeCount() int32         { return int32(r.p.LikeCount) }
func (r *PostResolver) CommentCount() int32      { return int32(r.p.CommentCount) }
func (r *PostResolver) VoteCount() int32         { return int32(r.p.VoteCount) }
func (r *PostResolver) CreatedAt() graphql.Time  { return graphql.Time{Time: r.p.CreatedAt} }

func (r *PostResolver) IsVotedByMe(ctx context.Context) (bool, error) {
	c, ok := middleware.ClaimsFromCtx(ctx)
	if !ok {
		return false, nil
	}
	return r.root.PostSvc.IsVotedByUser(ctx, r.p.ID, c.UserID)
}

func (r *PostResolver) Category(ctx context.Context) (*CategoryResolver, error) {
	if r.p.CategoryID == nil {
		return nil, nil
	}
	cat, err := r.root.ProductSvc.GetCategory(ctx, *r.p.CategoryID)
	if err != nil {
		return nil, nil
	}
	return &CategoryResolver{c: cat}, nil
}

func (r *PostResolver) User(ctx context.Context) (*UserResolver, error) {
	if r.p.User != nil {
		return &UserResolver{u: r.p.User, root: r.root}, nil
	}
	user, err := r.root.Auth.GetUser(ctx, r.p.UserID)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: user, root: r.root}, nil
}

func (r *PostResolver) IsLikedByMe(ctx context.Context) (bool, error) {
	c, ok := middleware.ClaimsFromCtx(ctx)
	if !ok {
		return false, nil
	}
	return r.root.PostSvc.IsLikedByUser(ctx, r.p.ID, c.UserID)
}

func (r *PostResolver) Comments(ctx context.Context, args struct {
	Limit  *int32
	Offset *int32
}) ([]*PostCommentResolver, error) {
	limit := 50
	offset := 0
	if args.Limit != nil {
		limit = int(*args.Limit)
	}
	if args.Offset != nil {
		offset = int(*args.Offset)
	}
	comments, err := r.root.PostSvc.GetComments(ctx, r.p.ID, limit, offset)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*PostCommentResolver, len(comments))
	for i, c := range comments {
		resolvers[i] = &PostCommentResolver{c: c, root: r.root}
	}
	return resolvers, nil
}

// ── PostCommentResolver ───────────────────────────────────────────────────────

type PostCommentResolver struct {
	c    *model.PostComment
	root *RootResolver
}

func (r *PostCommentResolver) ID() graphql.ID          { return graphql.ID(r.c.ID) }
func (r *PostCommentResolver) Body() string            { return r.c.Body }
func (r *PostCommentResolver) CreatedAt() graphql.Time { return graphql.Time{Time: r.c.CreatedAt} }

func (r *PostCommentResolver) Post(ctx context.Context) (*PostResolver, error) {
	if r.c.Post != nil {
		return &PostResolver{p: r.c.Post, root: r.root}, nil
	}
	post, err := r.root.PostSvc.GetPost(ctx, r.c.PostID)
	if err != nil {
		return nil, err
	}
	return &PostResolver{p: post, root: r.root}, nil
}

func (r *PostCommentResolver) User(ctx context.Context) (*UserResolver, error) {
	if r.c.User != nil {
		return &UserResolver{u: r.c.User, root: r.root}, nil
	}
	user, err := r.root.Auth.GetUser(ctx, r.c.UserID)
	if err != nil {
		return nil, err
	}
	return &UserResolver{u: user, root: r.root}, nil
}
