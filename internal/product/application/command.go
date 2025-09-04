// internal/product/application/command.go
package application

import (
	"errors"
	"time"
)

// --- Product Commands ---

// CreateProductCommand represents the intent to create a new product
type CreateProductCommand struct {
	BrandID     *string  `validate:"omitempty,uuid4"`
	Name        string   `validate:"required,min=1,max=255"`
	Slug        string   `validate:"omitempty,max=255"`
	Description *string  `validate:"omitempty,max=2000"`
	CategoryIDs []string `validate:"dive,uuid4"`
	TagIDs      []string `validate:"dive,uuid4"`

	// Images
	Images []CreateProductImageCommand `validate:"dive"`

	// Specifications
	Specifications []CreateProductSpecificationCommand `validate:"dive"`

	// Metadata
	CreatedBy string `validate:"required,uuid4"` // User ID who creates the product
}

// CreateProductImageCommand represents image data for product creation
type CreateProductImageCommand struct {
	URL       string `validate:"required,url"`
	IsPrimary bool
}

// CreateProductSpecificationCommand represents specification data for product creation
type CreateProductSpecificationCommand struct {
	AttributeID string `validate:"required,uuid4"`
	Value       string `validate:"required,max=500"`
}

// UpdateProductCommand represents the intent to update an existing product
type UpdateProductCommand struct {
	ID          string  `validate:"required,uuid4"`
	BrandID     string  `validate:"omitempty,uuid4"`
	Name        string  `validate:"required,min=1,max=255"`
	Slug        string  `validate:"omitempty,max=255"`
	Description *string `validate:"omitempty,max=2000"`

	// Images - complete replacement
	Images []UpdateProductImageCommand `validate:"dive"`

	// Specifications - complete replacement
	Specifications []UpdateProductSpecificationCommand `validate:"dive"`

	// Metadata
	UpdatedBy string `validate:"required,uuid4"` // User ID who updates the product
}

// UpdateProductImageCommand represents image data for product update
type UpdateProductImageCommand struct {
	ID        string `validate:"omitempty,uuid4"` // Empty for new images
	URL       string `validate:"required,url"`
	IsPrimary bool
}

// UpdateProductSpecificationCommand represents specification data for product update
type UpdateProductSpecificationCommand struct {
	ID          string `validate:"omitempty,uuid4"` // Empty for new specifications
	AttributeID string `validate:"required,uuid4"`
	Value       string `validate:"required,max=500"`
}

// DeleteProductCommand represents the intent to delete a product
type DeleteProductCommand struct {
	ID        string `validate:"required,uuid4"`
	DeletedBy string `validate:"required,uuid4"`    // User ID who deletes the product
	Reason    string `validate:"omitempty,max=500"` // Optional deletion reason
}

// ListProductsQuery represents the intent to list products with filtering
type ListProductsQuery struct {
	// Basic filters
	Name        string `validate:"omitempty,max=255"`
	Slug        string `validate:"omitempty,max=255"`
	BrandID     string `validate:"omitempty,uuid4"`
	CategoryID  string `validate:"omitempty,uuid4"`
	SearchQuery string `validate:"omitempty,max=500"`

	// Collection filters
	TagIDs      []string `validate:"dive,uuid4"`
	CategoryIDs []string `validate:"dive,uuid4"`

	// Price filters
	MinPrice *float64 `validate:"omitempty,min=0"`
	MaxPrice *float64 `validate:"omitempty,min=0"`

	// Date filters
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time

	// Status filters
	HasBrand          *bool
	HasImages         *bool
	HasSpecifications *bool
	HasVariants       *bool

	// Pagination
	Limit  int `validate:"min=1,max=100"`
	Offset int `validate:"min=0"`

	// Sorting
	SortBy    string `validate:"omitempty,oneof=name slug created_at updated_at rating review_count"`
	SortOrder string `validate:"omitempty,oneof=asc desc"`

	// Context
	RequestedBy string `validate:"required,uuid4"` // User ID making the request
}

// GetProductQuery represents the intent to get a single product
type GetProductQuery struct {
	ID   *string `validate:"omitempty,uuid4"`
	Slug *string `validate:"omitempty,max=255"`

	// Load relations
	IncludeBrand          bool
	IncludeCategories     bool
	IncludeTags           bool
	IncludeImages         bool
	IncludeSpecifications bool
	IncludeVariants       bool

	// Context
	RequestedBy string `validate:"required,uuid4"`
}

// --- Category Relation Commands ---

// AddProductCategoriesCommand represents the intent to add categories to a product
type AddProductCategoriesCommand struct {
	ProductID   string   `validate:"required,uuid4"`
	CategoryIDs []string `validate:"required,min=1,dive,uuid4"`
	AddedBy     string   `validate:"required,uuid4"`
}

// RemoveProductCategoriesCommand represents the intent to remove categories from a product
type RemoveProductCategoriesCommand struct {
	ProductID   string   `validate:"required,uuid4"`
	CategoryIDs []string `validate:"required,min=1,dive,uuid4"`
	RemovedBy   string   `validate:"required,uuid4"`
}

// GetProductCategoriesQuery represents the intent to get categories for a product
type GetProductCategoriesQuery struct {
	ProductID   string `validate:"required,uuid4"`
	RequestedBy string `validate:"required,uuid4"`
}

