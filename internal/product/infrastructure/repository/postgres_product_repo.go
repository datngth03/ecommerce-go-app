// internal/product/infrastructure/repository/postgres_product_repo.go
package repository

import (
	"context"
	"database/sql" // For handling JSONB array (image_urls)
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// PostgreSQLProductRepository implements the domain.ProductRepository interface
// for PostgreSQL database operations.
type PostgreSQLProductRepository struct {
	db *sql.DB
}

// NewPostgreSQLProductRepository creates a new instance of PostgreSQLProductRepository.
func NewPostgreSQLProductRepository(db *sql.DB) *PostgreSQLProductRepository {
	return &PostgreSQLProductRepository{db: db}
}

// Save creates a new product or updates an existing one in PostgreSQL.
func (r *PostgreSQLProductRepository) Save(ctx context.Context, product *domain.Product) error {
	// Convert []string to pq.StringArray for PostgreSQL array type
	imageURLsArray := pq.StringArray(product.ImageURLs)

	// Check if product exists to decide between INSERT or UPDATE
	existingProduct, err := r.FindByID(ctx, product.ID)
	if err != nil && err.Error() != "product not found" {
		return fmt.Errorf("failed to check existing product: %w", err)
	}

	if existingProduct != nil {
		// Product exists, perform UPDATE
		query := `
            UPDATE products
            SET name = $1, description = $2, price = $3, category_id = $4, image_urls = $5, stock_quantity = $6, updated_at = $7
            WHERE id = $8`
		_, err = r.db.ExecContext(ctx, query,
			product.Name, product.Description, product.Price, product.CategoryID, imageURLsArray, product.StockQuantity, time.Now(), product.ID)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}
		return nil
	}

	// Product does not exist, perform INSERT
	query := `
        INSERT INTO products (id, name, description, price, category_id, image_urls, stock_quantity, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = r.db.ExecContext(ctx, query,
		product.ID, product.Name, product.Description, product.Price, product.CategoryID, imageURLsArray, product.StockQuantity, product.CreatedAt, product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to save new product: %w", err)
	}
	return nil
}

// FindByID retrieves a product by its ID from PostgreSQL.
func (r *PostgreSQLProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	product := &domain.Product{}
	var imageURLs pq.StringArray // Use pq.StringArray for scanning
	query := `SELECT id, name, description, price, category_id, image_urls, stock_quantity, created_at, updated_at FROM products WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&product.ID, &product.Name, &product.Description, &product.Price,
		&product.CategoryID, &imageURLs, &product.StockQuantity, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to find product by ID: %w", err)
	}
	product.ImageURLs = []string(imageURLs) // Convert back to []string
	return product, nil
}

// FindAll retrieves a list of products based on filters and pagination from PostgreSQL.
func (r *PostgreSQLProductRepository) FindAll(ctx context.Context, categoryID string, limit, offset int32) ([]*domain.Product, int32, error) {
	var products []*domain.Product
	var totalCount int32

	// Build the query dynamically based on filters
	query := `SELECT id, name, description, price, category_id, image_urls, stock_quantity, created_at, updated_at FROM products`
	countQuery := `SELECT COUNT(*) FROM products`
	args := []interface{}{}
	whereClause := ""
	argCounter := 1

	if categoryID != "" {
		whereClause += fmt.Sprintf(" WHERE category_id = $%d", argCounter)
		args = append(args, categoryID)
		argCounter++
	}

	query += whereClause
	countQuery += whereClause

	// Get total count first
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total product count: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		product := &domain.Product{}
		var imageURLs pq.StringArray
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price,
			&product.CategoryID, &imageURLs, &product.StockQuantity, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product row: %w", err)
		}
		product.ImageURLs = []string(imageURLs)
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return products, totalCount, nil
}

// Delete removes a product from PostgreSQL by its ID.
func (r *PostgreSQLProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("product not found")
	}
	return nil
}

// --- Category Repository Implementation ---

// PostgreSQLCategoryRepository implements the domain.CategoryRepository interface
// for PostgreSQL database operations.
type PostgreSQLCategoryRepository struct {
	db *sql.DB
}

// NewPostgreSQLCategoryRepository creates a new instance of PostgreSQLCategoryRepository.
func NewPostgreSQLCategoryRepository(db *sql.DB) *PostgreSQLCategoryRepository {
	return &PostgreSQLCategoryRepository{db: db}
}

// Save creates a new category or updates an existing one in PostgreSQL.
func (r *PostgreSQLCategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	// Check if category exists to decide between INSERT or UPDATE
	existingCategory, err := r.FindByID(ctx, category.ID)
	if err != nil && err.Error() != "category not found" {
		return fmt.Errorf("failed to check existing category: %w", err)
	}

	if existingCategory != nil {
		// Category exists, perform UPDATE
		query := `
            UPDATE categories
            SET name = $1, description = $2, updated_at = $3
            WHERE id = $4`
		_, err = r.db.ExecContext(ctx, query,
			category.Name, category.Description, time.Now(), category.ID)
		if err != nil {
			return fmt.Errorf("failed to update category: %w", err)
		}
		return nil
	}

	// Category does not exist, perform INSERT
	query := `
        INSERT INTO categories (id, name, description, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)`
	_, err = r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		// Handle unique constraint violation for name
		if err.Error() == `pq: duplicate key value violates unique constraint "idx_categories_name"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "categories_name_key"` {
			return errors.New("category with this name already exists")
		}
		return fmt.Errorf("failed to save new category: %w", err)
	}
	return nil
}

// FindByID retrieves a category by its ID from PostgreSQL.
func (r *PostgreSQLCategoryRepository) FindByID(ctx context.Context, id string) (*domain.Category, error) {
	category := &domain.Category{}
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	err := row.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to find category by ID: %w", err)
	}
	return category, nil
}

// FindByName retrieves a category by its name from PostgreSQL.
func (r *PostgreSQLCategoryRepository) FindByName(ctx context.Context, name string) (*domain.Category, error) {
	category := &domain.Category{}
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE name = $1`
	row := r.db.QueryRowContext(ctx, query, name)

	err := row.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to find category by name: %w", err)
	}
	return category, nil
}

// FindAll retrieves a list of all categories from PostgreSQL.
func (r *PostgreSQLCategoryRepository) FindAll(ctx context.Context) ([]*domain.Category, error) {
	var categories []*domain.Category
	query := `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		category := &domain.Category{}
		err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return categories, nil
}

// Delete removes a category from PostgreSQL by its ID.
func (r *PostgreSQLCategoryRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM categories WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("category not found")
	}
	return nil
}
