package models

import (
	"strings"
	"time"
)

// Product represents a product in the system
type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=255"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price" validate:"required,gt=0"`
	CategoryID  string    `json:"category_id" db:"category_id" validate:"required"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relation (not stored in DB, populated when needed)
	Category *Category `json:"category,omitempty"`
}

// CreateProductRequest represents the request to create a new product
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	CategoryID  string  `json:"category_id" validate:"required"`
	ImageURL    string  `json:"image_url"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	CategoryID  string  `json:"category_id" validate:"required"`
	ImageURL    string  `json:"image_url"`
	IsActive    bool    `json:"is_active"`
}

// ProductResponse represents the response for product operations
type ProductResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	CategoryID  string            `json:"category_id"`
	ImageURL    string            `json:"image_url"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Category    *CategoryResponse `json:"category,omitempty"`
}

// ListProductsRequest represents the request for listing products
type ListProductsRequest struct {
	Page       int    `json:"page" form:"page" validate:"min=1"`
	PageSize   int    `json:"page_size" form:"page_size" validate:"min=1,max=100"`
	CategoryID string `json:"category_id" form:"category_id"`
}

// ListProductsResponse represents the response for listing products
type ListProductsResponse struct {
	Products   []ProductResponse `json:"products"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// GenerateSlug creates a URL-friendly slug from the product name
func (p *Product) GenerateSlug() {
	slug := strings.ToLower(p.Name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	p.Slug = slug
}

// ToResponse converts Product model to ProductResponse
func (p *Product) ToResponse() ProductResponse {
	response := ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Price:       p.Price,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.Category != nil {
		categoryResponse := p.Category.ToResponse()
		response.Category = &categoryResponse
	}

	return response
}

// TableName returns the table name for GORM
func (Product) TableName() string {
	return "products"
}
