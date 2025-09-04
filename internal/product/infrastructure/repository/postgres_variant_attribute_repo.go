package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// PostgresVariantAttributeRepository implements the domain.VariantAttributeRepository interface.
type PostgresVariantAttributeRepository struct {
	db *sql.DB
}

// NewPostgresVariantAttributeRepository creates a new PostgresVariantAttributeRepository.
func NewPostgresVariantAttributeRepository(db *sql.DB) *PostgresVariantAttributeRepository {
	return &PostgresVariantAttributeRepository{db: db}
}

// Create inserts a new variant attribute into the database.
func (r *PostgresVariantAttributeRepository) Create(attr *domain.VariantAttribute) error {
	query := `INSERT INTO variant_attributes (name, slug) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, attr.Name, attr.Slug).Scan(&attr.ID)
	if err != nil {
		return fmt.Errorf("failed to create variant attribute: %w", err)
	}
	return nil
}

// GetByID retrieves a variant attribute by its ID.
func (r *PostgresVariantAttributeRepository) GetByID(id string) (*domain.VariantAttribute, error) {
	return r.getAttribute("id", id)
}

// GetBySlug retrieves a variant attribute by its slug.
func (r *PostgresVariantAttributeRepository) GetBySlug(slug string) (*domain.VariantAttribute, error) {
	return r.getAttribute("slug", slug)
}

// getAttribute is a helper function to fetch a single attribute by a given column.
func (r *PostgresVariantAttributeRepository) getAttribute(column, value string) (*domain.VariantAttribute, error) {
	query := fmt.Sprintf(`SELECT id, name, slug FROM variant_attributes WHERE %s = $1`, column)
	attr := &domain.VariantAttribute{}
	err := r.db.QueryRow(query, value).Scan(&attr.ID, &attr.Name, &attr.Slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrVariantAttributeNotFound
		}
		return nil, fmt.Errorf("failed to get variant attribute: %w", err)
	}
	return attr, nil
}

// Update updates an existing variant attribute in the database.
func (r *PostgresVariantAttributeRepository) Update(attr *domain.VariantAttribute) error {
	query := `UPDATE variant_attributes SET name = $2, slug = $3 WHERE id = $1`
	result, err := r.db.Exec(query, attr.ID, attr.Name, attr.Slug)
	if err != nil {
		return fmt.Errorf("failed to update variant attribute: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrVariantAttributeNotFound
	}
	return nil
}

// Delete deletes a variant attribute and its associated values in a transaction.
func (r *PostgresVariantAttributeRepository) Delete(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete related variant attribute values first
	valueQuery := `DELETE FROM variant_attribute_values WHERE attribute_id = $1`
	if _, err := tx.Exec(valueQuery, id); err != nil {
		return fmt.Errorf("failed to delete related variant attribute values: %w", err)
	}

	// Delete the variant attribute itself
	attrQuery := `DELETE FROM variant_attributes WHERE id = $1`
	result, err := tx.Exec(attrQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete variant attribute: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrVariantAttributeNotFound
	}

	return tx.Commit()
}

// List retrieves a list of variant attributes based on the provided filter.
func (r *PostgresVariantAttributeRepository) List(filter domain.VariantAttributeFilter) ([]*domain.VariantAttribute, int, error) {
	var args []interface{}
	query := "SELECT id, name, slug FROM variant_attributes"
	countQuery := "SELECT count(*) FROM variant_attributes"

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
		return nil, 0, fmt.Errorf("failed to list variant attributes: %w", err)
	}
	defer rows.Close()

	var attributes []*domain.VariantAttribute
	for rows.Next() {
		attr := &domain.VariantAttribute{}
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
