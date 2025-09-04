package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) domain.ProductRepository {
	return &PostgresProductRepository{db: db}
}

// Create creates a new product
func (r *PostgresProductRepository) Create(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate UUID if not provided
	if product.ID == "" {
		product.ID = uuid.New().String()
	}

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Insert product
	query := `
		INSERT INTO products (id, brand_id, name, slug, description, rating, review_count, created_at, updated_at)
		VALUES (:id, :brand_id, :name, :slug, :description, :rating, :review_count, :created_at, :updated_at)`

	_, err = tx.ExecContext(ctx, query, product)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Insert product categories if provided
	if len(product.Categories) > 0 {
		categoryIDs := make([]string, len(product.Categories))
		for i, category := range product.Categories {
			categoryIDs[i] = category.ID
		}
		if err := r.addCategoriesToProductTx(ctx, tx, product.ID, categoryIDs); err != nil {
			return fmt.Errorf("failed to add categories: %w", err)
		}
	}

	// Insert product tags if provided
	if len(product.Tags) > 0 {
		tagIDs := make([]string, len(product.Tags))
		for i, tag := range product.Tags {
			tagIDs[i] = tag.ID
		}
		if err := r.addTagsToProductTx(ctx, tx, product.ID, tagIDs); err != nil {
			return fmt.Errorf("failed to add tags: %w", err)
		}
	}

	// Insert product images if provided
	if len(product.Images) > 0 {
		if err := r.addImagesToProductTx(ctx, tx, product.ID, product.Images); err != nil {
			return fmt.Errorf("failed to add images: %w", err)
		}
	}

	// // Insert product specifications if provided
	// if len(product.Specifications) > 0 {
	// 	if err := r.addSpecificationsToProductTx(ctx, tx, product.ID, product.Specifications); err != nil {
	// 		return fmt.Errorf("failed to add specifications: %w", err)
	// 	}
	// }

	return tx.Commit()
}

