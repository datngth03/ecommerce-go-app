package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

// UserInfo represents validated user information
type UserInfo struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	IsActive bool   `json:"is_active"`
}

// AuthMiddleware validates JWT token by calling User Service via proxy
func AuthMiddleware(userProxy *proxy.UserProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token with User Service via proxy
		userInfo, err := validateTokenWithUserProxy(userProxy, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user", userInfo)
		c.Set("user_id", userInfo.ID)
		// No role field available in current schema

		c.Next()
	}
}

// RequireAdmin checks if user has admin privileges
// Since role field is not in current schema, we check for admin email
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		userInfo, ok := user.(*UserInfo)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid user context",
			})
			c.Abort()
			return
		}

		// Check if user is admin by email (temporary solution until role field is added)
		if userInfo.Email != "admin@example.com" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth tries to validate token but doesn't fail if missing
func OptionalAuth(userProxy *proxy.UserProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			userInfo, err := validateTokenWithUserProxy(userProxy, token)
			if err == nil {
				c.Set("user", userInfo)
				c.Set("user_id", userInfo.ID)
				// No role field available in current schema
			}
		}

		c.Next()
	}
}

// validateTokenWithUserProxy calls User Service via gRPC proxy to validate token
func validateTokenWithUserProxy(userProxy *proxy.UserProxy, token string) (*UserInfo, error) {
	// Call user service via gRPC
	ctx := context.Background()
	response, err := userProxy.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token with user service: %w", err)
	}

	// Check if token is valid
	if !response.GetValid() {
		return nil, fmt.Errorf("invalid token: %s", response.GetMessage())
	}

	// Convert protobuf response to UserInfo
	userInfo := &UserInfo{
		ID:       response.GetUserId(),
		Email:    response.GetEmail(),
		Name:     "",   // Not available in ValidateTokenResponse - would need separate GetUser call
		Phone:    "",   // Not available in ValidateTokenResponse - would need separate GetUser call
		IsActive: true, // Assuming if token is valid, user is active
	}

	return userInfo, nil
}

// GetUserFromContext retrieves user info from context
func GetUserFromContext(c *gin.Context) (*UserInfo, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userInfo, ok := user.(*UserInfo)
	return userInfo, ok
}
