package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu            sync.Mutex
	requests      map[string]*clientInfo
	limit         int
	windowSeconds int
}

type clientInfo struct {
	count     int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMin int) *RateLimiter {
	rl := &RateLimiter{
		requests:      make(map[string]*clientInfo),
		limit:         requestsPerMin,
		windowSeconds: 60,
	}

	// Cleanup expired entries every minute
	go rl.cleanup()

	return rl
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, info := range rl.requests {
			if now.Sub(info.lastReset) > time.Duration(rl.windowSeconds)*time.Second {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	info, exists := rl.requests[ip]
	if !exists {
		rl.requests[ip] = &clientInfo{
			count:     1,
			lastReset: now,
		}
		return true
	}

	// Reset counter if window has passed
	if now.Sub(info.lastReset) > time.Duration(rl.windowSeconds)*time.Second {
		info.count = 1
		info.lastReset = now
		return true
	}

	// Check if limit exceeded
	if info.count >= rl.limit {
		return false
	}

	info.count++
	return true
}

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitWithConfig creates rate limit middleware with configuration
func RateLimitWithConfig(requestsPerMin int, burstSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerMin)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !limiter.Allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
				"limit":   requestsPerMin,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
