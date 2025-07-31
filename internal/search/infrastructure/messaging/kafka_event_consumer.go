// internal/search/infrastructure/messaging/kafka_event_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute" // THÊM: Import attribute
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap" // Use zap for structured logging

	"github.com/datngth03/ecommerce-go-app/internal/search/application"
	"github.com/datngth03/ecommerce-go-app/internal/shared/events"            // THÊM: Import shared events package
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"            // THÊM: Import shared logger
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Product gRPC client (nếu consumer cần gọi)
	search_client "github.com/datngth03/ecommerce-go-app/pkg/client/search"   // Generated Search gRPC client
)

// KafkaProductEventConsumer defines the consumer for product events.
// KafkaProductEventConsumer định nghĩa consumer cho các sự kiện sản phẩm.
type KafkaProductEventConsumer struct {
	reader        *kafka.Reader
	searchService application.SearchService
	productClient product_client.ProductServiceClient
	log           *zap.Logger                   // Use the structured logger
	propagator    propagation.TextMapPropagator // Propagator để trích xuất context
}

// NewKafkaProductEventConsumer creates a new Kafka event consumer for product events.
// NewKafkaProductEventConsumer tạo một consumer sự kiện Kafka mới cho các sự kiện sản phẩm.
func NewKafkaProductEventConsumer(brokerAddr, topic, groupID string, searchService application.SearchService, productClient product_client.ProductServiceClient) *KafkaProductEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaProductEventConsumer{
		reader:        reader,
		searchService: searchService,
		productClient: productClient,
		log:           logger.Logger,
		propagator:    otel.GetTextMapPropagator(), // Khởi tạo propagator
	}
}

