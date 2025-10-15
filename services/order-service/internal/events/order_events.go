package events

import (
	"time"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
)

// Event types
const (
	EventOrderCreated       = "order.created"
	EventOrderStatusChanged = "order.status.changed"
	EventOrderCancelled     = "order.cancelled"
	EventOrderCompleted     = "order.completed"
)

// OrderCreatedEvent represents order creation event
type OrderCreatedEvent struct {
	EventType       string           `json:"event_type"`
	OrderID         string           `json:"order_id"`
	UserID          int64            `json:"user_id"`
	TotalAmount     float64          `json:"total_amount"`
	ShippingAddress string           `json:"shipping_address"`
	PaymentMethod   string           `json:"payment_method"`
	Items           []OrderItemEvent `json:"items"`
	CreatedAt       time.Time        `json:"created_at"`
}

// OrderStatusChangedEvent represents order status change event
type OrderStatusChangedEvent struct {
	EventType string    `json:"event_type"`
	OrderID   string    `json:"order_id"`
	UserID    int64     `json:"user_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderCancelledEvent represents order cancellation event
type OrderCancelledEvent struct {
	EventType   string    `json:"event_type"`
	OrderID     string    `json:"order_id"`
	UserID      int64     `json:"user_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// OrderItemEvent represents an order item in events
type OrderItemEvent struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int32   `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

// Helper functions to convert models to events

func NewOrderCreatedEvent(order *models.Order) *OrderCreatedEvent {
	items := make([]OrderItemEvent, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemEvent{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	return &OrderCreatedEvent{
		EventType:       EventOrderCreated,
		OrderID:         order.ID,
		UserID:          order.UserID,
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		PaymentMethod:   order.PaymentMethod,
		Items:           items,
		CreatedAt:       order.CreatedAt,
	}
}

func NewOrderStatusChangedEvent(order *models.Order, oldStatus string) *OrderStatusChangedEvent {
	return &OrderStatusChangedEvent{
		EventType: EventOrderStatusChanged,
		OrderID:   order.ID,
		UserID:    order.UserID,
		OldStatus: oldStatus,
		NewStatus: order.Status,
		UpdatedAt: order.UpdatedAt,
	}
}

func NewOrderCancelledEvent(order *models.Order, reason string) *OrderCancelledEvent {
	return &OrderCancelledEvent{
		EventType:   EventOrderCancelled,
		OrderID:     order.ID,
		UserID:      order.UserID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}
}