// --- Tag Relation Commands ---

// AddProductTagsCommand represents the intent to add tags to a product
type AddProductTagsCommand struct {
	ProductID string   `validate:"required,uuid4"`
	TagIDs    []string `validate:"required,min=1,dive,uuid4"`
	AddedBy   string   `validate:"required,uuid4"`
}

// RemoveProductTagsCommand represents the intent to remove tags from a product
type RemoveProductTagsCommand struct {
	ProductID string   `validate:"required,uuid4"`
	TagIDs    []string `validate:"required,min=1,dive,uuid4"`
	RemovedBy string   `validate:"required,uuid4"`
}

// GetProductTagsQuery represents the intent to get tags for a product
type GetProductTagsQuery struct {
	ProductID   string `validate:"required,uuid4"`
	RequestedBy string `validate:"required,uuid4"`
}

// --- Command Results ---

// CreateProductResult represents the result of creating a product
type CreateProductResult struct {
	ProductID       string    `json:"product_id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description,omitempty"`
	Rating          float64   `json:"rating"`
	ReviewCount     int32     `json:"review_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CategoriesAdded int       `json:"categories_added"`
	TagsAdded       int       `json:"tags_added"`
	ImagesAdded     int       `json:"images_added"`
	SpecsAdded      int       `json:"specifications_added"`
}

// UpdateProductResult represents the result of updating a product
type UpdateProductResult struct {
	ProductID     string    `json:"product_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	UpdatedAt     time.Time `json:"updated_at"`
	ImagesUpdated int       `json:"images_updated"`
	SpecsUpdated  int       `json:"specifications_updated"`
}

// ListProductsResult represents the result of listing products
type ListProductsResult struct {
	Products []*ProductSummary `json:"products"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	HasNext  bool              `json:"has_next"`
	HasPrev  bool              `json:"has_prev"`
}

// ProductSummary represents a minimal product representation for listings
type ProductSummary struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	BrandID *string `json:"brand_id,omitempty"`
	// BrandName     *string   `json:"brand_name,omitempty"`
	PrimaryImage  *string   `json:"primary_image,omitempty"`
	CategoryCount int       `json:"category_count"`
	TagCount      int       `json:"tag_count"`
	ImageCount    int       `json:"image_count"`
	SpecCount     int       `json:"specification_count"`
	VariantCount  int       `json:"variant_count"`
	Rating        float64   `json:"rating"`
	ReviewCount   int32     `json:"review_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// GetProductResult represents the result of getting a single product
type GetProductResult struct {
	Product *ProductDetail `json:"product"`
}

// ProductDetail represents complete product information
type ProductDetail struct {
	ID          string    `json:"id"`
	BrandID     *string   `json:"brand_id,omitempty"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	Rating      float64   `json:"rating"`
	ReviewCount int32     `json:"review_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations (loaded based on query parameters)
	Brand          *BrandSummary           `json:"brand,omitempty"`
	Categories     []*CategorySummary      `json:"categories,omitempty"`
	Tags           []*TagSummary           `json:"tags,omitempty"`
	Images         []*ImageSummary         `json:"images,omitempty"`
	Specifications []*SpecificationSummary `json:"specifications,omitempty"`
	Variants       []*VariantSummary       `json:"variants,omitempty"`
}

// Related entity summaries
type BrandSummary struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Logo        *string `json:"logo,omitempty"`
	Description *string `json:"description,omitempty"`
}

type CategorySummary struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	Image       *string `json:"image,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

type TagSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ImageSummary struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
}

type SpecificationSummary struct {
	ID            string `json:"id"`
	AttributeID   string `json:"attribute_id"`
	AttributeName string `json:"attribute_name"`
	Value         string `json:"value"`
}

type VariantSummary struct {
	ID            string   `json:"id"`
	SKU           string   `json:"sku"`
	Price         float64  `json:"price"`
	OriginalPrice float64  `json:"original_price"`
	Discount      *float64 `json:"discount,omitempty"`
}

// --- Command Validation Methods ---

// Validate validates CreateProductCommand
func (cmd *CreateProductCommand) Validate() error {
	if cmd.Name == "" {
		return errors.New("product name is required")
	}
	if len(cmd.Name) > 255 {
		return errors.New("product name cannot exceed 255 characters")
	}
	if cmd.CreatedBy == "" {
		return errors.New("created_by is required")
	}
	return nil
}

// Validate validates UpdateProductCommand
func (cmd *UpdateProductCommand) Validate() error {
	if cmd.ID == "" {
		return errors.New("product ID is required")
	}
	if cmd.Name == "" {
		return errors.New("product name is required")
	}
	if len(cmd.Name) > 255 {
		return errors.New("product name cannot exceed 255 characters")
	}
	if cmd.UpdatedBy == "" {
		return errors.New("updated_by is required")
	}
	return nil
}

// SetDefaults sets default values for ListProductsQuery
func (q *ListProductsQuery) SetDefaults() {
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
	if q.SortOrder == "" {
		q.SortOrder = "desc"
	}
}

// SetDefaults sets default values for GetProductQuery
func (q *GetProductQuery) SetDefaults() {
	// By default, include basic relations
	q.IncludeBrand = true
	q.IncludeCategories = true
	q.IncludeTags = true
	q.IncludeImages = true
}
