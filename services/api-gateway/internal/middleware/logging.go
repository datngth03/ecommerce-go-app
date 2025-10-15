package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware logs all incoming requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log request details
		log.Printf("[%s] %s %s | Status: %d | Duration: %v | IP: %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("Error: %v", err.Error())
			}
		}
	}
}

// ErrorLogger logs errors with context
func ErrorLogger(c *gin.Context, err error) {
	log.Printf("[ERROR] Path: %s | Method: %s | IP: %s | Error: %v",
		c.Request.URL.Path,
		c.Request.Method,
		c.ClientIP(),
		err,
	)
}