// StartConsuming starts consuming messages from Kafka.
// StartConsuming bắt đầu tiêu thụ tin nhắn từ Kafka.
func (c *KafkaProductEventConsumer) StartConsuming(ctx context.Context) {
	c.log.Info("Search Service: Bắt đầu tiêu thụ sự kiện từ Kafka topic.",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("groupID", c.reader.Config().GroupID))

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled {
				c.log.Info("Consumer ngừng hoạt động do context bị hủy.")
				return
			}
			c.log.Error("Search Service Consumer: Lỗi khi đọc tin nhắn từ Kafka.", zap.Error(err))
			continue
		}

		c.log.Info("Search Service Consumer: Đã nhận tin nhắn từ Kafka.",
			zap.String("topic", m.Topic),
			zap.Int("partition", m.Partition),
			zap.Int64("offset", m.Offset),
			zap.String("key", string(m.Key)))

		// Trích xuất Trace Context từ Kafka message headers
		headers := make(map[string]string)
		for _, header := range m.Headers {
			headers[header.Key] = string(header.Value)
		}
		// Tạo một context mới với trace context đã trích xuất
		msgCtx := c.propagator.Extract(context.Background(), propagation.MapCarrier(headers))
		// Bắt đầu một span mới để xử lý tin nhắn Kafka
		// Tên span thường là "process <topic_name> message"
		_, span := otel.Tracer("kafka-consumer").Start(msgCtx, fmt.Sprintf("process %s message", m.Topic),
			trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End() // Đảm bảo span được đóng

		var event events.ProductEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			c.log.Error("Search Service Consumer: Lỗi khi giải mã sự kiện sản phẩm.",
				zap.Error(err),
				zap.String("message_value", string(m.Value)))
			span.RecordError(err) // Ghi lỗi vào span
			// Commit offset even if unmarshalling fails to avoid reprocessing bad messages
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Search Service Consumer: Lỗi khi commit offset sau lỗi giải mã.", zap.Error(commitErr))
			}
			continue
		}

		c.log.Info("Search Service Consumer: Đã giải mã sự kiện.", zap.String("event_type", event.Type), zap.String("aggregate_id", event.AggregateID))
		span.SetAttributes(attribute.String("event.type", event.Type), attribute.String("event.aggregate_id", event.AggregateID))

		var payload events.ProductEventPayload
		if len(event.Payload) > 0 {
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.log.Error("Search Service Consumer: Lỗi khi giải mã payload sản phẩm.", zap.Error(err), zap.String("payload_value", string(event.Payload)))
				span.RecordError(err)
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Search Service Consumer: Lỗi khi commit offset sau lỗi giải mã payload.", zap.Error(commitErr))
				}
				continue
			}
		}

		switch event.Type {
		case "ProductCreated", "ProductUpdated":
			c.log.Info("Search Service Consumer: Xử lý sự kiện ProductCreated/Updated cho Product ID.",
				zap.String("product_id", payload.ID),
				zap.String("product_name", payload.Name))

			req := &search_client.IndexProductRequest{
				Id:            payload.ID,
				Name:          payload.Name,
				Description:   payload.Description,
				Price:         payload.Price,
				CategoryId:    payload.CategoryID,
				ImageUrls:     payload.ImageURLs,
				StockQuantity: payload.StockQuantity,
				CreatedAt:     payload.CreatedAt.Format(time.RFC3339), // SỬA: Định dạng thời gian
				UpdatedAt:     payload.UpdatedAt.Format(time.RFC3339), // SỬA: Định dạng thời gian
			}
			// Truyền context đã được làm giàu với trace context
			_, err := c.searchService.IndexProduct(ctx, req)
			if err != nil {
				c.log.Error("Search Service Consumer: Lỗi khi lập chỉ mục/cập nhật sản phẩm trong Elasticsearch.",
					zap.String("product_id", payload.ID),
					zap.Error(err))
				span.RecordError(err)
				// Không commit offset để tin nhắn có thể được xử lý lại sau
			} else {
				c.log.Info("Search Service Consumer: Đã lập chỉ mục/cập nhật sản phẩm thành công trong Elasticsearch.",
					zap.String("product_id", payload.ID))
				// Commit offset chỉ khi xử lý thành công
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Search Service Consumer: Lỗi khi commit offset cho sự kiện %s.", zap.String("event_type", event.Type), zap.Error(commitErr))
				}
			}
		case "ProductDeleted":
			c.log.Info("Search Service Consumer: Đã nhận sự kiện ProductDeleted cho Product ID.",
				zap.String("product_id", event.AggregateID))

			// Truyền context đã được làm giàu với trace context
			_, err := c.searchService.DeleteProductFromIndex(ctx, &search_client.DeleteProductFromIndexRequest{ProductId: event.AggregateID})
			if err != nil {
				c.log.Error("Search Service Consumer: Lỗi khi xóa sản phẩm khỏi chỉ mục Elasticsearch.",
					zap.String("product_id", event.AggregateID),
					zap.Error(err))
				span.RecordError(err)
			} else {
				c.log.Info("Search Service Consumer: Đã xóa sản phẩm khỏi chỉ mục Elasticsearch thành công.",
					zap.String("product_id", event.AggregateID))
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Search Service Consumer: Lỗi khi commit offset cho sự kiện ProductDeleted.", zap.String("event_type", event.Type), zap.Error(commitErr))
				}
			}
		default:
			c.log.Warn("Search Service Consumer: Loại sự kiện sản phẩm không xác định hoặc không được xử lý.",
				zap.String("event_type", event.Type),
				zap.String("message_value", string(m.Value)))
			// Commit unknown message to avoid reprocessing indefinitely
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Search Service Consumer: Lỗi khi commit offset cho sự kiện không xác định.", zap.String("event_type", event.Type), zap.Error(commitErr))
			}
		}
	}
}

// Close closes the Kafka consumer reader.
func (c *KafkaProductEventConsumer) Close() error {
	c.log.Info("Đóng Kafka Product Event Consumer (Search Service)...")
	return c.reader.Close()
}