// GetByID retrieves a product by ID
func (r *PostgresProductRepository) GetByID(id string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product domain.Product
	query := `
        SELECT id, brand_id, name, slug, description, rating, review_count, created_at, updated_at
        FROM products
        WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&product.ID,
		&product.BrandID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Rating,
		&product.ReviewCount,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// GetBySlug retrieves a product by slug
// GetBySlug retrieves a product by slug
func (r *PostgresProductRepository) GetBySlug(slug string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var product domain.Product
	query := `
		SELECT id, brand_id, name, slug, description, rating, review_count, created_at, updated_at
		FROM products
		WHERE slug = $1`

	row := r.db.QueryRowContext(ctx, query, slug)
	err := row.Scan(
		&product.ID,
		&product.BrandID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Rating,
		&product.ReviewCount,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// Update updates a product
// Update updates a product
func (r *PostgresProductRepository) Update(product *domain.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product.UpdatedAt = time.Now()

	query := `
		UPDATE products
		SET brand_id = $1,
		    name = $2,
		    slug = $3,
		    description = $4,
		    rating = $5,
		    review_count = $6,
		    updated_at = $7
		WHERE id = $8`

	result, err := r.db.ExecContext(ctx, query,
		product.BrandID,
		product.Name,
		product.Slug,
		product.Description,
		product.Rating,
		product.ReviewCount,
		product.UpdatedAt,
		product.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found: %s", product.ID)
	}

	return nil
}

// Delete deletes a product
func (r *PostgresProductRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sử dụng BeginTx từ thư viện chuẩn database/sql
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Xóa các bản ghi liên quan trước (do ràng buộc khóa ngoại)
	// Trình tự này rất quan trọng để tránh lỗi
	queries := []string{
		"DELETE FROM product_categories WHERE product_id = $1",
		"DELETE FROM product_tags WHERE product_id = $1",
		"DELETE FROM product_images WHERE product_id = $1",
		"DELETE FROM product_specifications WHERE product_id = $1",
		"DELETE FROM product_variants WHERE product_id = $1",
		"DELETE FROM products WHERE id = $1",
	}

	for _, query := range queries {
		_, err := tx.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete product related data: %w", err)
		}
	}

	// Commit giao dịch
	return tx.Commit()
}

// List retrieves products with filtering and pagination
func (r *PostgresProductRepository) List(filter domain.ProductFilter) ([]*domain.Product, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build WHERE clause
	whereClause, args := r.buildProductWhereClause(filter)

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT p.id)
		FROM products p
		LEFT JOIN product_categories pc ON p.id = pc.product_id
		LEFT JOIN product_tags pt ON p.id = pt.product_id
		LEFT JOIN brands b ON p.brand_id = b.id
		%s`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Main query
	orderClause := r.buildOrderClause(filter.SortBy, filter.SortOrder)

	mainQuery := fmt.Sprintf(`
		SELECT DISTINCT p.id, p.brand_id, p.name, p.slug, p.description, 
		       p.rating, p.review_count, p.created_at, p.updated_at
		FROM products p
		LEFT JOIN product_categories pc ON p.id = pc.product_id
		LEFT JOIN product_tags pt ON p.id = pt.product_id
		LEFT JOIN brands b ON p.brand_id = b.id
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(args)+1, len(args)+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID,
			&p.BrandID,
			&p.Name,
			&p.Slug,
			&p.Description,
			&p.Rating,
			&p.ReviewCount,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %w", err)
	}

	// Load related data (brands, categories, images, etc.)
	if len(products) > 0 {
		if filter.IncludeBrand {
			if err := r.loadBrandsForProducts(ctx, products); err != nil {
				return nil, 0, err
			}
		}
		if filter.IncludeCategories {
			if err := r.loadCategoriesForProducts(ctx, products); err != nil {
				return nil, 0, err
			}
		}
		if filter.IncludeImages {
			if err := r.loadImagesForProducts(ctx, products); err != nil {
				return nil, 0, err
			}
		}
		if filter.IncludeVariants {
			if err := r.loadVariantsForProducts(ctx, products); err != nil {
				return nil, 0, err
			}
		}
		if filter.IncludeTags {
			if err := r.loadTagsForProducts(ctx, products); err != nil {
				return nil, 0, err
			}
		}
		if filter.IncludeSpecifications {
			if err := r.loadSpecificationsForProducts(ctx, products); err != nil {
				return nil, 0, err
			}
		}
	}

	return products, total, nil
}

// AddCategories adds categories to a product
func (r *PostgresProductRepository) AddCategories(productID string, categoryIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sử dụng BeginTx từ thư viện chuẩn database/sql
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Gọi hàm helper để thêm các danh mục vào sản phẩm trong cùng một transaction
	err = r.addCategoriesToProductTx(ctx, tx, productID, categoryIDs)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RemoveCategories removes categories from a product
func (r *PostgresProductRepository) RemoveCategories(productID string, categoryIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM product_categories WHERE product_id = $1 AND category_id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, productID, pq.Array(categoryIDs))
	if err != nil {
		return fmt.Errorf("failed to remove categories: %w", err)
	}

	return nil
}

// GetCategories retrieves categories for a product
func (r *PostgresProductRepository) GetCategories(productID string) ([]*domain.Category, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT c.id, c.name, c.slug, c.description, c.image, c.parent_id, c.created_at, c.updated_at
		FROM categories c
		INNER JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = $1
		ORDER BY c.name`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var c domain.Category
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Slug,
			&c.Description,
			&c.Image,
			&c.ParentID,
			&c.CreatedAt,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return categories, nil
}

// AddTags adds tags to a product
func (r *PostgresProductRepository) addTagsToProductTx(ctx context.Context, tx *sql.Tx, productID string, tagIDs []string) error {
	// Xóa các thẻ hiện có của sản phẩm trước khi thêm mới
	deleteQuery := "DELETE FROM product_tags WHERE product_id = $1"
	_, err := tx.ExecContext(ctx, deleteQuery, productID)
	if err != nil {
		return fmt.Errorf("failed to delete existing tags: %w", err)
	}

	// Chèn các thẻ mới
	insertQuery := "INSERT INTO product_tags (product_id, tag_id) VALUES ($1, $2)"
	for _, tagID := range tagIDs {
		_, err := tx.ExecContext(ctx, insertQuery, productID, tagID)
		if err != nil {
			return fmt.Errorf("failed to insert tag %s for product %s: %w", tagID, productID, err)
		}
	}

	return nil
}

func (r *PostgresProductRepository) AddTags(productID string, tagIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Xóa các tag cũ trước
	_, err = tx.ExecContext(ctx, `DELETE FROM product_tags WHERE product_id = $1`, productID)
	if err != nil {
		return fmt.Errorf("failed to clear old tags: %w", err)
	}

	// Thêm tag mới
	query := `INSERT INTO product_tags (product_id, tag_id) VALUES ($1, $2)`
	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx, query, productID, tagID); err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tagID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit AddTags transaction: %w", err)
	}

	return nil
}

