package models

import "time"

type Cart struct {
	ID          string     `json:"id"`
	UserID      int64      `json:"user_id"`
	Items       []CartItem `json:"items"`
	TotalAmount float64    `json:"total_amount"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CartItem struct {
	ID          string    `json:"id"`
	CartID      string    `json:"cart_id"`
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int32     `json:"quantity"`
	Price       float64   `json:"price"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
