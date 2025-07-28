// internal/search/domain/product_search.go
package domain

import (
	"time"
)

// SearchProduct represents a product as it's indexed in the search engine.
// It contains fields relevant for searching and displaying search results.
// SearchProduct đại diện cho một sản phẩm khi nó được lập chỉ mục trong công cụ tìm kiếm.
// Nó chứa các trường liên quan đến tìm kiếm và hiển thị kết quả tìm kiếm.
type SearchProduct struct {
	ID            string    `json:"id"`             // Unique identifier for the product
	Name          string    `json:"name"`           // Product name, primary search field
	Description   string    `json:"description"`    // Product description, also searchable
	Price         float64   `json:"price"`          // Price of the product
	CategoryID    string    `json:"category_id"`    // ID of the category the product belongs to
	ImageURLs     []string  `json:"image_urls"`     // List of image URLs for the product
	StockQuantity int32     `json:"stock_quantity"` // Current stock quantity
	CreatedAt     time.Time `json:"created_at"`     // Timestamp of product creation
	UpdatedAt     time.Time `json:"updated_at"`     // Timestamp of last update
}

// NewSearchProduct creates a new SearchProduct instance.
// This constructor helps to ensure consistent initialization of the domain entity.
// NewSearchProduct tạo một thể hiện SearchProduct mới.
// Hàm tạo này giúp đảm bảo việc khởi tạo nhất quán của thực thể miền.
func NewSearchProduct(id, name, description string, price float64, categoryID string, imageURLs []string, stockQuantity int32, createdAt, updatedAt time.Time) *SearchProduct {
	return &SearchProduct{
		ID:            id,
		Name:          name,
		Description:   description,
		Price:         price,
		CategoryID:    categoryID,
		ImageURLs:     imageURLs,
		StockQuantity: stockQuantity,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

// UpdateSearchProduct updates the mutable fields of a SearchProduct.
// It also updates the UpdatedAt timestamp to reflect the change.
// UpdateSearchProduct cập nhật các trường có thể thay đổi của một SearchProduct.
// Nó cũng cập nhật dấu thời gian UpdatedAt để phản ánh sự thay đổi.
func (p *SearchProduct) UpdateSearchProduct(name, description string, price float64, categoryID string, imageURLs []string, stockQuantity int32) {
	p.Name = name
	p.Description = description
	p.Price = price
	p.CategoryID = categoryID
	p.ImageURLs = imageURLs
	p.StockQuantity = stockQuantity
	p.UpdatedAt = time.Now() // Update timestamp when product info changes
}
