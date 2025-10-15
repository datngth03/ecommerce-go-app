package models

import (
	"strings"
	"time"
)

// Category represents a product category
type Category struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateCategoryRequest represents the request to create a new category
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CategoryResponse represents the response for category operations
type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListCategoriesResponse represents the response for listing categories
type ListCategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int64              `json:"total"`
}

// GenerateSlug creates a URL-friendly slug from the category name
func (c *Category) GenerateSlug() {
	slug := strings.ToLower(c.Name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	c.Slug = slug
}

// ToResponse converts Category model to CategoryResponse
func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// TableName returns the table name for GORM
func (Category) TableName() string {
	return "categories"
}
