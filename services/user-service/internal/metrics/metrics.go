package metrics

import (
	"time"

	// "github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	userHttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	userHttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	UserActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "user_active_sessions",
			Help: "Number of active user sessions",
		},
	)

	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	UserLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user logins",
		},
		[]string{"status"},
	)

	UserDatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	UserDatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"operation", "table"},
	)

	// gRPC metrics
	userGrpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	userGrpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_grpc_request_duration_seconds",
			Help:    "gRPC request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

// RecordUserRegistration increments user registration counter
func RecordUserRegistration() {
	UserRegistrationsTotal.Inc()
}

// RecordUserLogin records login attempts
func RecordUserLogin(status string) {
	UserLoginsTotal.WithLabelValues(status).Inc()
}

// RecordDatabaseQuery records database operation metrics
func RecordDatabaseQuery(operation, table string, duration time.Duration) {
	UserDatabaseQueriesTotal.WithLabelValues(operation, table).Inc()
	UserDatabaseQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// RecordGRPCRequest records gRPC request metrics
func RecordGRPCRequest(method, status string, duration time.Duration) {
	userGrpcRequestsTotal.WithLabelValues(method, status).Inc()
	userGrpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// PrometheusMiddleware records HTTP metrics for Gin
// func PrometheusMiddleware() gin.HandlerFunc {
// 	return gin.HandlerFunc(func(c *gin.Context) {
// 		start := time.Now()

// 		c.Next()

// 		duration := time.Since(start).Seconds()
// 		status := c.Writer.Status()
// 		path := c.Request.URL.Path

// 		httpRequestsTotal.WithLabelValues(c.Request.Method, path, string(rune(status))).Inc()
// 		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
// 	})
// }
