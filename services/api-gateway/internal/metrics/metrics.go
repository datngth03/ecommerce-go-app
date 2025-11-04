package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTP metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// gRPC client metrics
	grpcClientRequestsTotal   *prometheus.CounterVec
	grpcClientRequestDuration *prometheus.HistogramVec

	// Proxy metrics
	proxyRequestsTotal   *prometheus.CounterVec
	proxyRequestDuration *prometheus.HistogramVec

	// Authentication metrics
	authRequestsTotal *prometheus.CounterVec
	authFailuresTotal prometheus.Counter

	// Rate limit metrics
	rateLimitExceededTotal *prometheus.CounterVec

	// Active connections
	activeConnections prometheus.Gauge

	// Ensure metrics are initialized only once
	metricsOnce sync.Once
)

// initMetrics initializes all Prometheus metrics once
func initMetrics() {
	metricsOnce.Do(func() {
		// HTTP metrics
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_gateway_http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)

		// gRPC client metrics
		grpcClientRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_grpc_client_requests_total",
				Help: "Total number of gRPC client requests",
			},
			[]string{"service", "method", "status"},
		)

		grpcClientRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_gateway_grpc_client_request_duration_seconds",
				Help:    "gRPC client request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method"},
		)

		// Proxy metrics
		proxyRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_proxy_requests_total",
				Help: "Total number of proxied requests",
			},
			[]string{"service", "status"},
		)

		proxyRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_gateway_proxy_request_duration_seconds",
				Help:    "Proxied request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service"},
		)

		// Authentication metrics
		authRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_auth_requests_total",
				Help: "Total number of authentication requests",
			},
			[]string{"endpoint", "status"},
		)

		authFailuresTotal = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "api_gateway_auth_failures_total",
				Help: "Total number of authentication failures",
			},
		)

		// Rate limit metrics
		rateLimitExceededTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_gateway_rate_limit_exceeded_total",
				Help: "Total number of rate limit exceeded events",
			},
			[]string{"ip"},
		)

		// Active connections
		activeConnections = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "api_gateway_active_connections",
				Help: "Number of active connections",
			},
		)

		// Register all metrics with duplicate handling
		metrics := []prometheus.Collector{
			httpRequestsTotal,
			httpRequestDuration,
			grpcClientRequestsTotal,
			grpcClientRequestDuration,
			proxyRequestsTotal,
			proxyRequestDuration,
			authRequestsTotal,
			authFailuresTotal,
			rateLimitExceededTotal,
			activeConnections,
		}

		for _, metric := range metrics {
			if err := prometheus.Register(metric); err != nil {
				// Silently ignore duplicate registration errors
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					panic(err)
				}
			}
		}
	})
}

// PrometheusMiddleware records HTTP metrics for Gin framework
func PrometheusMiddleware() gin.HandlerFunc {
	initMetrics() // Initialize once

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
	initMetrics()
	grpcClientRequestsTotal.WithLabelValues(service, method, status).Inc()
	grpcClientRequestDuration.WithLabelValues(service, method).Observe(duration.Seconds())
}

// RecordProxyRequest records proxy request metrics
func RecordProxyRequest(service, status string, duration time.Duration) {
	initMetrics()
	proxyRequestsTotal.WithLabelValues(service, status).Inc()
	proxyRequestDuration.WithLabelValues(service).Observe(duration.Seconds())
}

// RecordAuthRequest records authentication request metrics
func RecordAuthRequest(endpoint, status string) {
	initMetrics()
	authRequestsTotal.WithLabelValues(endpoint, status).Inc()
}

// RecordAuthFailure increments authentication failure counter
func RecordAuthFailure() {
	initMetrics()
	authFailuresTotal.Inc()
}

// RecordRateLimitExceeded increments rate limit exceeded counter
func RecordRateLimitExceeded(ip string) {
	initMetrics()
	rateLimitExceededTotal.WithLabelValues(ip).Inc()
}
