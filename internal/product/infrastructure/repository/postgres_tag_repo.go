package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

// PostgresTagRepository implements the domain.TagRepository interface.
type PostgresTagRepository struct {
	db *sql.DB
}

// NewPostgresTagRepository creates a new PostgresTagRepository.
func NewPostgresTagRepository(db *sql.DB) *PostgresTagRepository {
	return &PostgresTagRepository{db: db}
}

// Create inserts a new tag into the database.
func (r *PostgresTagRepository) Create(tag *domain.Tag) error {
	query := `INSERT INTO tags (name, slug) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRow(query, tag.Name, tag.Slug).Scan(&tag.ID)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// GetByID retrieves a tag by its ID.
func (r *PostgresTagRepository) GetByID(id string) (*domain.Tag, error) {
	return r.getTag("id", id)
}

// GetBySlug retrieves a tag by its slug.
func (r *PostgresTagRepository) GetBySlug(slug string) (*domain.Tag, error) {
	return r.getTag("slug", slug)
}

// getTag is a helper function to fetch a single tag by column and value.
func (r *PostgresTagRepository) getTag(column, value string) (*domain.Tag, error) {
	query := fmt.Sprintf(`SELECT id, name, slug FROM tags WHERE %s = $1`, column)
	tag := &domain.Tag{}
	err := r.db.QueryRow(query, value).Scan(&tag.ID, &tag.Name, &tag.Slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTagNotFound
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

// Update updates an existing tag in the database.
func (r *PostgresTagRepository) Update(tag *domain.Tag) error {
	query := `UPDATE tags SET name = $2, slug = $3 WHERE id = $1`
	result, err := r.db.Exec(query, tag.ID, tag.Name, tag.Slug)
	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrTagNotFound
	}

	return nil
}

// Delete deletes a tag from the database.
func (r *PostgresTagRepository) Delete(id string) error {
	tx, err := r.db.Begin()

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	productTagsQuery := `DELETE FROM product_tags WHERE tag_id = $1`
	if _, err := tx.Exec(productTagsQuery, id); err != nil {
		return fmt.Errorf("failed to delete related product tags: %w", err)
	}

	tagQuery := `DELETE FROM tags WHERE id = $1`
	result, err := tx.Exec(tagQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrTagNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List retrieves a list of tags based on the provided filter.
func (r *PostgresTagRepository) List(filter domain.TagFilter) ([]*domain.Tag, int, error) {
	// Build the base query and arguments
	var args []interface{}
	query := "SELECT id, name, slug FROM tags"
	countQuery := "SELECT count(*) FROM tags"

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

	// Get total count
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to get tag count: %w", err)
	}

	// Add sorting
	if filter.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, filter.SortOrder.String())
	}

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
		args = append(args, filter.Limit, filter.Offset)
	}

	// Execute the query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tag row: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return tags, total, nil
}
