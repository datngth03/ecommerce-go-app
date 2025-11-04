// services/product-service/internal/repository/product_postgres.go

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
)

// ProductPostgresRepository implements ProductRepository for PostgreSQL
type ProductPostgresRepository struct {
	db *sql.DB
}

// CategoryPostgresRepository implements CategoryRepository for PostgreSQL
type CategoryPostgresRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new PostgreSQL product repository
func NewProductRepository(db *sql.DB) ProductRepository {
	return &ProductPostgresRepository{db: db}
}

// NewCategoryRepository creates a new PostgreSQL category repository
func NewCategoryRepository(db *sql.DB) CategoryRepository {
	return &CategoryPostgresRepository{db: db}
}

// =================== PRODUCT REPOSITORY IMPLEMENTATION ===================

// Create creates a new product in the database
func (r *ProductPostgresRepository) Create(ctx context.Context, product *models.Product) error {
	start := time.Now()
	defer func() {
		metrics.RecordDBQuery("INSERT", "products", "success", time.Since(start))
	}()

	product.ID = uuid.New().String()
	product.GenerateSlug()
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now
	product.IsActive = true

	query := `
		INSERT INTO products (id, name, slug, description, price, category_id, image_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		product.ID, product.Name, product.Slug, product.Description,
		product.Price, product.CategoryID, product.ImageURL, product.IsActive,
		product.CreatedAt, product.UpdatedAt,
	)

	if err != nil {
		metrics.RecordDBQuery("INSERT", "products", "error", time.Since(start))
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				if strings.Contains(pqErr.Message, "slug") {
					return fmt.Errorf("product with slug already exists")
				}
				return fmt.Errorf("product already exists")
			case "23503": // foreign key violation
				return fmt.Errorf("category not found")
			}
		}
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID
func (r *ProductPostgresRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	start := time.Now()
	defer func() {
		metrics.RecordDBQuery("SELECT", "products", "success", time.Since(start))
	}()

	query := `
		SELECT p.id, p.name, p.slug, p.description, p.price, p.category_id, 
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`

	product := &models.Product{}
	category := &models.Category{}
	var categoryID, categoryName, categorySlug sql.NullString
	var categoryCreatedAt, categoryUpdatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description,
		&product.Price, &product.CategoryID, &product.ImageURL, &product.IsActive,
		&product.CreatedAt, &product.UpdatedAt,
		&categoryID, &categoryName, &categorySlug, &categoryCreatedAt, &categoryUpdatedAt,
	)

	if err != nil {
		metrics.RecordDBQuery("SELECT", "products", "error", time.Since(start))
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Populate category if exists
	if categoryID.Valid {
		category.ID = categoryID.String
		category.Name = categoryName.String
		category.Slug = categorySlug.String
		category.CreatedAt = categoryCreatedAt.Time
		category.UpdatedAt = categoryUpdatedAt.Time
		product.Category = category
	}

	return product, nil
}

// GetBySlug retrieves a product by slug
func (r *ProductPostgresRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.price, p.category_id, 
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.slug = $1
	`

	product := &models.Product{}
	category := &models.Category{}
	var categoryID, categoryName, categorySlug sql.NullString
	var categoryCreatedAt, categoryUpdatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description,
		&product.Price, &product.CategoryID, &product.ImageURL, &product.IsActive,
		&product.CreatedAt, &product.UpdatedAt,
		&categoryID, &categoryName, &categorySlug, &categoryCreatedAt, &categoryUpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Populate category if exists
	if categoryID.Valid {
		category.ID = categoryID.String
		category.Name = categoryName.String
		category.Slug = categorySlug.String
		category.CreatedAt = categoryCreatedAt.Time
		category.UpdatedAt = categoryUpdatedAt.Time
		product.Category = category
	}

	return product, nil
}

