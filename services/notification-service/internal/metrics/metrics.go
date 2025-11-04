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

	// Notification-specific metrics
	notificationsSentTotal     *prometheus.CounterVec
	notificationDuration       *prometheus.HistogramVec
	emailsSentTotal            *prometheus.CounterVec
	smsSentTotal               *prometheus.CounterVec
	pushNotificationsSentTotal *prometheus.CounterVec
	notificationQueueSize      prometheus.Gauge

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
				Name: "notification_service_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		)

		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "notification_service_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		)

		// Database operation metrics
		dbQueriesTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		)

		dbQueryDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "notification_service_db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		)

		// Notification-specific metrics
		notificationsSentTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_notifications_sent_total",
				Help: "Total number of notifications sent",
			},
			[]string{"type", "status"},
		)

		notificationDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "notification_service_notification_duration_seconds",
				Help:    "Notification sending duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"type"},
		)

		emailsSentTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_emails_sent_total",
				Help: "Total number of emails sent",
			},
			[]string{"status"},
		)

		smsSentTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_sms_sent_total",
				Help: "Total number of SMS sent",
			},
			[]string{"status"},
		)

		pushNotificationsSentTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_push_notifications_sent_total",
				Help: "Total number of push notifications sent",
			},
			[]string{"status"},
		)

		notificationQueueSize = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "notification_service_queue_size",
				Help: "Current size of notification queue",
			},
		)

		// gRPC request metrics
		grpcRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "notification_service_grpc_requests_total",
				Help: "Total number of gRPC requests",
			},
			[]string{"method", "status"},
		)

		grpcRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "notification_service_grpc_request_duration_seconds",
				Help:    "gRPC request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		)

		// Active connections
		activeConnections = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "notification_service_active_connections",
				Help: "Number of active connections",
			},
		)

		// Register all metrics with duplicate handling
		metrics := []prometheus.Collector{
			httpRequestsTotal,
			httpRequestDuration,
			dbQueriesTotal,
			dbQueryDuration,
			notificationsSentTotal,
			notificationDuration,
			emailsSentTotal,
			smsSentTotal,
			pushNotificationsSentTotal,
			notificationQueueSize,
			grpcRequestsTotal,
			grpcRequestDuration,
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

// RecordNotificationSent records a notification sent
func RecordNotificationSent(notifType, status string, duration time.Duration) {
	initMetrics()
	notificationsSentTotal.WithLabelValues(notifType, status).Inc()
	notificationDuration.WithLabelValues(notifType).Observe(duration.Seconds())
}

// RecordEmailSent records an email sent
func RecordEmailSent(status string) {
	initMetrics()
	emailsSentTotal.WithLabelValues(status).Inc()
}

// RecordSMSSent records an SMS sent
func RecordSMSSent(status string) {
	initMetrics()
	smsSentTotal.WithLabelValues(status).Inc()
}

// RecordPushNotificationSent records a push notification sent
func RecordPushNotificationSent(status string) {
	initMetrics()
	pushNotificationsSentTotal.WithLabelValues(status).Inc()
}

// UpdateQueueSize updates the notification queue size
func UpdateQueueSize(size float64) {
	initMetrics()
	notificationQueueSize.Set(size)
}
