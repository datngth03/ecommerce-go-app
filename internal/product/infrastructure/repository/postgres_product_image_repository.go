package repository

import (
	"database/sql"
	"fmt"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// PostgresProductImageRepository implements the domain.ProductImageRepository interface.
type PostgresProductImageRepository struct {
	db *sql.DB
}

// NewPostgresProductImageRepository creates a new PostgresProductImageRepository.
func NewPostgresProductImageRepository(db *sql.DB) *PostgresProductImageRepository {
	return &PostgresProductImageRepository{db: db}
}

// Create inserts a new product image into the database.
func (r *PostgresProductImageRepository) Create(image *domain.ProductImage) error {
	query := `INSERT INTO product_images (product_id, url, is_primary) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(query, image.ProductID, image.URL, image.IsPrimary).Scan(&image.ID)
	if err != nil {
		return fmt.Errorf("failed to create product image: %w", err)
	}
	return nil
}

// GetByID retrieves a product image by its ID.
func (r *PostgresProductImageRepository) GetByID(id string) (*domain.ProductImage, error) {
	image := &domain.ProductImage{}
	query := `SELECT id, product_id, url, is_primary FROM product_images WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&image.ID, &image.ProductID, &image.URL, &image.IsPrimary)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProductImageNotFound
		}
		return nil, fmt.Errorf("failed to get product image: %w", err)
	}
	return image, nil
}

// Update updates an existing product image.
func (r *PostgresProductImageRepository) Update(image *domain.ProductImage) error {
	query := `UPDATE product_images SET product_id = $2, url = $3, is_primary = $4 WHERE id = $1`
	result, err := r.db.Exec(query, image.ID, image.ProductID, image.URL, image.IsPrimary)
	if err != nil {
		return fmt.Errorf("failed to update product image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProductImageNotFound
	}

	return nil
}

// Delete deletes a product image by its ID.
func (r *PostgresProductImageRepository) Delete(id string) error {
	query := `DELETE FROM product_images WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProductImageNotFound
	}

	return nil
}

// GetByProductID retrieves all images for a given product ID.
func (r *PostgresProductImageRepository) GetByProductID(productID string) ([]*domain.ProductImage, error) {
	query := `SELECT id, product_id, url, is_primary FROM product_images WHERE product_id = $1 ORDER BY is_primary DESC, id ASC`
	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product images by product ID: %w", err)
	}
	defer rows.Close()

	var images []*domain.ProductImage
	for rows.Next() {
		image := &domain.ProductImage{}
		if err := rows.Scan(&image.ID, &image.ProductID, &image.URL, &image.IsPrimary); err != nil {
			return nil, fmt.Errorf("failed to scan product image row: %w", err)
		}
		images = append(images, image)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return images, nil
}

// SetPrimaryImage updates the IsPrimary status for a product's images.
func (r *PostgresProductImageRepository) SetPrimaryImage(productID, imageID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Reset all images for the product to not primary
	resetQuery := `UPDATE product_images SET is_primary = FALSE WHERE product_id = $1`
	if _, err := tx.Exec(resetQuery, productID); err != nil {
		return fmt.Errorf("failed to reset primary image status: %w", err)
	}

	// 2. Set the specified image as primary
	setPrimaryQuery := `UPDATE product_images SET is_primary = TRUE WHERE id = $1 AND product_id = $2`
	result, err := tx.Exec(setPrimaryQuery, imageID, productID)
	if err != nil {
		return fmt.Errorf("failed to set new primary image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected for primary image: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProductImageNotFound // hoặc lỗi cụ thể hơn nếu không tìm thấy ảnh
	}

	return tx.Commit()
}
