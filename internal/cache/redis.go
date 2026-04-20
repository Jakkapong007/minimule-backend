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

func Connect(ctx context.Context, cfg *config.Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	const maxAttempts = 10
	var err error
	for i := 1; i <= maxAttempts; i++ {
		err = rdb.Ping(ctx).Err()
		if err == nil {
			break
		}
		wait := time.Duration(i*i) * time.Second
		slog.Warn("redis not ready, retrying", "attempt", i, "wait", wait, "err", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	if err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	slog.Info("redis connected", "addr", cfg.RedisAddr)
	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// ── Generic helpers ────────────────────────────────────────────────────────────

func (c *Client) SetJSON(ctx context.Context, key string, val any, ttl time.Duration) error {
	b, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("cache set %s: marshal: %w", key, err)
	}
	return c.rdb.Set(ctx, key, b, ttl).Err()
}

func (c *Client) GetJSON(ctx context.Context, key string, dest any) error {
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
	return c.rdb.Del(ctx, keys...).Err()
}

// ── Refresh tokens ─────────────────────────────────────────────────────────────

const refreshPrefix = "refresh:"

func (c *Client) SetRefreshToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return c.rdb.Set(ctx, refreshPrefix+token, userID, ttl).Err()
}

func (c *Client) GetRefreshToken(ctx context.Context, token string) (string, error) {
	userID, err := c.rdb.Get(ctx, refreshPrefix+token).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return userID, err
}

func (c *Client) DeleteRefreshToken(ctx context.Context, token string) error {
	return c.rdb.Del(ctx, refreshPrefix+token).Err()
}

// ── Rate limiting (fixed window counter) ──────────────────────────────────────

const rateLimitPrefix = "rl:"

// Increment returns the new count after incrementing the counter for the key.
// On the first increment it sets the TTL to window.
func (c *Client) IncrRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := c.rdb.Pipeline()
	incr := pipe.Incr(ctx, rateLimitPrefix+key)
	pipe.Expire(ctx, rateLimitPrefix+key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
