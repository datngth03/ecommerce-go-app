// internal/order/domain/repository.go
package domain

import (
	"context"
)

// OrderRepository defines the interface for order data operations.
// OrderRepository định nghĩa interface cho các thao tác dữ liệu đơn hàng.
type OrderRepository interface {
	// Save creates a new order or updates an existing one.
	// Save tạo đơn hàng mới hoặc cập nhật đơn hàng hiện có.
	Save(ctx context.Context, order *Order) error

	// FindByID retrieves an order by its ID.
	// FindByID lấy đơn hàng theo ID của nó.
	FindByID(ctx context.Context, id string) (*Order, error)

	// FindAll retrieves a list of orders based on filters and pagination.
	// FindAll lấy danh sách các đơn hàng dựa trên bộ lọc và phân trang.
	FindAll(ctx context.Context, userID, status string, limit, offset int32) ([]*Order, int32, error)

	// Delete removes an order from the repository.
	// Delete xóa đơn hàng khỏi repository.
	Delete(ctx context.Context, id string) error
}
