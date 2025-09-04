// internal/product/domain/repository.go
package domain

import (
	// "context"
	"errors"
)

var (
	// Brand errors
	ErrBrandNotFound    = errors.New("brand not found")
	ErrInvalidBrandData = errors.New("invalid brand data")

	// Category errors
	ErrCategoryNotFound    = errors.New("category not found")
	ErrCategoryHasChildren = errors.New("category has child categories, cannot delete")
	ErrInvalidCategoryData = errors.New("invalid category data")

	// Product errors
	ErrProductNotFound    = errors.New("product not found")
	ErrInvalidProductData = errors.New("invalid product data")

	// Product Variant errors
	ErrProductVariantNotFound    = errors.New("product variant not found")
	ErrInvalidProductVariantData = errors.New("invalid product variant data")
	ErrProductAlreadyExists      = errors.New("product with the same name already exists")

	// Tag errors
	ErrTagNotFound    = errors.New("tag not found")
	ErrInvalidTagData = errors.New("invalid tag data")

	// slug errors
	ErrNotFound = errors.New("not found")

	// Image errors
	ErrProductImageNotFound = errors.New("product image not found")

	// Specification Attribute errors
	ErrSpecificationAttributeNotFound    = errors.New("specification attribute not found")
	ErrInvalidSpecificationAttributeData = errors.New("invalid specification attribute data")

	// Variant Attribute errors
	ErrVariantAttributeNotFound    = errors.New("variant attribute not found")
	ErrInvalidVariantAttributeData = errors.New("invalid variant attribute data")

	// Variant Attribute Value errors
	ErrVariantAttributeValueNotFound    = errors.New("variant attribute value not found")
	ErrInvalidVariantAttributeValueData = errors.New("invalid variant attribute value data")
)

// ProductRepository defines the interface for product data operations.

type BrandRepository interface {
	Create(brand *Brand) error
	GetByID(id string) (*Brand, error)
	GetBySlug(slug string) (*Brand, error)
	Update(brand *Brand) error
	Delete(id string) error
	List(filter BrandFilter) ([]*Brand, int, error)
}

type CategoryRepository interface {
	Create(category *Category) error
	GetByID(id string) (*Category, error)
	GetBySlug(slug string) (*Category, error)
	Update(category *Category) error
	Delete(id string) error
	List(filter CategoryFilter) ([]*Category, int, error)
	GetChildren(parentID string) ([]*Category, error)
	GetParent(categoryID string) (*Category, error)
}

type ProductRepository interface {
	Create(product *Product) error
	GetByID(id string) (*Product, error)
	GetBySlug(slug string) (*Product, error)
	Update(product *Product) error
	Delete(id string) error
	List(filter ProductFilter) ([]*Product, int, error)

	// Category relations
	AddCategories(productID string, categoryIDs []string) error
	RemoveCategories(productID string, categoryIDs []string) error
	GetCategories(productID string) ([]*Category, error)

	// Tag relations
	AddTags(productID string, tagIDs []string) error
	RemoveTags(productID string, tagIDs []string) error
	GetTags(productID string) ([]*Tag, error)
}

type ProductVariantRepository interface {
	Create(variant *ProductVariant) error
	GetByID(id string) (*ProductVariant, error)
	GetBySKU(sku string) (*ProductVariant, error)
	Update(variant *ProductVariant) error
	Delete(id string) error
	List(filter ProductVariantFilter) ([]*ProductVariant, int, error)
	GetByProductID(productID string) ([]*ProductVariant, error)
}

type TagRepository interface {
	Create(tag *Tag) error
	GetByID(id string) (*Tag, error)
	GetBySlug(slug string) (*Tag, error)
	Update(tag *Tag) error
	Delete(id string) error
	List(filter TagFilter) ([]*Tag, int, error)
}

type SpecificationAttributeRepository interface {
	Create(attr *SpecificationAttribute) error
	GetByID(id string) (*SpecificationAttribute, error)
	GetBySlug(slug string) (*SpecificationAttribute, error)
	Update(attr *SpecificationAttribute) error
	Delete(id string) error
	List(filter SpecificationAttributeFilter) ([]*SpecificationAttribute, int, error)
}

type VariantAttributeRepository interface {
	Create(attr *VariantAttribute) error
	GetByID(id string) (*VariantAttribute, error)
	GetBySlug(slug string) (*VariantAttribute, error)
	Update(attr *VariantAttribute) error
	Delete(id string) error
	List(filter VariantAttributeFilter) ([]*VariantAttribute, int, error)
}

type VariantAttributeValueRepository interface {
	Create(value *VariantAttributeValue) error
	GetByID(id string) (*VariantAttributeValue, error)
	Update(value *VariantAttributeValue) error
	Delete(id string) error
	List(filter VariantAttributeValueFilter) ([]*VariantAttributeValue, int, error)
	GetByAttributeID(attributeID string) ([]*VariantAttributeValue, error)
}

