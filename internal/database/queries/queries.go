package queries

import (
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/jakka/minimule-backend/internal/database"
)

// Sentinel errors returned by query functions.
var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("record already exists")
)

// Queries holds the database pool and provides all SQL query methods.
type Queries struct {
	pool *database.Pool
}

// New creates a new Queries instance.
func New(pool *database.Pool) *Queries {
	return &Queries{pool: pool}
}

func isNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
