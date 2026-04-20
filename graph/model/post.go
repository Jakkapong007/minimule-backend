package model

import "time"

// Post represents a user-created post in the feed.
type Post struct {
	ID              string
	UserID          string
	Caption         *string
	ImageURL        string
	IsStickerDesign bool
	Visibility      string
	LikeCount       int
	CommentCount    int
	VoteCount       int
	CategoryID      *string
	CreatedAt       time.Time
	UpdatedAt       *time.Time

	User     *User
	Comments []*PostComment
}

// PostComment is a comment on a Post.
type PostComment struct {
	ID        string
	PostID    string
	UserID    string
	Body      string
	CreatedAt time.Time

	Post *Post
	User *User
}

// CreatePostInput is the input for the createPost mutation.
type CreatePostInput struct {
	Caption         *string
	ImageURL        string
	IsStickerDesign *bool
	Visibility      *string
	CategoryID      *string
}
