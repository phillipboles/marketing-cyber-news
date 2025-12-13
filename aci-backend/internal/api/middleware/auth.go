package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/pkg/jwt"
)

// contextKey is a custom type for context keys to avoid collisions
type authContextKey string

const (
	userClaimsKey authContextKey = "user_claims"
)

// Auth middleware extracts and validates JWT from Authorization header
func Auth(jwtService jwt.Service) func(http.Handler) http.Handler {
	if jwtService == nil {
		panic("jwtService cannot be nil")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Missing authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "Invalid authorization header format")
				return
			}

			token := parts[1]
			if token == "" {
				response.Unauthorized(w, "Missing access token")
				return
			}

			// Validate token
			claims, err := jwtService.ValidateAccessToken(token)
			if err != nil {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}

			// Store claims in context
			ctx := context.WithValue(r.Context(), userClaimsKey, claims)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware checks if user has required role
func RequireRole(role string) func(http.Handler) http.Handler {
	if role == "" {
		panic("role cannot be empty")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user claims from context
			claims, ok := GetUserFromContext(r.Context())
			if !ok {
				response.Unauthorized(w, "Authentication required")
				return
			}

			// Check if user has required role
			if claims.Role != role {
				response.Forbidden(w, fmt.Sprintf("Required role: %s", role))
				return
			}

			// User has required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware checks if user has admin role
func RequireAdmin() func(http.Handler) http.Handler {
	return RequireRole("admin")
}

// GetUserFromContext retrieves user claims from request context
func GetUserFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*jwt.Claims)
	return claims, ok
}
