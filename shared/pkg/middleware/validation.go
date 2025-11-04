package middleware

import (
	"net/http"
	"strings"

	"github.com/datngth03/ecommerce-go-app/shared/pkg/validator"

	"github.com/gin-gonic/gin"
)

// InputSanitizationMiddleware sanitizes all string inputs to prevent XSS attacks
func InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = validator.SanitizeHTML(value)
			}
		}

		c.Next()
	}
}

// RequestSizeLimit limits the maximum request body size
func RequestSizeLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// ValidateContentType ensures requests have appropriate content type
func ValidateContentTypeMiddleware(allowedTypes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, DELETE, HEAD requests
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Content-Type header is required",
			})
			c.Abort()
			return
		}

		// Remove charset and other parameters
		contentType = strings.Split(contentType, ";")[0]
		contentType = strings.TrimSpace(contentType)

		for _, allowed := range allowedTypes {
			if contentType == allowed {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error": "Unsupported Content-Type. Allowed types: " + strings.Join(allowedTypes, ", "),
		})
		c.Abort()
	}
}

// NoSQLInjectionMiddleware checks for common NoSQL injection patterns
func NoSQLInjectionMiddleware() gin.HandlerFunc {
	dangerousPatterns := []string{
		"$where",
		"$ne",
		"$gt",
		"$gte",
		"$lt",
		"$lte",
		"$in",
		"$nin",
		"$or",
		"$and",
		"$not",
		"$nor",
		"$exists",
		"$type",
		"$regex",
	}

	return func(c *gin.Context) {
		// Check query parameters
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				for _, pattern := range dangerousPatterns {
					if strings.Contains(strings.ToLower(value), strings.ToLower(pattern)) {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Potentially malicious input detected",
						})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// PathTraversalProtectionMiddleware prevents path traversal attacks
func PathTraversalProtectionMiddleware() gin.HandlerFunc {
	dangerousPatterns := []string{
		"../",
		"..\\",
		"..%2f",
		"..%5c",
		"%2e%2e%2f",
		"%2e%2e%5c",
	}

	return func(c *gin.Context) {
		path := strings.ToLower(c.Request.URL.Path)

		for _, pattern := range dangerousPatterns {
			if strings.Contains(path, pattern) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid path detected",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// ValidateJSONMiddleware ensures request body is valid JSON for JSON endpoints
func ValidateJSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET, DELETE, HEAD requests
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Try to bind to empty interface to validate JSON structure
		var jsonData interface{}
		if err := c.ShouldBindJSON(&jsonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON format: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Reset body for next middleware/handler
		c.Set("validatedJSON", jsonData)
		c.Next()
	}
}

// Enhanced validation middleware bundle for maximum security
func EnhancedValidationMiddlewares(maxRequestSize int64) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		RequestSizeLimitMiddleware(maxRequestSize),        // Limit request size
		ValidateContentTypeMiddleware("application/json"), // Validate content type
		InputSanitizationMiddleware(),                     // Sanitize inputs
		NoSQLInjectionMiddleware(),                        // Prevent NoSQL injection
		PathTraversalProtectionMiddleware(),               // Prevent path traversal
	}
}
