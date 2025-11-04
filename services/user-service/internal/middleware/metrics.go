// services/user-service/internal/middleware/metrics.go
package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	metricsOnce         sync.Once
	metricsRegistered   bool
)

func initMetrics() {
	metricsOnce.Do(func() {
		// Only create metrics if not already registered
		if !metricsRegistered {
			httpRequestsTotal = prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "user_http_requests_total",
					Help: "Total number of HTTP requests",
				},
				[]string{"method", "path", "status"},
			)

			httpRequestDuration = prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "user_http_request_duration_seconds",
					Help:    "HTTP request latency in seconds",
					Buckets: prometheus.DefBuckets,
				},
				[]string{"method", "path"},
			)

			// Try to register metrics, ignore if already registered
			if err := prometheus.Register(httpRequestsTotal); err != nil {
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					httpRequestsTotal = are.ExistingCollector.(*prometheus.CounterVec)
				}
			}

			if err := prometheus.Register(httpRequestDuration); err != nil {
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					httpRequestDuration = are.ExistingCollector.(*prometheus.HistogramVec)
				}
			}

			metricsRegistered = true
		}
	})
}

// PrometheusMiddleware records HTTP metrics for Gin
func PrometheusMiddleware() gin.HandlerFunc {
	// Initialize metrics once
	initMetrics()

	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		path := c.Request.URL.Path

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	})
}
