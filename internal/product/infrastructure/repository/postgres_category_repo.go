package repository

import (
	// "context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
)

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) Create(category *domain.Category) error {
	query := `
		INSERT INTO categories (id, name, slug, description, image, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	_, err := r.db.Exec(query, category.ID, category.Name, category.Slug, category.Description, category.Image, category.ParentID, category.CreatedAt, category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

func (r *PostgresCategoryRepository) GetByID(id string) (*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, image, parent_id, created_at, updated_at
		FROM categories
		WHERE id = $1`

	category := &domain.Category{}
	err := r.db.QueryRow(query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.Image,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetBySlug(slug string) (*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, image, parent_id, created_at, updated_at
		FROM categories
		WHERE slug = $1`

	category := &domain.Category{}
	err := r.db.QueryRow(query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.Image,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return category, nil
}

func (r *PostgresCategoryRepository) Update(category *domain.Category) error {
	query := `
		UPDATE categories 
		SET name = $2, slug = $3, description = $4, image = $5, parent_id = $6, updated_at = $7
		WHERE id = $1`

	category.UpdatedAt = time.Now()

	result, err := r.db.Exec(query, category.ID, category.Name, category.Slug, category.Description, category.Image, category.ParentID, category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

func (r *PostgresCategoryRepository) Delete(id string) error {
	// Check if category has children
	hasChildren, err := r.hasChildren(id)
	if err != nil {
		return fmt.Errorf("failed to check if category has children: %w", err)
	}

	if hasChildren {
		return domain.ErrCategoryHasChildren
	}

	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

func (r *PostgresCategoryRepository) List(filter domain.CategoryFilter) ([]*domain.Category, int, error) {
	var categories []*domain.Category
	var totalCount int

	// Build where clause
	whereClause, args := r.buildWhereClause(filter)

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM categories %s`, whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	// Main query with pagination
	orderClause := r.buildOrderClause(filter.SortBy, filter.SortOrder)
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, image, parent_id, created_at, updated_at
		FROM categories %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(args)+1, len(args)+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		category := &domain.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.Image,
			&category.ParentID,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating category rows: %w", err)
	}

	return categories, totalCount, nil
}

func (r *PostgresCategoryRepository) GetChildren(parentID string) ([]*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, image, parent_id, created_at, updated_at
		FROM categories
		WHERE parent_id = $1
		ORDER BY name ASC`

	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children categories: %w", err)
	}
	defer rows.Close()

	var children []*domain.Category
	for rows.Next() {
		category := &domain.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.Image,
			&category.ParentID,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child category: %w", err)
		}

		children = append(children, category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating children category rows: %w", err)
	}

	return children, nil
}

func (r *PostgresCategoryRepository) GetParent(categoryID string) (*domain.Category, error) {
	query := `
		SELECT p.id, p.name, p.slug, p.description, p.image, p.parent_id, p.created_at, p.updated_at
		FROM categories c
		INNER JOIN categories p ON c.parent_id = p.id
		WHERE c.id = $1`

	parent := &domain.Category{}
	err := r.db.QueryRow(query, categoryID).Scan(
		&parent.ID,
		&parent.Name,
		&parent.Slug,
		&parent.Description,
		&parent.Image,
		&parent.ParentID,
		&parent.CreatedAt,
		&parent.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Category has no parent (is root)
		}
		return nil, fmt.Errorf("failed to get parent category: %w", err)
	}

	return parent, nil
}

func (r *PostgresCategoryRepository) hasChildren(categoryID string) (bool, error) {
	query := `SELECT COUNT(*) FROM categories WHERE parent_id = $1`

	var count int
	err := r.db.QueryRow(query, categoryID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *PostgresCategoryRepository) buildWhereClause(filter domain.CategoryFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.ParentID != nil {
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, *filter.ParentID)
		argIndex++
	}

	if filter.OnlyRoot {
		conditions = append(conditions, "parent_id IS NULL")
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *PostgresCategoryRepository) buildOrderClause(sortBy string, sortOrder domain.SortOrder) string {
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
