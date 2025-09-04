package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// PostgresVariantAttributeValueRepository implements the domain.VariantAttributeValueRepository interface.
type PostgresVariantAttributeValueRepository struct {
	db *sql.DB
}

// NewPostgresVariantAttributeValueRepository creates a new PostgresVariantAttributeValueRepository.
func NewPostgresVariantAttributeValueRepository(db *sql.DB) *PostgresVariantAttributeValueRepository {
	return &PostgresVariantAttributeValueRepository{db: db}
}

// Create inserts a new variant attribute value into the database.
func (r *PostgresVariantAttributeValueRepository) Create(value *domain.VariantAttributeValue) error {
	query := `INSERT INTO variant_attribute_values (attribute_id, value) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, value.AttributeID, value.Value).Scan(&value.ID)
	if err != nil {
		return fmt.Errorf("failed to create variant attribute value: %w", err)
	}
	return nil
}

// GetByID retrieves a variant attribute value by its ID.
func (r *PostgresVariantAttributeValueRepository) GetByID(id string) (*domain.VariantAttributeValue, error) {
	value := &domain.VariantAttributeValue{}
	query := `SELECT id, attribute_id, value FROM variant_attribute_values WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&value.ID, &value.AttributeID, &value.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrVariantAttributeValueNotFound
		}
		return nil, fmt.Errorf("failed to get variant attribute value: %w", err)
	}
	return value, nil
}

// Update updates an existing variant attribute value.
func (r *PostgresVariantAttributeValueRepository) Update(value *domain.VariantAttributeValue) error {
	query := `UPDATE variant_attribute_values SET attribute_id = $2, value = $3 WHERE id = $1`
	result, err := r.db.Exec(query, value.ID, value.AttributeID, value.Value)
	if err != nil {
		return fmt.Errorf("failed to update variant attribute value: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrVariantAttributeValueNotFound
	}
	return nil
}

// Delete deletes a variant attribute value. Note: This assumes `product_variant_options` has ON DELETE CASCADE.
func (r *PostgresVariantAttributeValueRepository) Delete(id string) error {
	query := `DELETE FROM variant_attribute_values WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete variant attribute value: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrVariantAttributeValueNotFound
	}
	return nil
}

// List retrieves a list of variant attribute values based on the filter.
func (r *PostgresVariantAttributeValueRepository) List(filter domain.VariantAttributeValueFilter) ([]*domain.VariantAttributeValue, int, error) {
	var args []interface{}
	query := "SELECT id, attribute_id, value FROM variant_attribute_values"
	countQuery := "SELECT count(*) FROM variant_attribute_values"

	var conditions []string
	argCounter := 1

	if filter.AttributeID != "" {
		conditions = append(conditions, fmt.Sprintf("attribute_id = $%d", argCounter))
		args = append(args, filter.AttributeID)
		argCounter++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("value ILIKE $%d", argCounter))
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
		return nil, 0, fmt.Errorf("failed to get attribute value count: %w", err)
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
		return nil, 0, fmt.Errorf("failed to list variant attribute values: %w", err)
	}
	defer rows.Close()

	var values []*domain.VariantAttributeValue
	for rows.Next() {
		value := &domain.VariantAttributeValue{}
		if err := rows.Scan(&value.ID, &value.AttributeID, &value.Value); err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
		values = append(values, value)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return values, total, nil
}

// GetByAttributeID retrieves all variant attribute values for a given attribute.
func (r *PostgresVariantAttributeValueRepository) GetByAttributeID(attributeID string) ([]*domain.VariantAttributeValue, error) {
	query := `SELECT id, attribute_id, value FROM variant_attribute_values WHERE attribute_id = $1`
	rows, err := r.db.Query(query, attributeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get values by attribute ID: %w", err)
	}
	defer rows.Close()

	var values []*domain.VariantAttributeValue
	for rows.Next() {
		value := &domain.VariantAttributeValue{}
		if err := rows.Scan(&value.ID, &value.AttributeID, &value.Value); err != nil {
			return nil, fmt.Errorf("failed to scan value row: %w", err)
		}
		values = append(values, value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return values, nil
}
