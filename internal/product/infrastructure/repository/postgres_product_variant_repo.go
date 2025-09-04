package repository

import (
	// "context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
	"github.com/lib/pq"
)

type PostgresProductVariantRepository struct {
	db *sql.DB
}

func NewPostgresProductVariantRepository(db *sql.DB) *PostgresProductVariantRepository {
	return &PostgresProductVariantRepository{db: db}
}

func (r *PostgresProductVariantRepository) Create(variant *domain.ProductVariant) error {
	query := `
		INSERT INTO product_variants (id, product_id, sku, price, original_price, discount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	now := time.Now()
	variant.CreatedAt = now
	variant.UpdatedAt = now

	_, err := r.db.Exec(query, variant.ID, variant.ProductID, variant.SKU, variant.Price, variant.OriginalPrice, variant.Discount, variant.CreatedAt, variant.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create product variant: %w", err)
	}

	return nil
}

func (r *PostgresProductVariantRepository) GetByID(id string) (*domain.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, price, original_price, discount, created_at, updated_at
		FROM product_variants
		WHERE id = $1`

	variant := &domain.ProductVariant{}
	err := r.db.QueryRow(query, id).Scan(
		&variant.ID,
		&variant.ProductID,
		&variant.SKU,
		&variant.Price,
		&variant.OriginalPrice,
		&variant.Discount,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProductVariantNotFound
		}
		return nil, fmt.Errorf("failed to get product variant by ID: %w", err)
	}

	// Load attribute values
	err = r.loadAttributeValuesForVariant(variant)
	if err != nil {
		return nil, fmt.Errorf("failed to load attribute values: %w", err)
	}

	return variant, nil
}

