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

	// Product-specific metrics
	productsTotal   prometheus.Gauge
	categoriesTotal prometheus.Gauge

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
				Name: "product_service_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "product_service_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		)

		// Database operation metrics
		dbQueriesTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "product_service_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		)

		dbQueryDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "product_service_db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		)

		// Product-specific metrics
		productsTotal = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "product_service_products_total",
				Help: "Total number of products in the system",
			},
		)

		categoriesTotal = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "product_service_categories_total",
				Help: "Total number of categories in the system",
			},
		)

		// gRPC request metrics
		grpcRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "product_service_grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		)

		grpcRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "product_service_grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		)

		// Active connections
		activeConnections = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "product_service_active_connections",
				Help: "Number of active connections",
			},
		)

		// Register all metrics (with duplicate check)
		metrics := []prometheus.Collector{
			httpRequestsTotal,
			httpRequestDuration,
			dbQueriesTotal,
			dbQueryDuration,
			productsTotal,
			categoriesTotal,
			grpcRequestsTotal,
			grpcRequestDuration,
			activeConnections,
		}

		for _, metric := range metrics {
			if err := prometheus.Register(metric); err != nil {
				// If already registered, use existing collector
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					// Type assertion to reuse existing metric
					switch metric.(type) {
					case *prometheus.CounterVec:
						if existing, ok := are.ExistingCollector.(*prometheus.CounterVec); ok {
							_ = existing // Use existing
						}
					case *prometheus.HistogramVec:
						if existing, ok := are.ExistingCollector.(*prometheus.HistogramVec); ok {
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
	// Initialize metrics once
	initMetrics()

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
	initMetrics() // Ensure metrics are initialized
	dbQueriesTotal.WithLabelValues(operation, table, status).Inc()
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordGRPCRequest records a gRPC request metric
func RecordGRPCRequest(method, status string, duration time.Duration) {
	initMetrics() // Ensure metrics are initialized
	grpcRequestsTotal.WithLabelValues(method, status).Inc()
	grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// UpdateProductsTotal updates the total number of products gauge
func UpdateProductsTotal(count float64) {
	initMetrics() // Ensure metrics are initialized
	productsTotal.Set(count)
}

// UpdateCategoriesTotal updates the total number of categories gauge
func UpdateCategoriesTotal(count float64) {
	initMetrics() // Ensure metrics are initialized
	categoriesTotal.Set(count)
}
