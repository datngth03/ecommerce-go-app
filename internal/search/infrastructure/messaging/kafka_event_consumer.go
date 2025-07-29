// internal/search/infrastructure/messaging/kafka_event_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"log" // Using standard log for simplicity, consider zap/logrus for production
	"time"

	"github.com/segmentio/kafka-go" // Kafka Go client

	"github.com/datngth03/ecommerce-go-app/internal/search/application"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Import Product gRPC client (To fetch full product details if needed, but not directly used for unmarshalling ProductEvent)
	search_client "github.com/datngth03/ecommerce-go-app/pkg/client/search"   // Import Search gRPC client (cho IndexProductRequest)
)

// ProductEventPayload mirrors the structure of Product from Product Service's domain
// to enable unmarshalling of Kafka events without cross-package internal dependency.
// ProductEventPayload phản ánh cấu trúc của Product từ domain của Product Service
// để cho phép giải mã các sự kiện Kafka mà không cần phụ thuộc chéo vào gói nội bộ.
type ProductEventPayload struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Price         float64   `json:"price"`
	CategoryID    string    `json:"category_id"`
	ImageURLs     []string  `json:"image_urls,omitempty"`
	StockQuantity int32     `json:"stock_quantity,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProductEvent defines the structure of product events.
// This should match the event structure published by Product Service.
// ProductEvent định nghĩa cấu trúc của các sự kiện sản phẩm.
// Cấu trúc này phải khớp với cấu trúc sự kiện được phát bởi Product Service.
type ProductEvent struct {
	Type    string               `json:"type"`              // e.g., "ProductCreated", "ProductUpdated", "ProductDeleted"
	Product *ProductEventPayload `json:"product,omitempty"` // Sử dụng struct cục bộ
	ID      string               `json:"id,omitempty"`      // For ProductDeleted event
}

// KafkaProductEventConsumer defines the consumer for product events.
// KafkaProductEventConsumer định nghĩa consumer cho các sự kiện sản phẩm.
type KafkaProductEventConsumer struct {
	reader        *kafka.Reader
	searchService application.SearchService
	productClient product_client.ProductServiceClient // To fetch full product details if needed
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
	}
}

// StartConsuming starts consuming messages from Kafka.
// StartConsuming bắt đầu tiêu thụ tin nhắn từ Kafka.
func (c *KafkaProductEventConsumer) StartConsuming(ctx context.Context) {
	log.Printf("Bắt đầu tiêu thụ sự kiện từ Kafka topic '%s'...", c.reader.Config().Topic)

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled {
				log.Println("Consumer ngừng hoạt động do context bị hủy.")
				return // Context canceled, gracefully exit
			}
			log.Printf("Lỗi khi đọc tin nhắn từ Kafka: %v", err)
			continue
		}

		log.Printf("Đã nhận tin nhắn từ Kafka: topic=%s, partition=%d, offset=%d, key=%s",
			m.Topic, m.Partition, m.Offset, string(m.Key))

		var event ProductEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("Lỗi khi giải mã sự kiện sản phẩm: %v, tin nhắn: %s", err, string(m.Value))
			// Commit offset even if unmarshalling fails to avoid reprocessing bad messages
			if err := c.reader.CommitMessages(ctx, m); err != nil {
				log.Printf("Lỗi khi commit offset sau lỗi giải mã: %v", err)
			}
			continue
		}

		log.Printf("Đã giải mã sự kiện loại: %s", event.Type)

		switch event.Type {
		case "ProductCreated", "ProductUpdated":
			if event.Product == nil {
				log.Printf("Bỏ qua sự kiện %s: không có dữ liệu sản phẩm", event.Type)
				break
			}
			// Index or update product in Elasticsearch
			// Chuyển đổi ProductEventPayload sang search_client.IndexProductRequest
			req := &search_client.IndexProductRequest{
				Id:            event.Product.ID,
				Name:          event.Product.Name,
				Description:   event.Product.Description,
				Price:         event.Product.Price,
				CategoryId:    event.Product.CategoryID,
				ImageUrls:     event.Product.ImageURLs,
				StockQuantity: event.Product.StockQuantity,
				CreatedAt:     event.Product.CreatedAt.Format(time.RFC3339),
				UpdatedAt:     event.Product.UpdatedAt.Format(time.RFC3339),
			}
			_, err := c.searchService.IndexProduct(ctx, req)
			if err != nil {
				log.Printf("Lỗi khi lập chỉ mục/cập nhật sản phẩm %s trong Elasticsearch: %v", event.Product.ID, err)
				// Không commit offset để tin nhắn có thể được xử lý lại sau
			} else {
				log.Printf("Đã lập chỉ mục/cập nhật sản phẩm %s thành công trong Elasticsearch.", event.Product.ID)
				// Commit offset chỉ khi xử lý thành công
				if err := c.reader.CommitMessages(ctx, m); err != nil {
					log.Printf("Lỗi khi commit offset cho sự kiện %s: %v", event.Type, err)
				}
			}
		case "ProductDeleted":
			if event.ID == "" {
				log.Printf("Bỏ qua sự kiện ProductDeleted: không có ID sản phẩm")
				break
			}
			_, err := c.searchService.DeleteProductFromIndex(ctx, &search_client.DeleteProductFromIndexRequest{ProductId: event.ID})
			if err != nil {
				log.Printf("Lỗi khi xóa sản phẩm %s khỏi chỉ mục Elasticsearch: %v", event.ID, err)
			} else {
				log.Printf("Đã xóa sản phẩm %s khỏi chỉ mục Elasticsearch thành công.", event.ID)
				if err := c.reader.CommitMessages(ctx, m); err != nil {
					log.Printf("Lỗi khi commit offset cho sự kiện ProductDeleted: %v", err)
				}
			}
		default:
			log.Printf("Loại sự kiện sản phẩm không xác định: %s", event.Type)
			if err := c.reader.CommitMessages(ctx, m); err != nil { // Commit unknown message to avoid reprocessing
				log.Printf("Lỗi khi commit offset cho sự kiện không xác định: %v", err)
			}
		}
	}
}

// Close closes the Kafka reader.
// Close đóng Kafka reader.
func (c *KafkaProductEventConsumer) Close() error {
	log.Println("Đóng Kafka Product Event Consumer...")
	return c.reader.Close()
}
