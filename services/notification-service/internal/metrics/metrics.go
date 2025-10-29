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
			Name: "notification_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_service_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Database operation metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_service_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Notification-specific metrics
	notificationsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_notifications_sent_total",
			Help: "Total number of notifications sent",
		},
		[]string{"type", "status"},
	)

	notificationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_service_notification_duration_seconds",
			Help:    "Notification sending duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"type"},
	)

	emailsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_emails_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"status"},
	)

	smsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_sms_sent_total",
			Help: "Total number of SMS sent",
		},
		[]string{"status"},
	)

	pushNotificationsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_push_notifications_sent_total",
			Help: "Total number of push notifications sent",
		},
		[]string{"status"},
	)

	notificationQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_service_queue_size",
			Help: "Current size of notification queue",
		},
	)

	// gRPC request metrics
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_service_grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_service_grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Active connections
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "notification_service_active_connections",
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

// RecordNotificationSent records a notification sent
func RecordNotificationSent(notifType, status string, duration time.Duration) {
	notificationsSentTotal.WithLabelValues(notifType, status).Inc()
	notificationDuration.WithLabelValues(notifType).Observe(duration.Seconds())
}

// RecordEmailSent records an email sent
func RecordEmailSent(status string) {
	emailsSentTotal.WithLabelValues(status).Inc()
}

// RecordSMSSent records an SMS sent
func RecordSMSSent(status string) {
	smsSentTotal.WithLabelValues(status).Inc()
}

// RecordPushNotificationSent records a push notification sent
func RecordPushNotificationSent(status string) {
	pushNotificationsSentTotal.WithLabelValues(status).Inc()
}

// UpdateQueueSize updates the notification queue size
func UpdateQueueSize(size float64) {
	notificationQueueSize.Set(size)
}
