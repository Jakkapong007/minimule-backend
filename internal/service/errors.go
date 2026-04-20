package service

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("authentication required")
	ErrForbidden    = errors.New("permission denied")
	ErrBadRequest   = errors.New("bad request")
	ErrConflict     = errors.New("already exists")
	ErrInvalidCreds = errors.New("invalid email or password")
)
