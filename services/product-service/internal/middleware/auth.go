package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/client"
)

// AuthMiddleware xác thực token
func AuthMiddleware(userClient client.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Nếu không có userClient, skip authentication
		if userClient == nil {
			c.Next()
			return
		}

		// Lấy token từ header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			return
		}

		// Parse Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			return
		}

		token := parts[1]

		// Validate token với User Service
		userInfo, err := userClient.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Lưu user info vào context
		c.Set("user", userInfo)
		c.Set("user_id", userInfo.ID)
		c.Set("user_role", userInfo.Role)

		c.Next()
	}
}

// RequireAdmin middleware kiểm tra quyền admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
			})
			return
		}

		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware - không bắt buộc auth nhưng nếu có token thì validate
func OptionalAuthMiddleware(userClient client.UserServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userClient == nil {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Không có token, tiếp tục nhưng không set user
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			userInfo, err := userClient.ValidateToken(c.Request.Context(), token)
			if err == nil {
				c.Set("user", userInfo)
				c.Set("user_id", userInfo.ID)
				c.Set("user_role", userInfo.Role)
			}
		}

		c.Next()
	}
}

// GetUserFromContext helper để lấy user info từ context
func GetUserFromContext(c *gin.Context) (*client.UserInfo, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userInfo, ok := user.(*client.UserInfo)
	return userInfo, ok
}
