package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ecommerce/services/api-gateway/internal/proxy"
	"github.com/ecommerce/shared/pkg/response"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware creates authentication middleware for API Gateway
func AuthMiddleware(serviceProxy *proxy.ServiceProxy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "Authorization header required")
				return
			}

			// Check Bearer prefix
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				response.Unauthorized(w, "Invalid authorization header format")
				return
			}

			token := tokenParts[1]

			// Validate token via User Service gRPC
			user, err := serviceProxy.ValidateTokenViaGRPC(r.Context(), token)
			if err != nil {
				response.Unauthorized(w, "Invalid or expired token")
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware to check user role at gateway level
func RequireRole(serviceProxy *proxy.ServiceProxy, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (should be set by AuthMiddleware)
			userInterface := r.Context().Value(UserContextKey)
			if userInterface == nil {
				response.Unauthorized(w, "User not authenticated")
				return
			}

			// Type assertion to get user data
			// Note: This depends on the structure returned by ValidateTokenViaGRPC
			// You might need to adjust based on your actual user structure

			// For now, let's extract from token again (alternative approach)
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenParts := strings.Split(authHeader, " ")
				if len(tokenParts) == 2 {
					token := tokenParts[1]
					user, err := serviceProxy.ValidateTokenViaGRPC(r.Context(), token)
					if err != nil {
						response.Unauthorized(w, "Invalid token")
						return
					}

					// Check if user has required role
					hasRole := false
					for _, role := range roles {
						if user.Role == role {
							hasRole = true
							break
						}
					}

					if !hasRole {
						response.Forbidden(w, "Insufficient permissions")
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth middleware for endpoints that don't require auth but can use it
func OptionalAuth(serviceProxy *proxy.ServiceProxy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenParts := strings.Split(authHeader, " ")
				if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
					token := tokenParts[1]
					user, err := serviceProxy.ValidateTokenViaGRPC(r.Context(), token)
					if err == nil && user != nil {
						ctx := context.WithValue(r.Context(), UserContextKey, user)
						r = r.WithContext(ctx)
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
