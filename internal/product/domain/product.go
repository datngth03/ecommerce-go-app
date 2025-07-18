// internal/product/domain/product.go
package domain

import (
	"time"
)

// Product represents the core product entity in the domain.
// Product đại diện cho thực thể sản phẩm cốt lõi trong domain.
type Product struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Price         float64   `json:"price"`
	CategoryID    string    `json:"category_id"`
	ImageURLs     []string  `json:"image_urls,omitempty"`
	StockQuantity int32     `json:"stock_quantity"` // Có thể được đồng bộ từ Inventory Service
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewProduct creates a new Product instance.
// NewProduct tạo một thể hiện Product mới.
func NewProduct(id, name, description string, price float64, categoryID string, imageURLs []string) *Product {
	now := time.Now()
	return &Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		CategoryID:  categoryID,
		ImageURLs:   imageURLs,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateProductInfo updates product's information.
// UpdateProductInfo cập nhật thông tin sản phẩm.
func (p *Product) UpdateProductInfo(name, description string, price float64, categoryID string, imageURLs []string) {
	p.Name = name
	p.Description = description
	p.Price = price
	p.CategoryID = categoryID
	p.ImageURLs = imageURLs
	p.UpdatedAt = time.Now()
}

// Category represents the core category entity in the domain.
// Category đại diện cho thực thể danh mục cốt lõi trong domain.
type Category struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewCategory creates a new Category instance.
// NewCategory tạo một thể hiện Category mới.
func NewCategory(id, name, description string) *Category {
	now := time.Now()
	return &Category{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateCategoryInfo updates category's information.
// UpdateCategoryInfo cập nhật thông tin danh mục.
func (c *Category) UpdateCategoryInfo(name, description string) {
	c.Name = name
	c.Description = description
	c.UpdatedAt = time.Now()
}
