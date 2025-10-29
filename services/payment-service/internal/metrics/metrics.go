package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Database operation metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_service_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_service_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Payment-specific metrics
	paymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_service_payments_total",
			Help: "Total number of payments processed",
		},
		[]string{"method", "status"},
	)

	paymentAmountTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_service_payment_amount_total",
			Help: "Total payment amount processed",
		},
		[]string{"currency", "method"},
	)

	paymentDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_service_payment_duration_seconds",
			Help:    "Payment processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"method"},
	)

	refundsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_service_refunds_total",
			Help: "Total number of refunds processed",
		},
		[]string{"status"},
	)

	// gRPC request metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_service_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_service_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Active connections
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "payment_service_active_connections",
			Help: "Number of active connections",
		},
	)
)

// PrometheusGinMiddleware creates a Gin middleware for Prometheus metrics
func PrometheusGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		activeConnections.Inc()
		defer activeConnections.Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := string(rune(c.Writer.Status()/100)) + "xx"

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// RecordDBQuery records a database query metric
func RecordDBQuery(operation, table, status string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(operation, table, status).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordGRPCRequest records a gRPC request metric
func RecordGRPCRequest(method, status string, duration time.Duration) {
	grpcRequestsTotal.WithLabelValues(method, status).Inc()
	grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordPayment records a payment transaction
func RecordPayment(method, status string, amount float64, currency string, duration time.Duration) {
	paymentsTotal.WithLabelValues(method, status).Inc()
	paymentAmountTotal.WithLabelValues(currency, method).Add(amount)
	paymentDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordRefund records a refund transaction
func RecordRefund(status string) {
	refundsTotal.WithLabelValues(status).Inc()
}
