// internal/inventory/infrastructure/messaging/kafka_event_consumer.go
package messaging

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go" // Kafka Go client

	"github.com/datngth03/ecommerce-go-app/internal/inventory/application"
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Import Inventory gRPC client (cho SetStockRequest)
)

// ProductEventPayload mirrors the structure of Product from Product Service's domain
// to enable unmarshalling of Kafka events without cross-package internal dependency.
// Cấu trúc ProductEventPayload phản ánh cấu trúc của Product từ domain của Product Service
// để cho phép giải mã các sự kiện Kafka mà không cần phụ thuộc chéo vào gói nội bộ.
type ProductEventPayload struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Price         float64   `json:"price"` // Changed to float64 for consistency with Product domain
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
	Type        string          `json:"type"` // Kích hoạt từ 'type' sang 'event_type' để khớp với Publisher
	Timestamp   string          `json:"timestamp"`  // RFC3339 format
	Payload     json.RawMessage `json:"payload"`    // Chuyển lại thành json.RawMessage để giải mã thủ công payload
	AggregateID string          `json:"aggregate_id"`
}

// KafkaProductEventConsumer defines the consumer for product events in Inventory Service.
// KafkaProductEventConsumer định nghĩa consumer sự kiện sản phẩm trong Inventory Service.
type KafkaProductEventConsumer struct {
	reader           *kafka.Reader
	inventoryService application.InventoryService
}

// NewKafkaProductEventConsumer creates a new Kafka event consumer for product events.
// NewKafkaProductEventConsumer tạo một consumer sự kiện Kafka mới cho các sự kiện sản phẩm.
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
	}
}

// StartConsuming starts consuming messages from Kafka.
// StartConsuming bắt đầu tiêu thụ tin nhắn từ Kafka.
func (c *KafkaProductEventConsumer) StartConsuming(ctx context.Context) {
	log.Printf("Inventory Service Consumer: Bắt đầu tiêu thụ sự kiện từ Kafka topic '%s'...", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			log.Println("Inventory Service Consumer: Context cancelled, stopping consumer.")
			return
		default:
			m, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() == context.Canceled {
					log.Println("Inventory Service Consumer ngừng hoạt động do context bị hủy.")
					return
				}
				if err == context.DeadlineExceeded {
					continue // Just try again if it's a timeout
				}
				log.Printf("Inventory Service Consumer: Lỗi khi đọc tin nhắn từ Kafka: %v", err)
				continue
			}

			log.Printf("Inventory Service Consumer: Đã nhận tin nhắn từ Kafka: topic=%s, partition=%d, offset=%d, key=%s",
				m.Topic, m.Partition, m.Offset, string(m.Key))

			var event ProductEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("Inventory Service Consumer: Lỗi khi giải mã sự kiện sản phẩm (cấu trúc sự kiện ngoài cùng): %v, tin nhắn: %s", err, string(m.Value))
				// Commit offset để tránh xử lý lại tin nhắn bị lỗi định dạng
				if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
					log.Printf("Inventory Service Consumer: Lỗi khi commit offset sau lỗi giải mã sự kiện ngoài cùng: %v", commitErr)
				}
				continue
			}

			log.Printf("Inventory Service Consumer: Đã giải mã sự kiện loại: %s", event.Type)

			var payload ProductEventPayload // Khai báo payload riêng biệt
			if len(event.Payload) > 0 {
				if err := json.Unmarshal(event.Payload, &payload); err != nil { // Giải mã Payload cụ thể
					log.Printf("Inventory Service Consumer: Lỗi khi giải mã payload sản phẩm: %v, payload raw: %s", err, string(event.Payload))
					// Commit offset để tránh xử lý lại tin nhắn bị lỗi định dạng payload
					if commitErr := c.reader.CommitMessages(ctx, m); commitErr != nil {
						log.Printf("Inventory Service Consumer: Lỗi khi commit offset sau lỗi giải mã payload: %v", commitErr)
					}
					continue
				}
			}

			// Xử lý sự kiện dựa trên loại sự kiện
			switch event.Type {
			case "ProductCreated":
				log.Printf("Inventory Service Consumer: Xử lý sự kiện ProductCreated cho Product ID: %s", payload.ID)
				// Khởi tạo tồn kho cho sản phẩm mới
				req := &inventory_client.SetStockRequest{
					ProductId: payload.ID, // Sửa: Truy cập từ payload đã giải mã
					Quantity:  0,          // Khởi tạo tồn kho là 0 cho sản phẩm mới
				}
				_, err := c.inventoryService.SetStock(ctx, req)
				if err != nil {
					log.Printf("Inventory Service Consumer: Lỗi khi khởi tạo tồn kho cho sản phẩm %s: %v", payload.ID, err) // Sửa: Truy cập từ payload
					// Không commit offset để tin nhắn có thể được xử lý lại sau (ví dụ: lỗi DB tạm thời)
				} else {
					log.Printf("Inventory Service Consumer: Đã khởi tạo tồn kho cho sản phẩm %s thành công.", payload.ID) // Sửa: Truy cập từ payload
					// Commit offset chỉ khi xử lý thành công
					if err := c.reader.CommitMessages(ctx, m); err != nil {
						log.Printf("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện %s: %v", event.Type, err)
					}
				}
			case "ProductUpdated":
				log.Printf("Inventory Service Consumer: Xử lý sự kiện ProductUpdated cho Product ID: %s", payload.ID)
				// TODO: Triển khai logic cập nhật tồn kho dựa trên sự kiện cập nhật sản phẩm
				// Ví dụ: Có thể cập nhật trường tồn kho trên Product (nếu được cập nhật bởi admin) hoặc chỉ đơn giản là ghi log.
				// Nếu tồn kho thực sự được quản lý bởi Inventory, ProductUpdated thường không thay đổi tồn kho trực tiếp.
				if err := c.reader.CommitMessages(ctx, m); err != nil {
					log.Printf("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện %s: %v", event.Type, err)
				}
			case "ProductDeleted":
				log.Printf("Inventory Service Consumer: Xử lý sự kiện ProductDeleted cho Product ID: %s", event.AggregateID) // Sửa: Dùng AggregateID cho Deleted Event
				// TODO: Triển khai logic xóa tồn kho cho sản phẩm đã bị xóa
				// Ví dụ: Gọi inventoryService.DeleteInventoryItem(ctx, event.AggregateID)
				if err := c.reader.CommitMessages(ctx, m); err != nil {
					log.Printf("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện %s: %v", event.Type, err)
				}
			default:
				log.Printf("Inventory Service Consumer: Loại sự kiện sản phẩm không xác định hoặc không được xử lý: %s", event.Type)
				// Commit unknown message to avoid reprocessing
				if err := c.reader.CommitMessages(ctx, m); err != nil {
					log.Printf("Inventory Service Consumer: Lỗi khi commit offset cho sự kiện không xác định: %v", err)
				}
			}
		}
	}
}

// Close closes the Kafka reader.
// Close đóng Kafka reader.
func (c *KafkaProductEventConsumer) Close() error {
	log.Println("Đóng Kafka Product Event Consumer (Inventory Service)...")
	return c.reader.Close()
}
