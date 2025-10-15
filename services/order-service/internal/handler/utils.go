package handler

import (
	"github.com/gin-gonic/gin"
)

// getUserIDFromContext extracts user ID from the request context
// This is typically set by authentication middleware
func getUserIDFromContext(c *gin.Context) int64 {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}

	// Try to convert to int64
	switch v := userID.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case uint64:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}
