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
			Name: "order_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Database operation metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_service_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Order-specific metrics
	ordersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_orders_total",
			Help: "Total number of orders created",
		},
		[]string{"status"},
	)

	orderValueTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_order_value_total",
			Help: "Total order value",
		},
		[]string{"currency"},
	)

	activeOrders = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "order_service_active_orders",
			Help: "Number of active orders by status",
		},
		[]string{"status"},
	)

	// Cart metrics
	cartOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_cart_operations_total",
			Help: "Total number of cart operations",
		},
		[]string{"operation", "status"},
	)

	// gRPC request metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_service_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Active connections
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "order_service_active_connections",
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

// RecordOrderCreated records a new order creation
func RecordOrderCreated(status string) {
	ordersTotal.WithLabelValues(status).Inc()
}

// RecordOrderValue records the total order value
func RecordOrderValue(currency string, value float64) {
	orderValueTotal.WithLabelValues(currency).Add(value)
}

// UpdateActiveOrders updates the number of active orders by status
func UpdateActiveOrders(status string, count float64) {
	activeOrders.WithLabelValues(status).Set(count)
}

// RecordCartOperation records a cart operation
func RecordCartOperation(operation, status string) {
	cartOperationsTotal.WithLabelValues(operation, status).Inc()
}
