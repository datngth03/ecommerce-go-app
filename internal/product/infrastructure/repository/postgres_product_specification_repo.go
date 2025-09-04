package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// PostgresSpecificationAttributeRepository implements the domain.SpecificationAttributeRepository interface.
type PostgresSpecificationAttributeRepository struct {
	db *sql.DB
}

// NewPostgresSpecificationAttributeRepository creates and returns a new repository instance.
func NewPostgresSpecificationAttributeRepository(db *sql.DB) *PostgresSpecificationAttributeRepository {
	return &PostgresSpecificationAttributeRepository{db: db}
}

func (r *PostgresSpecificationAttributeRepository) Create(attr *domain.SpecificationAttribute) error {
	query := `INSERT INTO specification_attributes (name, slug) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, attr.Name, attr.Slug).Scan(&attr.ID)
	if err != nil {
		return fmt.Errorf("failed to create specification attribute: %w", err)
	}
	return nil
}

func (r *PostgresSpecificationAttributeRepository) GetByID(id string) (*domain.SpecificationAttribute, error) {
	return r.getAttribute("id", id)
}

func (r *PostgresSpecificationAttributeRepository) GetBySlug(slug string) (*domain.SpecificationAttribute, error) {
	return r.getAttribute("slug", slug)
}

// getAttribute is a helper function to fetch a single attribute by a given column.
func (r *PostgresSpecificationAttributeRepository) getAttribute(column, value string) (*domain.SpecificationAttribute, error) {
	query := fmt.Sprintf(`SELECT id, name, slug FROM specification_attributes WHERE %s = $1`, column)
	attr := &domain.SpecificationAttribute{}
	err := r.db.QueryRow(query, value).Scan(&attr.ID, &attr.Name, &attr.Slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSpecificationAttributeNotFound
		}
		return nil, fmt.Errorf("failed to get specification attribute: %w", err)
	}
	return attr, nil
}

func (r *PostgresSpecificationAttributeRepository) Update(attr *domain.SpecificationAttribute) error {
	query := `UPDATE specification_attributes SET name = $2, slug = $3 WHERE id = $1`
	result, err := r.db.Exec(query, attr.ID, attr.Name, attr.Slug)
	if err != nil {
		return fmt.Errorf("failed to update specification attribute: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrSpecificationAttributeNotFound
	}
	return nil
}

func (r *PostgresSpecificationAttributeRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	// Delete related ProductSpecification records first
	specQuery := `DELETE FROM product_specifications WHERE attribute_id = $1`
	if _, err := tx.Exec(specQuery, id); err != nil {
		return fmt.Errorf("failed to delete related product specifications: %w", err)
	}

	// Delete the SpecificationAttribute itself
	attrQuery := `DELETE FROM specification_attributes WHERE id = $1`
	result, err := tx.Exec(attrQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete specification attribute: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrSpecificationAttributeNotFound
	}

	return tx.Commit()
}

func (r *PostgresSpecificationAttributeRepository) List(filter domain.SpecificationAttributeFilter) ([]*domain.SpecificationAttribute, int, error) {
	var args []interface{}
	query := "SELECT id, name, slug FROM specification_attributes"
	countQuery := "SELECT count(*) FROM specification_attributes"

	var conditions []string
	argCounter := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR slug ILIKE $%d)", argCounter, argCounter))
		args = append(args, "%"+filter.Search+"%")
		argCounter++
	}

	if len(conditions) > 0 {
		whereClause := " WHERE " + strings.Join(conditions, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get attribute count: %w", err)
	}

	if filter.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, filter.SortOrder.String())
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list specification attributes: %w", err)
	}
	defer rows.Close()

	var attributes []*domain.SpecificationAttribute
	for rows.Next() {
		attr := &domain.SpecificationAttribute{}
		if err := rows.Scan(&attr.ID, &attr.Name, &attr.Slug); err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return attributes, total, nil
}
