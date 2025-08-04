// internal/inventory/infrastructure/messaging/kafka_event_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute" // Import attribute
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap" // Use zap for structured logging

	"github.com/datngth03/ecommerce-go-app/internal/inventory/application"
	"github.com/datngth03/ecommerce-go-app/internal/shared/events"                // Import shared events package
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                // Import shared logger
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Generated Inventory gRPC client
)

// KafkaProductEventConsumer is responsible for consuming product events from Kafka for Inventory Service.
type KafkaProductEventConsumer struct {
	reader           *kafka.Reader
	inventoryService application.InventoryService
	log              *zap.Logger
	propagator       propagation.TextMapPropagator
}

// NewKafkaProductEventConsumer creates a new KafkaProductEventConsumer.
func NewKafkaProductEventConsumer(brokerAddr, topic, groupID string, inventoryService application.InventoryService) *KafkaProductEventConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddr},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		Dialer: &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		},
	})

	return &KafkaProductEventConsumer{
		reader:           reader,
		inventoryService: inventoryService,
		log:              logger.Logger,
		propagator:       otel.GetTextMapPropagator(),
	}
}

// StartConsuming starts consuming messages from Kafka.
func (c *KafkaProductEventConsumer) StartConsuming(ctx context.Context) {
	c.log.Info("Inventory Service Consumer: Bắt đầu tiêu thụ sự kiện từ Kafka topic.",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("groupID", c.reader.Config().GroupID))

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled {
				c.log.Info("Consumer ngừng hoạt động do context bị hủy.")
				return
			}
			c.log.Error("Inventory Service Consumer: Lỗi khi đọc tin nhắn từ Kafka.", zap.Error(err))
			continue
		}

		c.log.Info("Inventory Service Consumer: Đã nhận tin nhắn từ Kafka.",
			zap.String("topic", m.Topic),
			zap.Int("partition", m.Partition),
			zap.Int64("offset", m.Offset),
			zap.String("key", string(m.Key)))

		// Trích xuất Trace Context từ Kafka message headers
		headers := make(map[string]string)
		for _, header := range m.Headers {
			headers[header.Key] = string(header.Value)
		}
		msgCtx := c.propagator.Extract(context.Background(), propagation.MapCarrier(headers))
		_, span := otel.Tracer("kafka-consumer").Start(msgCtx, fmt.Sprintf("process %s message", m.Topic),
			trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End()

		var event events.ProductEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			c.log.Error("Inventory Service Consumer: Lỗi khi giải mã sự kiện sản phẩm.",
				zap.Error(err),
				zap.String("message_value", string(m.Value)))
			span.RecordError(err)
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Inventory Service Consumer: Lỗi khi commit offset sau lỗi giải mã.", zap.Error(commitErr))
			}
			continue
		}

		c.log.Info("Inventory Service Consumer: Đã giải mã sự kiện.", zap.String("event_type", event.Type), zap.String("aggregate_id", event.AggregateID))
		span.SetAttributes(attribute.String("event.type", event.Type), attribute.String("event.aggregate_id", event.AggregateID))

		var payload events.ProductEventPayload
		if len(event.Payload) > 0 {
			if err := json.Unmarshal(event.Payload, &payload); err != nil {
				c.log.Error("Inventory Service Consumer: Lỗi khi giải mã payload sản phẩm.", zap.Error(err), zap.String("payload_value", string(event.Payload)))
				span.RecordError(err)
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Inventory Service Consumer: Lỗi khi commit offset sau lỗi giải mã payload.", zap.Error(commitErr))
				}
				continue
			}
		}

		// Tạo context mới cho các thao tác xuống tầng ứng dụng
		// appCtx := span.Context()

		switch event.Type {
		case "ProductCreated":
			c.log.Info("Inventory Service Consumer: Xử lý sự kiện ProductCreated cho Product ID.",
				zap.String("product_id", payload.ID),
				zap.String("product_name", payload.Name))
			setStockReq := &inventory_client.SetStockRequest{
				ProductId: payload.ID,
				Quantity:  0,
			}
			_, err := c.inventoryService.SetStock(ctx, setStockReq) // Dùng appCtx
			if err != nil {
				c.log.Error("Inventory Service Consumer: Lỗi khi khởi tạo tồn kho cho sản phẩm.",
					zap.String("product_id", payload.ID),
					zap.Error(err))
				span.RecordError(err)
			} else {
				c.log.Info("Inventory Service Consumer: Đã khởi tạo tồn kho cho sản phẩm thành công.",
					zap.String("product_id", payload.ID))
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					c.log.Error("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện ProductCreated.", zap.String("event_type", event.Type), zap.Error(commitErr))
				}
			}
		case "ProductUpdated":
			c.log.Info("Inventory Service Consumer: Xử lý sự kiện ProductUpdated cho Product ID.",
				zap.String("product_id", payload.ID),
				zap.String("product_name", payload.Name))
			// Hiện tại không có logic cập nhật tồn kho trực tiếp từ sự kiện này.
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện ProductUpdated.", zap.String("event_type", event.Type), zap.Error(commitErr))
			}
		case "ProductDeleted":
			c.log.Info("Inventory Service Consumer: Xử lý sự kiện ProductDeleted cho Product ID.",
				zap.String("product_id", event.AggregateID))
			// deleteItemReq := &inventory_client.DeleteItemRequest{
			// 	ProductId: event.AggregateID,
			// }
			// // THÊM: Truyền context đã được làm giàu với trace context
			// _, err := c.inventoryService.DeleteItem(appCtx, deleteItemReq) // Dùng appCtx
			// if err != nil {
			// 	c.log.Error("Inventory Service Consumer: Lỗi khi xóa tồn kho cho sản phẩm.",
			// 		zap.String("product_id", event.AggregateID),
			// 		zap.Error(err))
			// 	span.RecordError(err)
			// } else {
			// 	c.log.Info("Inventory Service Consumer: Đã xóa tồn kho cho sản phẩm thành công.",
			// 		zap.String("product_id", event.AggregateID))
			// 	if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
			// 		c.log.Error("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện ProductDeleted.", zap.String("event_type", event.Type), zap.Error(commitErr))
			// 	}
			// }
		default:
			c.log.Warn("Inventory Service Consumer: Loại sự kiện sản phẩm không xác định hoặc không được xử lý.",
				zap.String("event_type", event.Type),
				zap.String("message_value", string(m.Value)))
			if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
				c.log.Error("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện không xác định.", zap.String("event_type", event.Type), zap.Error(commitErr))
			}
		}
	}
}

// Close closes the Kafka consumer reader.
func (c *KafkaProductEventConsumer) Close() error {
	c.log.Info("Đóng Kafka Product Event Consumer (Inventory Service)...")
	return c.reader.Close()
}