// Update updates an existing product
func (r *ProductPostgresRepository) Update(ctx context.Context, product *models.Product) error {
	product.GenerateSlug()
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products 
		SET name = $2, slug = $3, description = $4, price = $5, 
		    category_id = $6, image_url = $7, is_active = $8, updated_at = $9
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		product.ID, product.Name, product.Slug, product.Description,
		product.Price, product.CategoryID, product.ImageURL, product.IsActive,
		product.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				return fmt.Errorf("product with slug already exists")
			case "23503": // foreign key violation
				return fmt.Errorf("category not found")
			}
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Delete deletes a product by ID
func (r *ProductPostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// List retrieves a paginated list of products
func (r *ProductPostgresRepository) List(ctx context.Context, req *models.ListProductsRequest) ([]models.Product, int64, error) {
	// Convert empty category_id to nil for proper SQL handling
	var categoryIDParam interface{}
	if req.CategoryID == "" {
		categoryIDParam = nil
	} else {
		categoryIDParam = req.CategoryID
	}

	// Count total products
	countQuery := `SELECT COUNT(*) FROM products WHERE ($1::uuid IS NULL OR category_id = $1::uuid)`

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, categoryIDParam).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Query products with pagination
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.price, p.category_id, 
		       p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name, c.slug, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE ($1::uuid IS NULL OR p.category_id = $1::uuid)
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, categoryIDParam, req.PageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		product := models.Product{}
		category := models.Category{}
		var categoryID, categoryName, categorySlug sql.NullString
		var categoryCreatedAt, categoryUpdatedAt sql.NullTime

		err := rows.Scan(
			&product.ID, &product.Name, &product.Slug, &product.Description,
			&product.Price, &product.CategoryID, &product.ImageURL, &product.IsActive,
			&product.CreatedAt, &product.UpdatedAt,
			&categoryID, &categoryName, &categorySlug, &categoryCreatedAt, &categoryUpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Populate category if exists
		if categoryID.Valid {
			category.ID = categoryID.String
			category.Name = categoryName.String
			category.Slug = categorySlug.String
			category.CreatedAt = categoryCreatedAt.Time
			category.UpdatedAt = categoryUpdatedAt.Time
			product.Category = &category
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate products: %w", err)
	}

	return products, total, nil
}

// ListByCategoryID retrieves products by category ID
func (r *ProductPostgresRepository) ListByCategoryID(ctx context.Context, categoryID string, req *models.ListProductsRequest) ([]models.Product, int64, error) {
	req.CategoryID = categoryID
	return r.List(ctx, req)
}

// ExistsByName checks if a product exists by name
func (r *ProductPostgresRepository) ExistsByName(ctx context.Context, name string, excludeID ...string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE name = $1`
	args := []interface{}{name}

	if len(excludeID) > 0 && excludeID[0] != "" {
		query += ` AND id != $2`
		args = append(args, excludeID[0])
	}
	query += `)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}

// CountByCategory counts products in a category
func (r *ProductPostgresRepository) CountByCategory(ctx context.Context, categoryID string) (int64, error) {
	query := `SELECT COUNT(*) FROM products WHERE category_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products by category: %w", err)
	}

	return count, nil
}

// =================== CATEGORY REPOSITORY IMPLEMENTATION ===================

// Create creates a new category in the database
func (r *CategoryPostgresRepository) Create(ctx context.Context, category *models.Category) error {
	category.ID = uuid.New().String()
	category.GenerateSlug()
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `
		INSERT INTO categories (id, name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Slug, category.CreatedAt, category.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				if strings.Contains(pqErr.Message, "name") {
					return fmt.Errorf("category with name already exists")
				}
				if strings.Contains(pqErr.Message, "slug") {
					return fmt.Errorf("category with slug already exists")
				}
				return fmt.Errorf("category already exists")
			}
		}
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// GetByID retrieves a category by ID
func (r *CategoryPostgresRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	query := `SELECT id, name, slug, created_at, updated_at FROM categories WHERE id = $1`

	category := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Slug,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// GetBySlug retrieves a category by slug
func (r *CategoryPostgresRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	query := `SELECT id, name, slug, created_at, updated_at FROM categories WHERE slug = $1`

	category := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&category.ID, &category.Name, &category.Slug,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// Update updates an existing category
func (r *CategoryPostgresRepository) Update(ctx context.Context, category *models.Category) error {
	category.GenerateSlug()
	category.UpdatedAt = time.Now()

	query := `
		UPDATE categories 
		SET name = $2, slug = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Slug, category.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				if strings.Contains(pqErr.Message, "name") {
					return fmt.Errorf("category with name already exists")
				}
				return fmt.Errorf("category with slug already exists")
			}
		}
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// Delete deletes a category by ID
func (r *CategoryPostgresRepository) Delete(ctx context.Context, id string) error {
	// Check if category has products
	countQuery := `SELECT COUNT(*) FROM products WHERE category_id = $1`
	var productCount int64
	err := r.db.QueryRowContext(ctx, countQuery, id).Scan(&productCount)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if productCount > 0 {
		return fmt.Errorf("cannot delete category: it contains %d products", productCount)
	}

	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// List retrieves all categories
func (r *CategoryPostgresRepository) List(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, slug, created_at, updated_at 
		FROM categories 
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category

	for rows.Next() {
		category := models.Category{}
		err := rows.Scan(
			&category.ID, &category.Name, &category.Slug,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate categories: %w", err)
	}

	return categories, nil
}

// ExistsByName checks if a category exists by name
func (r *CategoryPostgresRepository) ExistsByName(ctx context.Context, name string, excludeID ...string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1`
	args := []interface{}{name}

	if len(excludeID) > 0 && excludeID[0] != "" {
		query += ` AND id != $2`
		args = append(args, excludeID[0])
	}
	query += `)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return exists, nil
}

// ExistsByID checks if a category exists by ID
func (r *CategoryPostgresRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return exists, nil
}
