// internal/shared/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto" // For auto-registering metrics
)

// Define common metrics for your microservices
// Đặt định nghĩa các metrics chung cho các microservice của bạn

// RequestCount is a counter for total HTTP/gRPC requests received.
// RequestCount là một counter cho tổng số yêu cầu HTTP/gRPC đã nhận.
var (
	RequestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"service", "path", "method", "code"})

	// RequestDuration is a histogram for HTTP/gRPC request latencies.
	// RequestDuration là một histogram cho độ trễ của yêu cầu HTTP/gRPC.
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latencies in seconds.",
		Buckets: prometheus.DefBuckets, // Default buckets (0.005, 0.01, ..., 10)
	}, []string{"service", "path", "method"})

	// Example: Database query duration
	// Ví dụ: Thời gian truy vấn cơ sở dữ liệu
	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Database query latencies in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "query", "status"})

	// Example: Kafka message production count
	// Ví dụ: Số lượng tin nhắn Kafka được sản xuất
	KafkaProducedMessages = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_produced_messages_total",
		Help: "Total number of Kafka messages produced.",
	}, []string{"service", "topic", "status"})

	// Example: Kafka message consumption count
	// Ví dụ: Số lượng tin nhắn Kafka đã tiêu thụ
	KafkaConsumedMessages = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_consumed_messages_total",
		Help: "Total number of Kafka messages consumed.",
	}, []string{"service", "topic", "status"})
)
