// internal/product/infrastructure/messaging/kafka_event_publisher.go
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go" // Kafka client library

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// ProductEvent represents a generic product event structure.
// ProductEvent đại diện cho một cấu trúc sự kiện sản phẩm chung.
type ProductEvent struct {
	Type        string          `json:"type"`         // Ví dụ: "ProductCreated", "ProductUpdated", "ProductDeleted"
	Timestamp   string          `json:"timestamp"`    // Thời gian xảy ra sự kiện
	Payload     json.RawMessage `json:"payload"`      // Dữ liệu sản phẩm hoặc dữ liệu liên quan
	AggregateID string          `json:"aggregate_id"` // Product ID
}

// ProductEventPublisher defines the interface for publishing product events.
// ProductEventPublisher định nghĩa giao diện để phát các sự kiện sản phẩm.
type ProductEventPublisher interface {
	PublishProductCreated(ctx context.Context, product *domain.Product) error
	PublishProductUpdated(ctx context.Context, product *domain.Product) error
	PublishProductDeleted(ctx context.Context, productID string) error
	Close() error
}

// kafkaProductEventPublisher implements ProductEventPublisher using Kafka.
// kafkaProductEventPublisher triển khai ProductEventPublisher bằng Kafka.
type kafkaProductEventPublisher struct {
	writer *kafka.Writer
}

// NewKafkaProductEventPublisher creates a new KafkaProductEventPublisher.
// NewKafkaProductEventPublisher tạo một KafkaProductEventPublisher mới.
func NewKafkaProductEventPublisher(kafkaBroker string, topic string) *kafkaProductEventPublisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(kafkaBroker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{}, // Chọn broker có ít byte nhất đang chờ xử lý
		Logger:   kafka.LoggerFunc(log.Printf),
		// Có thể cấu hình thêm các tùy chọn như BatchSize, Linger, requiredACKS
	}
	log.Printf("Kafka Product Event Publisher initialized for topic %s on broker %s", topic, kafkaBroker)
	return &kafkaProductEventPublisher{writer: writer}
}

// PublishProductCreated publishes a ProductCreated event to Kafka.
// PublishProductCreated phát một sự kiện ProductCreated tới Kafka.
func (p *kafkaProductEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product for created event: %w", err)
	}

	event := ProductEvent{
		Type:        "ProductCreated",
		Timestamp:   time.Now().Format(time.RFC3339),
		Payload:     productJSON,
		AggregateID: product.ID,
	}
	return p.publishEvent(ctx, event)
}

// PublishProductUpdated publishes a ProductUpdated event to Kafka.
// PublishProductUpdated phát một sự kiện ProductUpdated tới Kafka.
func (p *kafkaProductEventPublisher) PublishProductUpdated(ctx context.Context, product *domain.Product) error {
	productJSON, err := json.Marshal(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product for updated event: %w", err)
	}

	event := ProductEvent{
		Type:        "ProductUpdated",
		Timestamp:   time.Now().Format(time.RFC3339),
		Payload:     productJSON,
		AggregateID: product.ID,
	}
	return p.publishEvent(ctx, event)
}

// PublishProductDeleted publishes a ProductDeleted event to Kafka.
// PublishProductDeleted phát một sự kiện ProductDeleted tới Kafka.
func (p *kafkaProductEventPublisher) PublishProductDeleted(ctx context.Context, productID string) error {
	payload := map[string]string{"product_id": productID}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal product ID for deleted event: %w", err)
	}

	event := ProductEvent{
		Type:        "ProductDeleted",
		Timestamp:   time.Now().Format(time.RFC3339),
		Payload:     payloadJSON,
		AggregateID: productID,
	}
	return p.publishEvent(ctx, event)
}

// publishEvent sends the generic event to Kafka.
// publishEvent gửi sự kiện chung tới Kafka.
func (p *kafkaProductEventPublisher) publishEvent(ctx context.Context, event ProductEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.AggregateID), // Sử dụng Product ID làm key để đảm bảo thứ tự trong cùng một partition
		Value: eventJSON,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.Type)},
		},
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	log.Printf("Published %s event for Product ID: %s", event.Type, event.AggregateID)
	return nil
}

// Close closes the Kafka writer connection.
// Close đóng kết nối Kafka writer.
func (p *kafkaProductEventPublisher) Close() error {
	if p.writer != nil {
		log.Println("Closing Kafka Product Event Publisher...")
		return p.writer.Close()
	}
	return nil
}
