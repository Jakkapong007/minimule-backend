package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port    string
	Env     string
	IsProd  bool

	DatabaseURL string
	DBMaxConns  int32
	DBMinConns  int32

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret     string
	JWTAccessTTL  time.Duration
	JWTRefreshTTL time.Duration

	GraphQLPlayground    bool
	QueryDepthLimit      int
	QueryComplexityLimit int

	CORSOrigins []string

	RateLimitRequests int
	RateLimitWindow   time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load() // no-op if .env is absent (e.g. in containers)

	accessTTL, err := parseDuration("JWT_ACCESS_TTL", "15m")
	if err != nil {
		return nil, err
	}
	refreshTTL, err := parseDuration("JWT_REFRESH_TTL", "168h")
	if err != nil {
		return nil, err
	}
	rateLimitWindow, err := parseDuration("RATE_LIMIT_WINDOW", "1m")
	if err != nil {
		return nil, err
	}

	env := getEnv("ENV", "development")

	cfg := &Config{
		Port:   getEnv("PORT", "8080"),
		Env:    env,
		IsProd: env == "production",

		DatabaseURL: mustEnv("DATABASE_URL"),
		DBMaxConns:  int32(getEnvInt("DB_MAX_CONNS", 25)),
		DBMinConns:  int32(getEnvInt("DB_MIN_CONNS", 5)),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		JWTSecret:     mustEnv("JWT_SECRET"),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,

		GraphQLPlayground:    getEnvBool("GRAPHQL_PLAYGROUND", !isProd(env)),
		QueryDepthLimit:      getEnvInt("QUERY_DEPTH_LIMIT", 10),
		QueryComplexityLimit: getEnvInt("QUERY_COMPLEXITY_LIMIT", 1000),

		CORSOrigins: splitComma(os.Getenv("CORS_ORIGINS")),

		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   rateLimitWindow,
	}

	return cfg, nil
}

func isProd(env string) bool { return env == "production" }

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func parseDuration(key, def string) (time.Duration, error) {
	raw := getEnv(key, def)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, raw, err)
	}
	return d, nil
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
