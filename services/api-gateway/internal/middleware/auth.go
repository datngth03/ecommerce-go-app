package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UserInfo represents validated user information
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

// AuthMiddleware validates JWT token by calling User Service
func AuthMiddleware(userServiceAddr string) gin.HandlerFunc {
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

		// Validate token with User Service
		userInfo, err := validateTokenWithUserService(userServiceAddr, token)
		if err != nil {
			ErrorLogger(c, err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user", userInfo)
		c.Set("user_id", userInfo.ID)
		c.Set("user_role", userInfo.Role)

		c.Next()
	}
}

// RequireAdmin checks if user has admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			c.Abort()
			return
		}

		if role != "admin" {
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
func OptionalAuth(userServiceAddr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			userInfo, err := validateTokenWithUserService(userServiceAddr, token)
			if err == nil {
				c.Set("user", userInfo)
				c.Set("user_id", userInfo.ID)
				c.Set("user_role", userInfo.Role)
			}
		}

		c.Next()
	}
}

// validateTokenWithUserService calls User Service HTTP API to validate token
func validateTokenWithUserService(userServiceAddr, token string) (*UserInfo, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Build request URL
	url := fmt.Sprintf("http://%s/api/v1/auth/validate", userServiceAddr)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add token to header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call user service: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		User UserInfo `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.User, nil
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
