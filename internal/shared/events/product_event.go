// internal/shared/events/product_event.go
package events

import (
	"encoding/json" // Import for json.RawMessage
	"time"
)

// ProductEventPayload mirrors the structure of Product from Product Service's domain
// to enable unmarshalling of Kafka events consistently across services.
// ProductEventPayload phản ánh cấu trúc của Product từ domain của Product Service
// để cho phép giải mã các sự kiện Kafka một cách nhất quán giữa các dịch vụ.
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

// ProductEvent defines the universal structure of product-related events.
// This structure should be used by both event publishers and consumers.
// ProductEvent định nghĩa cấu trúc chung cho các sự kiện liên quan đến sản phẩm.
// Cấu trúc này nên được sử dụng bởi cả bên phát và bên tiêu thụ sự kiện.
type ProductEvent struct {
	Type        string          `json:"type"`               // e.g., "ProductCreated", "ProductUpdated", "ProductDeleted"
	Timestamp   string          `json:"timestamp"`          // RFC3339 format
	Payload     json.RawMessage `json:"payload"`            // Raw JSON bytes of the ProductEventPayload
	AggregateID string          `json:"aggregate_id"`       // ID of the aggregate (product ID)
	TraceID     string          `json:"trace_id,omitempty"` // Trace ID để theo dõi phân tán qua Kafka
	SpanID      string          `json:"span_id,omitempty"`  // Span ID để theo dõi phân tán qua Kafka
}
