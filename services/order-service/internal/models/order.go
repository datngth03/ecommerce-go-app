package models

import "time"

type Order struct {
	ID              string      `db:"id" json:"id"`
	UserID          int64       `db:"user_id" json:"user_id"`
	Status          string      `db:"status" json:"status"`
	TotalAmount     float64     `db:"total_amount" json:"total_amount"`
	ShippingAddress string      `db:"shipping_address" json:"shipping_address"`
	PaymentMethod   string      `db:"payment_method" json:"payment_method"`
	CreatedAt       time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time   `db:"updated_at" json:"updated_at"`
	Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID          string    `db:"id" json:"id"`
	OrderID     string    `db:"order_id" json:"order_id"`
	ProductID   string    `db:"product_id" json:"product_id"`
	ProductName string    `db:"product_name" json:"product_name"`
	Quantity    int32     `db:"quantity" json:"quantity"`
	Price       float64   `db:"price" json:"price"`
	Subtotal    float64   `db:"subtotal" json:"subtotal"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)
