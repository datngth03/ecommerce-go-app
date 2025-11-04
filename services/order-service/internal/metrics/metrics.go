package metrics

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HTTP request metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec

	// Database operation metrics
	dbQueriesTotal  *prometheus.CounterVec
	dbQueryDuration *prometheus.HistogramVec

	// Order-specific metrics
	ordersTotal     *prometheus.CounterVec
	orderValueTotal *prometheus.CounterVec
	activeOrders    *prometheus.GaugeVec

	// Cart metrics
	cartOperationsTotal *prometheus.CounterVec

	// gRPC request metrics
	grpcRequestsTotal   *prometheus.CounterVec
	grpcRequestDuration *prometheus.HistogramVec

	// Active connections
	activeConnections prometheus.Gauge

	// Ensure metrics are initialized only once
	metricsOnce sync.Once
)

// initMetrics initializes all Prometheus metrics once
func initMetrics() {
	metricsOnce.Do(func() {
		// HTTP request metrics
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_service_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		)

		// Database operation metrics
		dbQueriesTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		)

		dbQueryDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_service_db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		)

		// Order-specific metrics
		ordersTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_orders_total",
				Help: "Total number of orders created",
			},
			[]string{"status"},
		)

		orderValueTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_order_value_total",
				Help: "Total order value",
			},
			[]string{"currency"},
		)

		activeOrders = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "order_service_active_orders",
				Help: "Number of active orders by status",
			},
			[]string{"status"},
		)

		// Cart metrics
		cartOperationsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_cart_operations_total",
				Help: "Total number of cart operations",
			},
			[]string{"operation", "status"},
		)

		// gRPC request metrics
		grpcRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "order_service_grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		)

		grpcRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "order_service_grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		)

		// Active connections
		activeConnections = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "order_service_active_connections",
				Help: "Number of active connections",
			},
		)

		// Register all metrics (with duplicate check)
		metrics := []prometheus.Collector{
			httpRequestsTotal,
			httpRequestDuration,
			dbQueriesTotal,
			dbQueryDuration,
			ordersTotal,
			orderValueTotal,
			activeOrders,
			cartOperationsTotal,
			grpcRequestsTotal,
			grpcRequestDuration,
			activeConnections,
		}

		for _, metric := range metrics {
			if err := prometheus.Register(metric); err != nil {
				// If already registered, use existing collector
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					switch metric.(type) {
					case *prometheus.CounterVec:
						if existing, ok := are.ExistingCollector.(*prometheus.CounterVec); ok {
							_ = existing
						}
					case *prometheus.HistogramVec:
						if existing, ok := are.ExistingCollector.(*prometheus.HistogramVec); ok {
							_ = existing
						}
					case *prometheus.GaugeVec:
						if existing, ok := are.ExistingCollector.(*prometheus.GaugeVec); ok {
							_ = existing
						}
					case prometheus.Gauge:
						if existing, ok := are.ExistingCollector.(prometheus.Gauge); ok {
							_ = existing
						}
					}
				}
			}
		}
	})
}

// PrometheusGinMiddleware creates a Gin middleware for Prometheus metrics
func PrometheusGinMiddleware() gin.HandlerFunc {
	initMetrics() // Initialize once

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
	initMetrics()
	dbQueriesTotal.WithLabelValues(operation, table, status).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordGRPCRequest records a gRPC request metric
func RecordGRPCRequest(method, status string, duration time.Duration) {
	initMetrics()
	grpcRequestsTotal.WithLabelValues(method, status).Inc()
	grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordOrderCreated records a new order creation
func RecordOrderCreated(status string) {
	initMetrics()
	ordersTotal.WithLabelValues(status).Inc()
}

// RecordOrderValue records the total order value
func RecordOrderValue(currency string, value float64) {
	initMetrics()
	orderValueTotal.WithLabelValues(currency).Add(value)
}

// UpdateActiveOrders updates the number of active orders by status
func UpdateActiveOrders(status string, count float64) {
	initMetrics()
	activeOrders.WithLabelValues(status).Set(count)
}

// RecordCartOperation records a cart operation
func RecordCartOperation(operation, status string) {
	initMetrics()
	cartOperationsTotal.WithLabelValues(operation, status).Inc()
}
