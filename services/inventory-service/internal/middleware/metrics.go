package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	metricsOnce         sync.Once
)

func initMetrics() {
	metricsOnce.Do(func() {
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		)

		// Register with duplicate handling
		registerMetric(httpRequestsTotal)
		registerMetric(httpRequestDuration)
	})
}

var (
	StockLevelGauge         *prometheus.GaugeVec
	ReservationsActive      prometheus.Gauge
	ReservationExpiredTotal prometheus.Counter
	StockMovementsTotal     *prometheus.CounterVec
	DatabaseQueriesTotal    *prometheus.CounterVec
	DatabaseQueryDuration   *prometheus.HistogramVec
	grpcRequestsTotal       *prometheus.CounterVec
	grpcRequestDuration     *prometheus.HistogramVec
	businessMetricsOnce     sync.Once
)

func initBusinessMetrics() {
	businessMetricsOnce.Do(func() {
		StockLevelGauge = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "inventory_stock_available",
				Help: "Current available stock for products",
			},
			[]string{"product_id", "warehouse_id"},
		)

		ReservationsActive = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "inventory_reservations_active",
				Help: "Number of active reservations",
			},
		)

		ReservationExpiredTotal = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "inventory_reservation_expired_total",
				Help: "Total number of expired reservations",
			},
		)

		StockMovementsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_stock_movements_total",
				Help: "Total number of stock movements",
			},
			[]string{"movement_type", "product_id"},
		)

		DatabaseQueriesTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table"},
		)

		DatabaseQueryDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"operation", "table"},
		)

		// gRPC metrics
		grpcRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inventory_grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		)

		grpcRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "inventory_grpc_request_duration_seconds",
				Help:    "gRPC request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		)

		// Register all business metrics with duplicate handling
		registerMetric(StockLevelGauge)
		registerMetric(ReservationsActive)
		registerMetric(ReservationExpiredTotal)
		registerMetric(StockMovementsTotal)
		registerMetric(DatabaseQueriesTotal)
		registerMetric(DatabaseQueryDuration)
		registerMetric(grpcRequestsTotal)
		registerMetric(grpcRequestDuration)
	})
}

// registerMetric registers a metric and handles duplicates
func registerMetric(c prometheus.Collector) {
	if err := prometheus.Register(c); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
			// Only panic if it's not a duplicate registration error
			panic(err)
		}
		// If duplicate, silently ignore (metric already registered)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// PrometheusMiddleware records HTTP metrics for net/http
func PrometheusMiddleware(next http.Handler) http.Handler {
	initMetrics()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		path := r.URL.Path

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	})
}

// PrometheusHandler wraps a single http.HandlerFunc
func PrometheusHandler(next http.HandlerFunc) http.HandlerFunc {
	initMetrics()

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := newResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		path := r.URL.Path

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
	}
}

// RecordStockLevel updates stock level gauge
func RecordStockLevel(productID, warehouseID string, available int32) {
	initBusinessMetrics()
	StockLevelGauge.WithLabelValues(productID, warehouseID).Set(float64(available))
}

// RecordStockMovement increments stock movement counter
func RecordStockMovement(movementType, productID string) {
	initBusinessMetrics()
	StockMovementsTotal.WithLabelValues(movementType, productID).Inc()
}

// RecordDatabaseQuery records database operation metrics
func RecordDatabaseQuery(operation, table string, duration time.Duration) {
	initBusinessMetrics()
	DatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
	DatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordGRPCRequest records gRPC request metrics
func RecordGRPCRequest(method, status string, duration time.Duration) {
	initBusinessMetrics()
	grpcRequestsTotal.WithLabelValues(method, status).Inc()
	grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// PrometheusGinMiddleware records HTTP metrics for Gin framework
func PrometheusGinMiddleware() gin.HandlerFunc {
	initMetrics()

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
