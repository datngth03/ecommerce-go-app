// internal/search/domain/repository.go
package domain

import (
	"context"
	"errors"
)

// Sentinel errors for Search Service.
// Lỗi Sentinel cho Search Service.
var (
	ErrProductNotFoundInSearch = errors.New("product not found in search index")
	ErrFailedToIndexProduct    = errors.New("failed to index product")
	ErrFailedToSearchProducts  = errors.New("failed to search products")
	ErrFailedToDeleteFromIndex = errors.New("failed to delete product from index")
)

// SearchProductRepository defines the interface for interacting with product search data.
// SearchProductRepository định nghĩa giao diện để tương tác với dữ liệu tìm kiếm sản phẩm.
type SearchProductRepository interface {
	// IndexProduct adds or updates a product in the search engine.
	// IndexProduct thêm hoặc cập nhật một sản phẩm vào công cụ tìm kiếm.
	IndexProduct(ctx context.Context, product *SearchProduct) error

	// SearchProducts searches for products based on a query and filters.
	// Returns a list of products and the total count of matching products.
	// SearchProducts tìm kiếm sản phẩm dựa trên truy vấn và bộ lọc.
	// Trả về danh sách sản phẩm và tổng số lượng sản phẩm khớp.
	SearchProducts(ctx context.Context, query, categoryID string, minPrice, maxPrice float64, limit, offset int32) ([]*SearchProduct, int64, error)

	// DeleteProductFromIndex deletes a product from the search engine by its ID.
	// DeleteProductFromIndex xóa một sản phẩm khỏi chỉ mục tìm kiếm bằng ID của nó.
	DeleteProductFromIndex(ctx context.Context, productID string) error
}
