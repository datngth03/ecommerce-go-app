// internal/recommendation/infrastructure/messaging/kafka_event_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"time" // Import time for time.RFC3339

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap" // Use zap for structured logging

	"github.com/datngth03/ecommerce-go-app/internal/recommendation/application"
	"github.com/datngth03/ecommerce-go-app/internal/shared/events"                          // Import shared events package
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                          // Import shared logger
	recommendation_client "github.com/datngth03/ecommerce-go-app/pkg/client/recommendation" // THÊM: Import Recommendation gRPC client
)

// KafkaProductEventConsumer defines the consumer for product events in Recommendation Service.
// KafkaProductEventConsumer định nghĩa consumer sự kiện sản phẩm trong Recommendation Service.
type KafkaProductEventConsumer struct {
	reader                *kafka.Reader
	recommendationService application.RecommendationService
	log                   *zap.Logger // Use the structured logger
}

// NewKafkaProductEventConsumer creates a new Kafka event consumer for product events.
// NewKafkaProductEventConsumer tạo một consumer sự kiện Kafka mới cho các sự kiện sản phẩm.
func NewKafkaProductEventConsumer(brokerAddr, topic, groupID string, recommendationService application.RecommendationService) *KafkaProductEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaProductEventConsumer{
		reader:                reader,
		recommendationService: recommendationService,
		log:                   logger.Logger, // Use the shared logger instance
	}
}

// StartConsuming starts consuming messages from Kafka.
// StartConsuming bắt đầu tiêu thụ tin nhắn từ Kafka.
func (c *KafkaProductEventConsumer) StartConsuming(ctx context.Context) {
	c.log.Info("Recommendation Service: Bắt đầu tiêu thụ sự kiện từ Kafka topic.",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("groupID", c.reader.Config().GroupID))

	for {
		// Fetch message from Kafka
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			// Check if context was cancelled (graceful shutdown)
			if ctx.Err() == context.Canceled {
				c.log.Info("Recommendation Service Consumer ngừng hoạt động do context bị hủy.")
				return
			}
			// Log other errors and continue
			c.log.Error("Recommendation Service Consumer: Lỗi khi đọc tin nhắn từ Kafka.", zap.Error(err))
			continue
		}

		c.log.Info("Recommendation Service Consumer: Đã nhận tin nhắn từ Kafka.",
			zap.String("topic", m.Topic),
			zap.Int("partition", m.Partition),
			zap.Int64("offset", m.Offset),
			zap.String("key", string(m.Key)))

		// Unmarshal the raw Kafka message value into our shared ProductEvent structure
		var event events.ProductEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			c.log.Error("Recommendation Service Consumer: Lỗi khi giải mã sự kiện sản phẩm.",
				zap.Error(err),
				zap.String("message_value", string(m.Value)))
			// Commit offset even if unmarshalling fails to avoid reprocessing bad messages
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi commit offset sau lỗi giải mã.", zap.Error(commitErr))
			}
			continue
		}

		c.log.Info("Recommendation Service Consumer: Đã giải mã sự kiện.", zap.String("event_type", event.Type), zap.String("aggregate_id", event.AggregateID))

		// Handle event based on its type
		switch event.Type {
		case "ProductCreated":
			// Unmarshal the payload into the specific ProductEventPayload structure
			var payload events.ProductEventPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi giải mã payload ProductCreated.", zap.Error(err), zap.String("payload_value", string(event.Payload)))
				break // Exit switch, don't process further
			}
			c.log.Info("Recommendation Service Consumer: Đã ghi lại tương tác 'product_created' cho Product ID.",
				zap.String("product_id", payload.ID),
				zap.String("product_name", payload.Name))

			// Call application service to record the interaction (e.g., as a view event)
			_, err := c.recommendationService.RecordInteraction(ctx, &recommendation_client.RecordInteractionRequest{
				UserId:    "system", // Interaction by system for product creation
				ProductId: payload.ID,
				EventType: "product_created", // Or "system_indexed"
				Timestamp: time.Now().Unix(),
			})
			if err != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi ghi lại tương tác 'product_created' cho sản phẩm.",
					zap.String("product_id", payload.ID),
					zap.Error(err))
				// Do NOT commit offset if processing failed, allow reprocessing
			} else {
				c.log.Info("Recommendation Service Consumer: Đã ghi lại tương tác 'product_created' cho sản phẩm thành công.",
					zap.String("product_id", payload.ID))
				// Commit offset only if processing was successful
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Recommendation Service Consumer: Lỗi khi commit offset cho sự kiện ProductCreated.", zap.Error(commitErr))
				}
			}

		case "ProductUpdated":
			var payload events.ProductEventPayload
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi giải mã payload ProductUpdated.", zap.Error(err), zap.String("payload_value", string(event.Payload)))
				break
			}
			c.log.Info("Recommendation Service Consumer: Đã ghi lại tương tác 'product_updated' cho Product ID.",
				zap.String("product_id", payload.ID),
				zap.String("product_name", payload.Name))

			// Example: Update product details in recommendation system if needed
			// For simplicity, we just record interaction
			_, err := c.recommendationService.RecordInteraction(ctx, &recommendation_client.RecordInteractionRequest{
				UserId:    "system",
				ProductId: payload.ID,
				EventType: "product_updated",
				Timestamp: time.Now().Unix(),
			})
			if err != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi ghi lại tương tác 'product_updated' cho sản phẩm.",
					zap.String("product_id", payload.ID),
					zap.Error(err))
			} else {
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Recommendation Service Consumer: Lỗi khi commit offset cho sự kiện ProductUpdated.", zap.Error(commitErr))
				}
			}

		case "ProductDeleted":
			c.log.Info("Recommendation Service Consumer: Đã nhận sự kiện ProductDeleted cho Product ID.",
				zap.String("product_id", event.AggregateID))
			// Example: Remove product from recommendation lists/cache if needed
			// For simplicity, we just record interaction
			_, err := c.recommendationService.RecordInteraction(ctx, &recommendation_client.RecordInteractionRequest{
				UserId:    "system",
				ProductId: event.AggregateID,
				EventType: "product_deleted",
				Timestamp: time.Now().Unix(),
			})
			if err != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi ghi lại tương tác 'product_deleted' cho sản phẩm.",
					zap.String("product_id", event.AggregateID),
					zap.Error(err))
			} else {
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Recommendation Service Consumer: Lỗi khi commit offset cho sự kiện ProductDeleted.", zap.Error(commitErr))
				}
			}

		default:
			c.log.Warn("Recommendation Service Consumer: Loại sự kiện sản phẩm không xác định hoặc không được xử lý.",
				zap.String("event_type", event.Type),
				zap.String("message_value", string(m.Value)))
			// Commit unknown message to avoid reprocessing indefinitely
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Recommendation Service Consumer: Lỗi khi commit offset cho sự kiện không xác định.", zap.Error(commitErr))
			}
		}
	}
}

// Close closes the Kafka reader.
// Close đóng Kafka reader.
func (c *KafkaProductEventConsumer) Close() error {
	c.log.Info("Đóng Kafka Product Event Consumer (Recommendation Service)...")
	return c.reader.Close()
}