// Filter types for repository queries

type BrandFilter struct {
	Search    string
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int
}

type CategoryFilter struct {
	Search    string
	ParentID  *string
	OnlyRoot  bool
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int
}

type ProductFilter struct {
	Search      string
	CategoryIDs []string
	BrandID     string
	TagIDs      []string
	MinPrice    *float64
	MaxPrice    *float64
	MinRating   *float64
	SortBy      string
	SortOrder   SortOrder
	Limit       int
	Offset      int

	// Include options
	IncludeBrand          bool
	IncludeCategories     bool
	IncludeImages         bool
	IncludeVariants       bool
	IncludeTags           bool
	IncludeSpecifications bool
}

type ProductVariantFilter struct {
	ProductID string
	Search    string
	MinPrice  *float64
	MaxPrice  *float64
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int

	IncludeAttributeValues bool
}

type TagFilter struct {
	Search    string
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int
}

type SpecificationAttributeFilter struct {
	Search    string
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int
}

type VariantAttributeFilter struct {
	Search    string
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int
}

type VariantAttributeValueFilter struct {
	AttributeID string
	Search      string
	SortBy      string
	SortOrder   SortOrder
	Limit       int
	Offset      int
}

// SortOrder enum for consistency with protobuf
type SortOrder int

const (
	SortOrderUnspecified SortOrder = iota
	SortOrderAsc
	SortOrderDesc
)

func (s SortOrder) String() string {
	switch s {
	case SortOrderAsc:
		return "ASC"
	case SortOrderDesc:
		return "DESC"
	default:
		return "ASC"
	}
}

// // Service interfaces

// type ProductService interface {
// 	// Brand operations
// 	CreateBrand(brand *Brand) error
// 	GetBrand(id string) (*Brand, error)
// 	UpdateBrand(brand *Brand, fieldMask []string) error
// 	DeleteBrand(id string) error
// 	ListBrands(filter BrandFilter) ([]*Brand, int, error)

// 	// Category operations
// 	CreateCategory(category *Category) error
// 	GetCategory(id string, includeChildren, includeParent bool) (*Category, error)
// 	UpdateCategory(category *Category, fieldMask []string) error
// 	DeleteCategory(id string) error
// 	ListCategories(filter CategoryFilter) ([]*Category, int, error)

// 	// Product operations
// 	CreateProduct(product *Product) error
// 	GetProduct(id string, includes ProductIncludes) (*Product, error)
// 	UpdateProduct(product *Product, fieldMask []string) error
// 	DeleteProduct(id string) error
// 	ListProducts(filter ProductFilter) ([]*Product, int, error)

// 	// Product Variant operations
// 	CreateProductVariant(variant *ProductVariant) error
// 	GetProductVariant(id string, includeAttributeValues bool) (*ProductVariant, error)
// 	UpdateProductVariant(variant *ProductVariant, fieldMask []string) error
// 	DeleteProductVariant(id string) error
// 	ListProductVariants(filter ProductVariantFilter) ([]*ProductVariant, int, error)

// 	// Tag operations
// 	CreateTag(tag *Tag) error
// 	GetTag(id string) (*Tag, error)
// 	UpdateTag(tag *Tag, fieldMask []string) error
// 	DeleteTag(id string) error
// 	ListTags(filter TagFilter) ([]*Tag, int, error)

// 	// Specification Attribute operations
// 	CreateSpecificationAttribute(attr *SpecificationAttribute) error
// 	GetSpecificationAttribute(id string) (*SpecificationAttribute, error)
// 	UpdateSpecificationAttribute(attr *SpecificationAttribute, fieldMask []string) error
// 	DeleteSpecificationAttribute(id string) error
// 	ListSpecificationAttributes(filter SpecificationAttributeFilter) ([]*SpecificationAttribute, int, error)

// 	// Variant Attribute operations
// 	CreateVariantAttribute(attr *VariantAttribute) error
// 	GetVariantAttribute(id string) (*VariantAttribute, error)
// 	UpdateVariantAttribute(attr *VariantAttribute, fieldMask []string) error
// 	DeleteVariantAttribute(id string) error
// 	ListVariantAttributes(filter VariantAttributeFilter) ([]*VariantAttribute, int, error)

// 	CreateVariantAttributeValue(value *VariantAttributeValue) error
// 	GetVariantAttributeValue(id string) (*VariantAttributeValue, error)
// 	UpdateVariantAttributeValue(value *VariantAttributeValue, fieldMask []string) error
// 	DeleteVariantAttributeValue(id string) error
// 	ListVariantAttributeValues(filter VariantAttributeValueFilter) ([]*VariantAttributeValue, int, error)
// }

// ProductIncludes defines what related data to include
type ProductIncludes struct {
	Brand          bool
	Categories     bool
	Images         bool
	Variants       bool
	Tags           bool
	Specifications bool
}
