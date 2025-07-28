// internal/recommendation/domain/interaction.go
package domain

import "time"

// UserInteraction represents a record of a user interacting with a product.
// UserInteraction đại diện cho một bản ghi về việc người dùng tương tác với sản phẩm.
type UserInteraction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	EventType string    `json:"event_type"` // e.g., "view", "add_to_cart", "purchase"
	Timestamp time.Time `json:"timestamp"`
}

// NewUserInteraction creates a new UserInteraction instance.
// NewUserInteraction tạo một thể hiện mới của UserInteraction.
func NewUserInteraction(id, userID, productID, eventType string, timestamp time.Time) *UserInteraction {
	return &UserInteraction{
		ID:        id,
		UserID:    userID,
		ProductID: productID,
		EventType: eventType,
		Timestamp: timestamp,
	}
}
