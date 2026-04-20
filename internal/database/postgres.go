package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jakka/minimule-backend/internal/config"
)

// Pool wraps pgxpool.Pool so callers don't need to import pgx directly.
type Pool struct {
	*pgxpool.Pool
}

// Connect creates a pgx connection pool and verifies connectivity.
// It retries up to maxAttempts with exponential back-off.
func Connect(ctx context.Context, cfg *config.Config) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MinConns = cfg.DBMinConns
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 1 * time.Minute

	const maxAttempts = 10
	var pool *pgxpool.Pool
	for i := 1; i <= maxAttempts; i++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolCfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				break
			} else {
				err = pingErr
				pool.Close()
			}
		}
		wait := time.Duration(i*i) * time.Second
		slog.Warn("database not ready, retrying",
			"attempt", i, "max", maxAttempts, "wait", wait, "err", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	if err != nil {
		return nil, fmt.Errorf("connect to database after %d attempts: %w", maxAttempts, err)
	}

	slog.Info("database connected", "max_conns", cfg.DBMaxConns)
	return &Pool{pool}, nil
}
