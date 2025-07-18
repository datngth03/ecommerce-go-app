// internal/product/domain/repository.go
package domain

import (
	"context"
)

// ProductRepository defines the interface for product data operations.
// ProductRepository định nghĩa interface cho các thao tác dữ liệu sản phẩm.
type ProductRepository interface {
	// Save creates a new product or updates an existing one.
	// Save tạo sản phẩm mới hoặc cập nhật sản phẩm hiện có.
	Save(ctx context.Context, product *Product) error

	// FindByID retrieves a product by its ID.
	// FindByID lấy sản phẩm theo ID của nó.
	FindByID(ctx context.Context, id string) (*Product, error)

	// FindAll retrieves a list of products based on filters and pagination.
	// FindAll lấy danh sách các sản phẩm dựa trên bộ lọc và phân trang.
	FindAll(ctx context.Context, categoryID string, limit, offset int32) ([]*Product, int32, error)

	// Delete removes a product from the repository.
	// Delete xóa sản phẩm khỏi repository.
	Delete(ctx context.Context, id string) error
}

// CategoryRepository defines the interface for category data operations.
// CategoryRepository định nghĩa interface cho các thao tác dữ liệu danh mục.
type CategoryRepository interface {
	// Save creates a new category or updates an existing one.
	// Save tạo danh mục mới hoặc cập nhật danh mục hiện có.
	Save(ctx context.Context, category *Category) error

	// FindByID retrieves a category by its ID.
	// FindByID lấy danh mục theo ID của nó.
	FindByID(ctx context.Context, id string) (*Category, error)

	// FindByName retrieves a category by its name.
	// FindByName lấy danh mục theo tên của nó.
	FindByName(ctx context.Context, name string) (*Category, error)

	// FindAll retrieves a list of all categories.
	// FindAll lấy danh sách tất cả các danh mục.
	FindAll(ctx context.Context) ([]*Category, error)

	// Delete removes a category from the repository.
	// Delete xóa danh mục khỏi repository.
	Delete(ctx context.Context, id string) error
}
