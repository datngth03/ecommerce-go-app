package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// gRPC client metrics
	grpcClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_grpc_client_requests_total",
			Help: "Total number of gRPC client requests",
		},
		[]string{"service", "method", "status"},
	)

	grpcClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_grpc_client_request_duration_seconds",
			Help:    "gRPC client request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	// Proxy metrics
	proxyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_proxy_requests_total",
			Help: "Total number of proxied requests",
		},
		[]string{"service", "status"},
	)

	proxyRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_proxy_request_duration_seconds",
			Help:    "Proxied request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service"},
	)

	// Authentication metrics
	authRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_auth_requests_total",
			Help: "Total number of authentication requests",
		},
		[]string{"endpoint", "status"},
	)

	authFailuresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "api_gateway_auth_failures_total",
			Help: "Total number of authentication failures",
		},
	)

	// Rate limit metrics
	rateLimitExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"ip"},
	)

	// Active connections
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "api_gateway_active_connections",
			Help: "Number of active connections",
		},
	)
)

// PrometheusMiddleware records HTTP metrics for Gin framework
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		activeConnections.Inc()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
		activeConnections.Dec()
	}
}

// RecordGRPCClientRequest records gRPC client request metrics
func RecordGRPCClientRequest(service, method, status string, duration time.Duration) {
	grpcClientRequestsTotal.WithLabelValues(service, method, status).Inc()
	grpcClientRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// RecordProxyRequest records proxy request metrics
func RecordProxyRequest(service, status string, duration time.Duration) {
	proxyRequestsTotal.WithLabelValues(service, status).Inc()
	proxyRequestDuration.WithLabelValues(service).Observe(duration.Seconds())
}

// RecordAuthRequest records authentication request metrics
func RecordAuthRequest(endpoint, status string) {
	authRequestsTotal.WithLabelValues(endpoint, status).Inc()
}

// RecordAuthFailure increments authentication failure counter
func RecordAuthFailure() {
	authFailuresTotal.Inc()
}

// RecordRateLimitExceeded increments rate limit exceeded counter
func RecordRateLimitExceeded(ip string) {
	rateLimitExceededTotal.WithLabelValues(ip).Inc()
}
