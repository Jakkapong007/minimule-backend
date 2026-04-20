package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/jakka/minimule-backend/internal/cache"
	"github.com/jakka/minimule-backend/internal/config"
)

// RateLimit applies a fixed-window counter per client IP. Failures in Redis
// are fail-open: the request is allowed and the error is silently ignored.
func RateLimit(c *cache.Client, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			key := fmt.Sprintf("rl:%s", ip)

			count, err := c.IncrRateLimit(r.Context(), key, cfg.RateLimitWindow)
			if err == nil && count > int64(cfg.RateLimitRequests) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", cfg.RateLimitWindow.Seconds()))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"errors":[{"message":"rate limit exceeded"}]}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// realIP extracts the client IP, honouring X-Real-IP and X-Forwarded-For set
// by a trusted proxy (e.g. nginx / k8s ingress).
func realIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For: client, proxy1, proxy2 — leftmost is the client
		for i := 0; i < len(ip); i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
