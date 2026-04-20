package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jakka/minimule-backend/graph/model"
	"github.com/jakka/minimule-backend/internal/auth"
	"github.com/jakka/minimule-backend/internal/database/queries"
)

type AuthService struct {
	q   *queries.Queries
	jwt *auth.JWTService
}

func NewAuthService(q *queries.Queries, jwt *auth.JWTService) *AuthService {
	return &AuthService{q: q, jwt: jwt}
}

// Register creates a new user account. Returns the created User.
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*model.User, error) {
	if email == "" || password == "" || name == "" {
		return nil, fmt.Errorf("%w: email, password and name are required", ErrBadRequest)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", ErrBadRequest)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	fullName := name
	user, err := s.q.CreateUser(ctx, email, hash, &fullName, nil, model.UserRoleCustomer)
	if errors.Is(err, queries.ErrDuplicate) {
		return nil, fmt.Errorf("%w: email already registered", ErrConflict)
	}
	return user, err
}

// Login verifies credentials and returns a signed JWT access token.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if errors.Is(err, queries.ErrNotFound) {
		return "", ErrInvalidCreds
	}
	if err != nil {
		return "", err
	}

	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		return "", ErrInvalidCreds
	}

	if !user.IsActive {
		return "", fmt.Errorf("%w: account is disabled", ErrForbidden)
	}

	// For mobile apps we use the refresh TTL as the token lifetime (e.g. 7 days).
	token, err := s.jwt.GenerateLongLivedToken(user)
	if err != nil {
		return "", err
	}
	return token, nil
}

// GetUser fetches a user by ID.
func (s *AuthService) GetUser(ctx context.Context, id string) (*model.User, error) {
	user, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return user, err
}

// GetUserProfile fetches the extended profile for a user.
func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*model.UserProfile, error) {
	p, err := s.q.GetUserProfile(ctx, userID)
	if errors.Is(err, queries.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

// GetUserAddresses returns all addresses for a user.
func (s *AuthService) GetUserAddresses(ctx context.Context, userID string) ([]*model.UserAddress, error) {
	return s.q.GetUserAddresses(ctx, userID)
}