func (r *PostgresProductVariantRepository) GetBySKU(sku string) (*domain.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, price, original_price, discount, created_at, updated_at
		FROM product_variants
		WHERE sku = $1`

	variant := &domain.ProductVariant{}
	err := r.db.QueryRow(query, sku).Scan(
		&variant.ID,
		&variant.ProductID,
		&variant.SKU,
		&variant.Price,
		&variant.OriginalPrice,
		&variant.Discount,
		&variant.CreatedAt,
		&variant.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrProductVariantNotFound
		}
		return nil, fmt.Errorf("failed to get product variant by SKU: %w", err)
	}

	// Load attribute values
	err = r.loadAttributeValuesForVariant(variant)
	if err != nil {
		return nil, fmt.Errorf("failed to load attribute values: %w", err)
	}

	return variant, nil
}

func (r *PostgresProductVariantRepository) Update(variant *domain.ProductVariant) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Cập nhật dữ liệu chính của ProductVariant
	variant.UpdatedAt = time.Now()
	updateQuery := `
        UPDATE product_variants 
        SET product_id = $2, sku = $3, price = $4, original_price = $5, discount = $6, updated_at = $7
        WHERE id = $1`

	result, err := tx.Exec(updateQuery, variant.ID, variant.ProductID, variant.SKU, variant.Price, variant.OriginalPrice, variant.Discount, variant.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update product variant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrProductVariantNotFound
	}

	// 2. Xóa tất cả các ProductVariantOption cũ của biến thể này
	deleteOptionsQuery := `DELETE FROM product_variant_options WHERE variant_id = $1`
	if _, err := tx.Exec(deleteOptionsQuery, variant.ID); err != nil {
		return fmt.Errorf("failed to delete old product variant options: %w", err)
	}

	// 3. Commit giao dịch
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresProductVariantRepository) Delete(id string) error {
	// 1. Bắt đầu một giao dịch
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Luôn rollback nếu có lỗi xảy ra sau đó
	defer tx.Rollback()

	// 2. Xóa các ProductVariantOption liên quan
	optionsQuery := `DELETE FROM product_variant_options WHERE variant_id = $1`
	if _, err := tx.Exec(optionsQuery, id); err != nil {
		return fmt.Errorf("failed to delete product variant options: %w", err)
	}

	// 3. Xóa ProductVariant chính
	variantQuery := `DELETE FROM product_variants WHERE id = $1`
	result, err := tx.Exec(variantQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete product variant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		// Dù đã xóa options, nếu không tìm thấy variant, vẫn trả về lỗi
		return domain.ErrProductVariantNotFound
	}

	// 4. Commit giao dịch nếu tất cả các bước thành công
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (r *PostgresProductVariantRepository) List(filter domain.ProductVariantFilter) ([]*domain.ProductVariant, int, error) {
	var variants []*domain.ProductVariant
	var totalCount int

	// Build where clause
	whereClause, args := r.buildWhereClause(filter)

	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM product_variants %s`, whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count product variants: %w", err)
	}

	// Main query with pagination
	orderClause := r.buildOrderClause(filter.SortBy, filter.SortOrder)
	query := fmt.Sprintf(`
		SELECT id, product_id, sku, price, original_price, discount, created_at, updated_at
		FROM product_variants %s %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(args)+1, len(args)+2)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list product variants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		variant := &domain.ProductVariant{}
		err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.SKU,
			&variant.Price,
			&variant.OriginalPrice,
			&variant.Discount,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product variant: %w", err)
		}

		variants = append(variants, variant)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating product variant rows: %w", err)
	}

	// Load attribute values for all variants
	err = r.loadAttributeValuesForVariants(variants)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to load attribute values: %w", err)
	}

	return variants, totalCount, nil
}

func (r *PostgresProductVariantRepository) GetByProductID(productID string) ([]*domain.ProductVariant, error) {
	query := `
		SELECT id, product_id, sku, price, original_price, discount, created_at, updated_at
		FROM product_variants
		WHERE product_id = $1
		ORDER BY sku ASC`

	rows, err := r.db.Query(query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variants by product ID: %w", err)
	}
	defer rows.Close()

	var variants []*domain.ProductVariant
	for rows.Next() {
		variant := &domain.ProductVariant{}
		err := rows.Scan(
			&variant.ID,
			&variant.ProductID,
			&variant.SKU,
			&variant.Price,
			&variant.OriginalPrice,
			&variant.Discount,
			&variant.CreatedAt,
			&variant.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}

		variants = append(variants, variant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating variant rows: %w", err)
	}

	// Load attribute values for all variants
	err = r.loadAttributeValuesForVariants(variants)
	if err != nil {
		return nil, fmt.Errorf("failed to load attribute values: %w", err)
	}

	return variants, nil
}

func (r *PostgresProductVariantRepository) loadAttributeValuesForVariant(variant *domain.ProductVariant) error {
	query := `
		SELECT vav.id, vav.attribute_id, vav.value, va.name, va.slug
		FROM product_variant_attributes pva
		INNER JOIN variant_attribute_values vav ON pva.attribute_value_id = vav.id
		INNER JOIN variant_attributes va ON vav.attribute_id = va.id
		WHERE pva.variant_id = $1
		ORDER BY va.name`

	rows, err := r.db.Query(query, variant.ID)
	if err != nil {
		return fmt.Errorf("failed to load attribute values for variant: %w", err)
	}
	defer rows.Close()

	var attributeValues []*domain.VariantAttributeValue
	for rows.Next() {
		var (
			valueID     string
			attributeID string
			value       string
			attrName    string
			attrSlug    string
		)

		err := rows.Scan(&valueID, &attributeID, &value, &attrName, &attrSlug)
		if err != nil {
			return fmt.Errorf("failed to scan attribute value: %w", err)
		}

		attributeValue := &domain.VariantAttributeValue{
			ID:          valueID,
			AttributeID: attributeID,
			Value:       value,
			Attribute: &domain.VariantAttribute{
				ID:   attributeID,
				Name: attrName,
				Slug: attrSlug,
			},
		}

		attributeValues = append(attributeValues, attributeValue)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating attribute value rows: %w", err)
	}

	variant.AttributeValues = attributeValues
	return nil
}

func (r *PostgresProductVariantRepository) loadAttributeValuesForVariants(variants []*domain.ProductVariant) error {
	if len(variants) == 0 {
		return nil
	}

	variantIDs := make([]string, len(variants))
	for i, v := range variants {
		variantIDs[i] = v.ID
	}

	query := `
		SELECT pva.variant_id, vav.id, vav.attribute_id, vav.value, va.name, va.slug
		FROM product_variant_attributes pva
		INNER JOIN variant_attribute_values vav ON pva.attribute_value_id = vav.id
		INNER JOIN variant_attributes va ON vav.attribute_id = va.id
		WHERE pva.variant_id = ANY($1)
		ORDER BY va.name`

	rows, err := r.db.Query(query, pq.Array(variantIDs))
	if err != nil {
		return fmt.Errorf("failed to load attribute values for variants: %w", err)
	}
	defer rows.Close()

	// Map để group attribute values theo variant_id
	variantAttributeValues := make(map[string][]*domain.VariantAttributeValue)

	for rows.Next() {
		var (
			variantID   string
			valueID     string
			attributeID string
			value       string
			attrName    string
			attrSlug    string
		)

		err := rows.Scan(&variantID, &valueID, &attributeID, &value, &attrName, &attrSlug)
		if err != nil {
			return fmt.Errorf("failed to scan attribute value: %w", err)
		}

		attributeValue := &domain.VariantAttributeValue{
			ID:          valueID,
			AttributeID: attributeID,
			Value:       value,
			Attribute: &domain.VariantAttribute{
				ID:   attributeID,
				Name: attrName,
				Slug: attrSlug,
			},
		}

		variantAttributeValues[variantID] = append(variantAttributeValues[variantID], attributeValue)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating attribute value rows: %w", err)
	}

	// Assign attribute values to variants
	for _, variant := range variants {
		if attributeValues, exists := variantAttributeValues[variant.ID]; exists {
			variant.AttributeValues = attributeValues
		}
	}

	return nil
}

func (r *PostgresProductVariantRepository) buildWhereClause(filter domain.ProductVariantFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.ProductID != "" {
		conditions = append(conditions, fmt.Sprintf("product_id = $%d", argIndex))
		args = append(args, filter.ProductID)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("sku ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func (r *PostgresProductVariantRepository) buildOrderClause(sortBy string, sortOrder domain.SortOrder) string {
	var column string

	switch sortBy {
	case "sku":
		column = "sku"
	case "price":
		column = "price"
	case "original_price":
		column = "original_price"
	case "created_at":
		column = "created_at"
	case "updated_at":
		column = "updated_at"
	default:
		column = "sku" // default sort by sku
	}

	direction := "ASC"
	if sortOrder == domain.SortOrderDesc {
		direction = "DESC"
	}

	return fmt.Sprintf("ORDER BY %s %s", column, direction)
}
