// internal/order/domain/order.go
package domain

import (
	"time"
)

// OrderItem represents a product within an order.
// OrderItem đại diện cho một sản phẩm trong một đơn hàng.
type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int32   `json:"quantity"`
}

// Order represents the core order entity in the domain.
// Order đại diện cho thực thể đơn hàng cốt lõi trong domain.
type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Items           []OrderItem `json:"items"`
	TotalAmount     float64     `json:"total_amount"`
	Status          string      `json:"status"` // e.g., "pending", "paid", "shipped", "cancelled"
	ShippingAddress string      `json:"shipping_address"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// NewOrder creates a new Order instance.
// NewOrder tạo một thể hiện Order mới.
func NewOrder(id, userID string, items []OrderItem, totalAmount float64, shippingAddress string) *Order {
	now := time.Now()
	return &Order{
		ID:              id,
		UserID:          userID,
		Items:           items,
		TotalAmount:     totalAmount,
		Status:          "pending", // Default status
		ShippingAddress: shippingAddress,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// UpdateStatus updates the order's status.
// UpdateStatus cập nhật trạng thái của đơn hàng.
func (o *Order) UpdateStatus(newStatus string) {
	o.Status = newStatus
	o.UpdatedAt = time.Now()
}
