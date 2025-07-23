// internal/api_gateway/middleware/auth_middleware.go
package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth" // Generated Auth gRPC client
	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides authentication and authorization middleware.
// AuthMiddleware cung cấp middleware xác thực và ủy quyền.
type AuthMiddleware struct {
	AuthClient auth_client.AuthServiceClient
}

// NewAuthMiddleware creates a new instance of AuthMiddleware.
// NewAuthMiddleware tạo một thể hiện mới của AuthMiddleware.
func NewAuthMiddleware(authClient auth_client.AuthServiceClient) *AuthMiddleware {
	return &AuthMiddleware{
		AuthClient: authClient,
	}
}

// AuthRequired is a Gin middleware to validate JWT tokens.
// AuthRequired là một Gin middleware để xác thực JWT tokens.
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		accessToken := parts[1]

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second) // Short timeout for token validation
		defer cancel()

		validateResp, err := m.AuthClient.ValidateToken(ctx, &auth_client.ValidateTokenRequest{
			AccessToken: accessToken,
		})

		if err != nil {
			log.Printf("Error validating token with Auth Service: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate token"})
			return
		}

		if !validateResp.GetIsValid() {
			log.Printf("Token validation failed: %s", validateResp.GetErrorMessage())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": validateResp.GetErrorMessage()})
			return
		}

		// Token is valid, set user ID in Gin context for subsequent handlers
		c.Set("userID", validateResp.GetUserId())
		c.Next() // Continue to the next handler
	}
}

// GetUserIDFromContext is a helper function to retrieve user ID from Gin context.
// GetUserIDFromContext là một hàm trợ giúp để lấy ID người dùng từ Gin context.
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	strUserID, ok := userID.(string)
	return strUserID, ok
}
