package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jakka/minimule-backend/internal/auth"
	"github.com/jakka/minimule-backend/graph/model"
)

type contextKey string

const ContextKeyClaims contextKey = "claims"

// AuthClaims is a minimal identity snapshot stored in the request context.
// It is populated from the validated JWT; no DB round-trip occurs here.
type AuthClaims struct {
	UserID string
	Email  string
	Role   model.UserRole
}

// Auth validates Bearer tokens and, when valid, injects AuthClaims into the
// request context. Requests without a valid token proceed unauthenticated —
// individual resolvers decide whether to enforce auth.
func Auth(jwtSvc *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if strings.HasPrefix(header, "Bearer ") {
				raw := strings.TrimPrefix(header, "Bearer ")
				if claims, err := jwtSvc.ValidateAccessToken(raw); err == nil {
					ac := &AuthClaims{
						UserID: claims.UserID,
						Email:  claims.Email,
						Role:   claims.Role,
					}
					r = r.WithContext(context.WithValue(r.Context(), ContextKeyClaims, ac))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromCtx retrieves the AuthClaims stored by the Auth middleware.
func ClaimsFromCtx(ctx context.Context) (*AuthClaims, bool) {
	c, ok := ctx.Value(ContextKeyClaims).(*AuthClaims)
	return c, ok && c != nil
}
