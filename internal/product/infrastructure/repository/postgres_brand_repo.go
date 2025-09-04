package repository

import (
	// "context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

type PostgresBrandRepository struct {
	db *sql.DB
}

func NewPostgresBrandRepository(db *sql.DB) *PostgresBrandRepository {
	return &PostgresBrandRepository{db: db}
}

func (r *PostgresBrandRepository) Create(brand *domain.Brand) error {
	query := `
		INSERT INTO brands (id, name, slug, logo, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	now := time.Now()
	brand.CreatedAt = now
	brand.UpdatedAt = now

	_, err := r.db.Exec(query, brand.ID, brand.Name, brand.Slug, brand.Logo, brand.Description, brand.CreatedAt, brand.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create brand: %w", err)
	}

	return nil
}

func (r *PostgresBrandRepository) GetByID(id string) (*domain.Brand, error) {
	query := `
		SELECT id, name, slug, logo, description, created_at, updated_at
		FROM brands
		WHERE id = $1`

	brand := &domain.Brand{}
	err := r.db.QueryRow(query, id).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.Logo,
		&brand.Description,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrBrandNotFound
		}
		return nil, fmt.Errorf("failed to get brand by ID: %w", err)
	}

	return brand, nil
}

func (r *PostgresBrandRepository) GetBySlug(slug string) (*domain.Brand, error) {
	query := `
		SELECT id, name, slug, logo, description, created_at, updated_at
		FROM brands
		WHERE slug = $1`

	brand := &domain.Brand{}
	err := r.db.QueryRow(query, slug).Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.Logo,
		&brand.Description,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrBrandNotFound
		}
		return nil, fmt.Errorf("failed to get brand by slug: %w", err)
	}

	return brand, nil
}

func (r *PostgresBrandRepository) Update(brand *domain.Brand) error {
	query := `
		UPDATE brands 
		SET name = $2, slug = $3, logo = $4, description = $5, updated_at = $6
		WHERE id = $1`

	brand.UpdatedAt = time.Now()

	result, err := r.db.Exec(query, brand.ID, brand.Name, brand.Slug, brand.Logo, brand.Description, brand.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update brand: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrBrandNotFound
	}

	return nil
}

func (r *PostgresBrandRepository) Delete(id string) error {
	query := `DELETE FROM brands WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete brand: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrBrandNotFound
	}

	return nil
}

func (r *PostgresBrandRepository) List(filter domain.BrandFilter) ([]*domain.Brand, int, error) {
	var brands []*domain.Brand
	var totalCount int

	// Build where clause
	whereClause, args := r.buildWhereClause(filter)

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM brands %s`, whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count brands: %w", err)
	}

	// Main query with pagination
	orderClause := r.buildOrderClause(filter.SortBy, filter.SortOrder)
	query := fmt.Sprintf(`
		SELECT id, name, slug, logo, description, created_at, updated_at
		FROM brands %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(args)+1, len(args)+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list brands: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		brand := &domain.Brand{}
		err := rows.Scan(
			&brand.ID,
			&brand.Name,
			&brand.Slug,
			&brand.Logo,
			&brand.Description,
			&brand.CreatedAt,
			&brand.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan brand: %w", err)
		}

		brands = append(brands, brand)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating brand rows: %w", err)
	}

	return brands, totalCount, nil
}

func (r *PostgresBrandRepository) buildWhereClause(filter domain.BrandFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *PostgresBrandRepository) buildOrderClause(sortBy string, sortOrder domain.SortOrder) string {
	var column string

	switch sortBy {
	case "name":
		column = "name"
	case "created_at":
		column = "created_at"
	case "updated_at":
		column = "updated_at"
	default:
		column = "name" // default sort by name
	}

	direction := "ASC"
	if sortOrder == domain.SortOrderDesc {
		direction = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", column, direction)
}
