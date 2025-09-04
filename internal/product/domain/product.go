package domain

import (
	"time"
)

// Brand represents a product brand
type Brand struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Logo        *string   `json:"logo,omitempty" db:"logo"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Category represents a product category with hierarchical support
type Category struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Image       *string   `json:"image,omitempty" db:"image"`
	ParentID    *string   `json:"parent_id,omitempty" db:"parent_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relations (not stored in database)
	Parent   *Category   `json:"parent,omitempty" db:"-"`
	Children []*Category `json:"children,omitempty" db:"-"`
}

// Product represents a product entity
type Product struct {
	ID          string    `json:"id" db:"id"`
	BrandID     *string   `json:"brand_id,omitempty" db:"brand_id"`
	Name        string    `json:"name" db:"name"`
	Slug        string    `json:"slug" db:"slug"`
	Description *string   `json:"description,omitempty" db:"description"`
	Rating      float64   `json:"rating" db:"rating"`
	ReviewCount int32     `json:"review_count" db:"review_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relations (not stored in database)
	Brand          *Brand                  `json:"brand,omitempty" db:"-"`
	Categories     []*Category             `json:"categories,omitempty" db:"-"`
	Images         []*ProductImage         `json:"images,omitempty" db:"-"`
	Variants       []*ProductVariant       `json:"variants,omitempty" db:"-"`
	Tags           []*Tag                  `json:"tags,omitempty" db:"-"`
	Specifications []*ProductSpecification `json:"specifications,omitempty" db:"-"`
}

// ProductImage represents a product image
type ProductImage struct {
	ID        string `json:"id" db:"id"`
	ProductID string `json:"product_id" db:"product_id"`
	URL       string `json:"url" db:"url"`
	IsPrimary bool   `json:"is_primary" db:"is_primary"`
}

// ProductVariant represents a product variant with pricing
type ProductVariant struct {
	ID            string    `json:"id" db:"id"`
	ProductID     string    `json:"product_id" db:"product_id"`
	SKU           string    `json:"sku" db:"sku"`
	Price         float64   `json:"price" db:"price"`
	OriginalPrice float64   `json:"original_price" db:"original_price"`
	Discount      *float64  `json:"discount,omitempty" db:"discount"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// Relations (not stored in database)
	AttributeValues []*VariantAttributeValue `json:"attribute_values,omitempty" db:"-"`
}

// Tag represents a product tag
type Tag struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Slug string `json:"slug" db:"slug"`
}

// SpecificationAttribute represents a specification attribute definition
type SpecificationAttribute struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Slug string `json:"slug" db:"slug"`
}

// ProductSpecification represents a product specification value
type ProductSpecification struct {
	ID          string `json:"id" db:"id"`
	ProductID   string `json:"product_id" db:"product_id"`
	AttributeID string `json:"attribute_id" db:"attribute_id"`
	Value       string `json:"value" db:"value"`

	// Relations (not stored in database)
	Attribute *SpecificationAttribute `json:"attribute,omitempty" db:"-"`
}

// VariantAttribute represents a variant attribute definition (e.g., Color, Size)
type VariantAttribute struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Slug string `json:"slug" db:"slug"`
}

// VariantAttributeValue represents a variant attribute value (e.g., Red, Large)
type VariantAttributeValue struct {
	ID          string `json:"id" db:"id"`
	AttributeID string `json:"attribute_id" db:"attribute_id"`
	Value       string `json:"value" db:"value"`

	// Relations (not stored in database)
	Attribute *VariantAttribute `json:"attribute,omitempty" db:"-"`
}

// Junction tables for many-to-many relationships

// ProductCategory represents the product-category relationship
type ProductCategory struct {
	ProductID  string `json:"product_id" db:"product_id"`
	CategoryID string `json:"category_id" db:"category_id"`
}

// ProductTag represents the product-tag relationship
type ProductTag struct {
	ProductID string `json:"product_id" db:"product_id"`
	TagID     string `json:"tag_id" db:"tag_id"`
}

// ProductVariantOption represents the variant-attribute-value relationship
type ProductVariantOption struct {
	VariantID        string `json:"variant_id" db:"variant_id"`
	AttributeValueID string `json:"attribute_value_id" db:"attribute_value_id"`
}

// // Domain methods

// // IsRoot checks if category is a root category
// func (c *Category) IsRoot() bool {
// 	return c.ParentID == nil
// }

// // HasDiscount checks if product variant has a discount
// func (pv *ProductVariant) HasDiscount() bool {
// 	return pv.Discount != nil && *pv.Discount > 0
// }

// // GetDiscountedPrice calculates the final price after discount
// func (pv *ProductVariant) GetDiscountedPrice() float64 {
// 	if !pv.HasDiscount() {
// 		return pv.Price
// 	}
// 	return pv.Price * (1 - *pv.Discount/100)
// }

// // GetSavings calculates the amount saved with discount
// func (pv *ProductVariant) GetSavings() float64 {
// 	if !pv.HasDiscount() {
// 		return 0
// 	}
// 	return pv.OriginalPrice - pv.GetDiscountedPrice()
// }

// // HasBrand checks if product has a brand
// func (p *Product) HasBrand() bool {
// 	return p.BrandID != nil && *p.BrandID != ""
// }

// // GetPrimaryImage returns the primary image of the product
// func (p *Product) GetPrimaryImage() *ProductImage {
// 	for _, img := range p.Images {
// 		if img.IsPrimary {
// 			return img
// 		}
// 	}
// 	// Return first image if no primary image is set
// 	if len(p.Images) > 0 {
// 		return p.Images[0]
// 	}
// 	return nil
// }

// // GetPriceRange returns the min and max price of product variants
// func (p *Product) GetPriceRange() (float64, float64) {
// 	if len(p.Variants) == 0 {
// 		return 0, 0
// 	}

// 	min := p.Variants[0].GetDiscountedPrice()
// 	max := min

// 	for _, variant := range p.Variants {
// 		price := variant.GetDiscountedPrice()
// 		if price < min {
// 			min = price
// 		}
// 		if price > max {
// 			max = price
// 		}
// 	}

// 	return min, max
// }

// // HasVariants checks if product has variants
// func (p *Product) HasVariants() bool {
// 	return len(p.Variants) > 0
// }

// // IsInStock checks if any variant is available (this would need inventory data)
// // This is a placeholder - you'd need to integrate with inventory service
// func (p *Product) IsInStock() bool {
// 	// TODO: Implement inventory check
// 	return p.HasVariants()
// }

// // Repository interfaces for dependency injection