// RemoveTags removes tags from a product
func (r *PostgresProductRepository) RemoveTags(productID string, tagIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM product_tags WHERE product_id = $1 AND tag_id = ANY($2)`
	_, err := r.db.ExecContext(ctx, query, productID, pq.Array(tagIDs))
	if err != nil {
		return fmt.Errorf("failed to remove tags: %w", err)
	}

	return nil
}

// GetTags retrieves tags for a product
func (r *PostgresProductRepository) GetTags(productID string) ([]*domain.Tag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT t.id, t.name, t.slug
		FROM tags t
		INNER JOIN product_tags pt ON t.id = pt.tag_id
		WHERE pt.product_id = $1
		ORDER BY t.name`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return tags, nil
}

// Helper methods
func (r *PostgresProductRepository) buildProductWhereClause(filter domain.ProductFilter) (string, []interface{}) {
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d OR b.name ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if len(filter.CategoryIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("pc.category_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(filter.CategoryIDs))
		argIndex++
	}

	if len(filter.TagIDs) > 0 {
		conditions = append(conditions, fmt.Sprintf("pt.tag_id = ANY($%d)", argIndex))
		args = append(args, pq.Array(filter.TagIDs))
		argIndex++
	}

	if filter.MinRating != nil {
		conditions = append(conditions, fmt.Sprintf("p.rating >= $%d", argIndex))
		args = append(args, *filter.MinRating)
		argIndex++
	}

	// Price filtering would require joining with variants table
	if filter.MinPrice != nil || filter.MaxPrice != nil {
		conditions = append(conditions, "EXISTS (SELECT 1 FROM product_variants pv WHERE pv.product_id = p.id")

		if filter.MinPrice != nil {
			conditions[len(conditions)-1] += fmt.Sprintf(" AND pv.price >= $%d", argIndex)
			args = append(args, *filter.MinPrice)
			argIndex++
		}

		if filter.MaxPrice != nil {
			conditions[len(conditions)-1] += fmt.Sprintf(" AND pv.price <= $%d", argIndex)
			args = append(args, *filter.MaxPrice)
			argIndex++
		}

		conditions[len(conditions)-1] += ")"
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *PostgresProductRepository) buildOrderClause(sortBy string, sortOrder domain.SortOrder) string {
	var column string
	switch sortBy {
	case "name":
		column = "p.name"
	case "rating":
		column = "p.rating"
	case "created_at":
		column = "p.created_at"
	case "price":
		// For price sorting, we'd need to join with variants and use MIN/MAX
		column = "(SELECT MIN(pv.price) FROM product_variants pv WHERE pv.product_id = p.id)"
	default:
		column = "p.created_at"
	}

	order := "DESC"
	if sortOrder == domain.SortOrderAsc {
		order = "ASC"
	}

	return fmt.Sprintf("ORDER BY %s %s", column, order)
}

// Transaction helper methods

func (r *PostgresProductRepository) addCategoriesToProductTx(ctx context.Context, tx *sql.Tx, productID string, categoryIDs []string) error {
	if len(categoryIDs) == 0 {
		return nil
	}

	// Remove existing categories first to avoid duplicates
	_, err := tx.ExecContext(ctx, "DELETE FROM product_categories WHERE product_id = $1", productID)
	if err != nil {
		return fmt.Errorf("failed to remove existing categories: %w", err)
	}

	// Insert new categories
	valueStrings := make([]string, 0, len(categoryIDs))
	valueArgs := make([]interface{}, 0, len(categoryIDs)*2)

	for i, categoryID := range categoryIDs {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		valueArgs = append(valueArgs, productID, categoryID)
	}

	query := fmt.Sprintf("INSERT INTO product_categories (product_id, category_id) VALUES %s",
		strings.Join(valueStrings, ","))

	_, err = tx.ExecContext(ctx, query, valueArgs...)
	if err != nil {
		return fmt.Errorf("failed to insert categories: %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) addImagesToProductTx(ctx context.Context, tx *sql.Tx, productID string, images []*domain.ProductImage) error {
	if len(images) == 0 {
		return nil
	}

	for _, img := range images {
		if img.ID == "" {
			img.ID = uuid.New().String()
		}
		img.ProductID = productID

		query := `INSERT INTO product_images (id, product_id, url, is_primary) VALUES ($1, $2, $3, $4)`
		_, err := tx.ExecContext(ctx, query, img.ID, img.ProductID, img.URL, img.IsPrimary)
		if err != nil {
			return fmt.Errorf("failed to insert image: %w", err)
		}
	}

	return nil
}

func (r *PostgresProductRepository) loadBrandsForProducts(ctx context.Context, products []*domain.Product) error {
	brandIDs := make([]string, 0, len(products))
	brandMap := make(map[string]*domain.Brand)

	// Collect brand IDs
	for _, product := range products {
		if product.BrandID != nil && *product.BrandID != "" {
			if _, exists := brandMap[*product.BrandID]; !exists {
				brandIDs = append(brandIDs, *product.BrandID)
				brandMap[*product.BrandID] = nil
			}
		}
	}

	if len(brandIDs) == 0 {
		return nil
	}

	// Query brands
	query := `
		SELECT id, name, slug, logo, description, created_at, updated_at
		FROM brands WHERE id = ANY($1)
	`
	rows, err := r.db.QueryContext(ctx, query, pq.Array(brandIDs))
	if err != nil {
		return fmt.Errorf("failed to load brands: %w", err)
	}
	defer rows.Close()

	var brands []*domain.Brand
	for rows.Next() {
		var b domain.Brand
		err := rows.Scan(
			&b.ID,
			&b.Name,
			&b.Slug,
			&b.Logo,
			&b.Description,
			&b.CreatedAt,
			&b.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan brand: %w", err)
		}
		brands = append(brands, &b)
	}

	// Map brands
	for _, brand := range brands {
		brandMap[brand.ID] = brand
	}

	// Assign to products
	for _, product := range products {
		if product.BrandID != nil {
			product.Brand = brandMap[*product.BrandID]
		}
	}

	return nil
}

func (r *PostgresProductRepository) loadCategoriesForProducts(ctx context.Context, products []*domain.Product) error {
	productIDs := make([]string, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	query := `
		SELECT pc.product_id, c.id, c.name, c.slug, c.description, c.image, c.parent_id, c.created_at, c.updated_at
		FROM categories c
		INNER JOIN product_categories pc ON c.id = pc.category_id
		WHERE pc.product_id = ANY($1)
		ORDER BY c.name`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}
	defer rows.Close()

	productCategories := make(map[string][]*domain.Category)
	for rows.Next() {
		var productID string
		var category domain.Category

		err := rows.Scan(
			&productID, &category.ID, &category.Name, &category.Slug,
			&category.Description, &category.Image, &category.ParentID,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan category: %w", err)
		}

		productCategories[productID] = append(productCategories[productID], &category)
	}

	// Assign categories to products
	for _, product := range products {
		product.Categories = productCategories[product.ID]
	}

	return nil
}

func (r *PostgresProductRepository) loadImagesForProducts(ctx context.Context, products []*domain.Product) error {
	productIDs := make([]string, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	query := `
		SELECT product_id, id, url, is_primary
		FROM product_images
		WHERE product_id = ANY($1)
		ORDER BY is_primary DESC, id`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return fmt.Errorf("failed to load images: %w", err)
	}
	defer rows.Close()

	productImages := make(map[string][]*domain.ProductImage)
	for rows.Next() {
		var productID string
		var image domain.ProductImage

		err := rows.Scan(&productID, &image.ID, &image.URL, &image.IsPrimary)
		if err != nil {
			return fmt.Errorf("failed to scan image: %w", err)
		}

		image.ProductID = productID
		productImages[productID] = append(productImages[productID], &image)
	}

	// Assign images to products
	for _, product := range products {
		product.Images = productImages[product.ID]
	}

	return nil
}

func (r *PostgresProductRepository) loadVariantsForProducts(ctx context.Context, products []*domain.Product) error {
	productIDs := make([]string, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	query := `
		SELECT product_id, id, sku, price, original_price, discount, created_at, updated_at
		FROM product_variants
		WHERE product_id = ANY($1)
		ORDER BY price`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return fmt.Errorf("failed to load variants: %w", err)
	}
	defer rows.Close()

	productVariants := make(map[string][]*domain.ProductVariant)
	for rows.Next() {
		var productID string
		var variant domain.ProductVariant

		err := rows.Scan(
			&productID, &variant.ID, &variant.SKU, &variant.Price,
			&variant.OriginalPrice, &variant.Discount, &variant.CreatedAt, &variant.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan variant: %w", err)
		}

		variant.ProductID = productID
		productVariants[productID] = append(productVariants[productID], &variant)
	}

	// Assign variants to products
	for _, product := range products {
		product.Variants = productVariants[product.ID]
	}

	return nil
}

func (r *PostgresProductRepository) loadTagsForProducts(ctx context.Context, products []*domain.Product) error {
	productIDs := make([]string, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	query := `
		SELECT pt.product_id, t.id, t.name, t.slug
		FROM tags t
		INNER JOIN product_tags pt ON t.id = pt.tag_id
		WHERE pt.product_id = ANY($1)
		ORDER BY t.name`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	defer rows.Close()

	productTags := make(map[string][]*domain.Tag)
	for rows.Next() {
		var productID string
		var tag domain.Tag

		err := rows.Scan(&productID, &tag.ID, &tag.Name, &tag.Slug)
		if err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}

		productTags[productID] = append(productTags[productID], &tag)
	}

	// Assign tags to products
	for _, product := range products {
		product.Tags = productTags[product.ID]
	}

	return nil
}

func (r *PostgresProductRepository) loadSpecificationsForProducts(ctx context.Context, products []*domain.Product) error {
	productIDs := make([]string, len(products))
	for i, p := range products {
		productIDs[i] = p.ID
	}

	query := `
		SELECT ps.product_id, ps.id, ps.attribute_id, ps.value,
			   sa.name, sa.slug
		FROM product_specifications ps
		INNER JOIN specification_attributes sa ON ps.attribute_id = sa.id
		WHERE ps.product_id = ANY($1)
		ORDER BY sa.name`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(productIDs))
	if err != nil {
		return fmt.Errorf("failed to load specifications: %w", err)
	}
	defer rows.Close()

	// Map để group specifications theo product_id
	productSpecs := make(map[string][]*domain.ProductSpecification)

	for rows.Next() {
		var (
			productID   string
			specID      string
			attributeID string
			value       string
			attrName    string
			attrSlug    string
		)

		err := rows.Scan(&productID, &specID, &attributeID, &value, &attrName, &attrSlug)
		if err != nil {
			return fmt.Errorf("failed to scan specification: %w", err)
		}

		spec := &domain.ProductSpecification{
			ID:          specID,
			ProductID:   productID,
			AttributeID: attributeID,
			Value:       value,
			Attribute: &domain.SpecificationAttribute{
				ID:   attributeID,
				Name: attrName,
				Slug: attrSlug,
			},
		}

		productSpecs[productID] = append(productSpecs[productID], spec)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over specification rows: %w", err)
	}

	// Gán specifications cho từng product
	for _, product := range products {
		if specs, exists := productSpecs[product.ID]; exists {
			product.Specifications = specs
		}
	}

	return nil
}
