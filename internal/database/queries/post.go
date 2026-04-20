package queries

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jakka/minimule-backend/graph/model"
)

// ── Feed / Posts ──────────────────────────────────────────────────────────────

func (q *Queries) GetFeed(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, caption, image_url, is_sticker_design,
		       COALESCE(visibility, 'public'), like_count, comment_count, COALESCE(vote_count,0), category_id, created_at, updated_at
		FROM posts
		WHERE COALESCE(visibility,'public') = 'public'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsFull(rows)
}

func (q *Queries) GetPostsByUser(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, caption, image_url, is_sticker_design,
		       COALESCE(visibility, 'public'), like_count, comment_count, COALESCE(vote_count,0), category_id, created_at, updated_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsFull(rows)
}

func (q *Queries) GetStickerDesigns(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, user_id, caption, image_url, is_sticker_design,
		       COALESCE(visibility,'public'), like_count, comment_count, COALESCE(vote_count,0), category_id, created_at, updated_at
		FROM posts
		WHERE is_sticker_design = TRUE AND COALESCE(visibility,'public') = 'public'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostsFull(rows)
}

func (q *Queries) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
	p := &model.Post{}
	var updatedAt *time.Time
	err := q.pool.QueryRow(ctx, `
		SELECT id, user_id, caption, image_url, is_sticker_design,
		       COALESCE(visibility,'public'), like_count, comment_count, COALESCE(vote_count,0), category_id, created_at, updated_at
		FROM posts WHERE id = $1
	`, id).Scan(&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.IsStickerDesign,
		&p.Visibility, &p.LikeCount, &p.CommentCount, &p.VoteCount, &p.CategoryID, &p.CreatedAt, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.UpdatedAt = updatedAt
	return p, nil
}

func (q *Queries) CreatePost(ctx context.Context, userID, imageURL string, caption *string, isStickerDesign bool, visibility string, categoryID *string) (*model.Post, error) {
	p := &model.Post{}
	var updatedAt *time.Time
	err := q.pool.QueryRow(ctx, `
		INSERT INTO posts (user_id, caption, image_url, is_sticker_design, visibility, category_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, caption, image_url, is_sticker_design,
		          COALESCE(visibility,'public'), like_count, comment_count, COALESCE(vote_count,0), category_id, created_at, updated_at
	`, userID, caption, imageURL, isStickerDesign, visibility, categoryID).Scan(
		&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.IsStickerDesign,
		&p.Visibility, &p.LikeCount, &p.CommentCount, &p.VoteCount, &p.CategoryID, &p.CreatedAt, &updatedAt,
	)
	p.UpdatedAt = updatedAt
	return p, err
}

func (q *Queries) DeletePost(ctx context.Context, id, ownerID string) error {
	tag, err := q.pool.Exec(ctx, `DELETE FROM posts WHERE id = $1 AND user_id = $2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Likes ─────────────────────────────────────────────────────────────────────

func (q *Queries) LikePost(ctx context.Context, postID, userID string) error {
	_, err := q.pool.Exec(ctx, `
		INSERT INTO post_likes (post_id, user_id) VALUES ($1, $2)
	`, postID, userID)
	if err != nil && strings.Contains(err.Error(), "unique") {
		return ErrDuplicate
	}
	return err
}

func (q *Queries) UnlikePost(ctx context.Context, postID, userID string) error {
	_, err := q.pool.Exec(ctx, `
		DELETE FROM post_likes WHERE post_id = $1 AND user_id = $2
	`, postID, userID)
	return err
}

func (q *Queries) IsPostLikedByUser(ctx context.Context, postID, userID string) (bool, error) {
	var exists bool
	err := q.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM post_likes WHERE post_id = $1 AND user_id = $2)
	`, postID, userID).Scan(&exists)
	return exists, err
}

// ── Comments ──────────────────────────────────────────────────────────────────

func (q *Queries) GetCommentsByPost(ctx context.Context, postID string, limit, offset int) ([]*model.PostComment, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, post_id, user_id, body, created_at
		FROM post_comments
		WHERE post_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*model.PostComment
	for rows.Next() {
		c := &model.PostComment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (q *Queries) AddComment(ctx context.Context, postID, userID, body string) (*model.PostComment, error) {
	c := &model.PostComment{}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO post_comments (post_id, user_id, body)
		VALUES ($1, $2, $3)
		RETURNING id, post_id, user_id, body, created_at
	`, postID, userID, body).Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.CreatedAt)
	return c, err
}

func (q *Queries) DeleteComment(ctx context.Context, commentID, ownerID string) error {
	tag, err := q.pool.Exec(ctx, `
		DELETE FROM post_comments WHERE id = $1 AND user_id = $2
	`, commentID, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (q *Queries) GetCommentByID(ctx context.Context, id string) (*model.PostComment, error) {
	c := &model.PostComment{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, post_id, user_id, body, created_at
		FROM post_comments WHERE id = $1
	`, id).Scan(&c.ID, &c.PostID, &c.UserID, &c.Body, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func scanPostsFull(rows pgx.Rows) ([]*model.Post, error) {
	var out []*model.Post
	for rows.Next() {
		p := &model.Post{}
		var updatedAt *time.Time
		if err := rows.Scan(&p.ID, &p.UserID, &p.Caption, &p.ImageURL, &p.IsStickerDesign,
			&p.Visibility, &p.LikeCount, &p.CommentCount, &p.VoteCount, &p.CategoryID, &p.CreatedAt, &updatedAt); err != nil {
			return nil, err
		}
		p.UpdatedAt = updatedAt
		out = append(out, p)
	}
	return out, rows.Err()
}

// scanPosts kept for backward compatibility — delegates to scanPostsFull
var scanPosts = scanPostsFull
