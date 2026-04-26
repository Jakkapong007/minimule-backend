package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jakka/minimule-backend/internal/config"
)

var ErrNotFound = errors.New("cache: key not found")

type Client struct {
	rdb *redis.Client
}

// Connect returns a Redis client. If REDIS_ADDR is empty or the server is
// unreachable, it returns a no-op client so the server starts without Redis
// (rate limiting and caching are simply skipped).
func Connect(ctx context.Context, cfg *config.Config) (*Client, error) {
	if cfg.RedisAddr == "" {
		slog.Warn("redis not configured — rate limiting and caching disabled")
		return &Client{rdb: nil}, nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Warn("redis unavailable — rate limiting and caching disabled", "addr", cfg.RedisAddr, "err", err)
		_ = rdb.Close()
		return &Client{rdb: nil}, nil
	}

	slog.Info("redis connected", "addr", cfg.RedisAddr)
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// ── Generic helpers ────────────────────────────────────────────────────────────

func (c *Client) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	b, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache set %s: marshal: %w", key, err)
	}
	return c.rdb.Set(ctx, key, b, ttl).Err()
}

func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
	if c.rdb == nil {
		return ErrNotFound
	}
	b, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("cache get %s: %w", key, err)
	}
	if err := json.Unmarshal(b, dest); err != nil {
		return fmt.Errorf("cache get %s: unmarshal: %w", key, err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// ── Refresh tokens ─────────────────────────────────────────────────────────────

const refreshPrefix = "refresh:"

func (c *Client) SetRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Set(ctx, refreshPrefix+token, userID, ttl).Err()
}

func (c *Client) GetRefreshToken(ctx context.Context, token string) (string, error) {
	if c.rdb == nil {
		return "", ErrNotFound
	}
	userID, err := c.rdb.Get(ctx, refreshPrefix+token).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return userID, err
}

func (c *Client) DeleteRefreshToken(ctx context.Context, token string) error {
	if c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, refreshPrefix+token).Err()
}

// ── Rate limiting (fixed window counter) ──────────────────────────────────────

const rateLimitPrefix = "rl:"

// Increment returns the new count after incrementing the counter for the key.
// On the first increment it sets the TTL to window.
func (c *Client) IncrRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	if c.rdb == nil {
		return 0, nil
	}
	pipe := c.rdb.Pipeline()
	incr := pipe.Incr(ctx, rateLimitPrefix+key)
	pipe.Expire(ctx, rateLimitPrefix+key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
