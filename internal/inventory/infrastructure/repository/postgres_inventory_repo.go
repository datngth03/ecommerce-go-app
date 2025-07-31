// internal/inventory/infrastructure/repository/postgres_inventory_repo.go
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/inventory/domain"
)

// PostgreSQLInventoryRepository implements the domain.InventoryRepository interface
// for PostgreSQL database operations.
type PostgreSQLInventoryRepository struct {
	db *sql.DB
}

// NewPostgreSQLInventoryRepository creates a new instance of PostgreSQLInventoryRepository.
func NewPostgreSQLInventoryRepository(db *sql.DB) *PostgreSQLInventoryRepository {
	return &PostgreSQLInventoryRepository{db: db}
}

// Save stores an inventory item in PostgreSQL.
// If the item already exists (based on product_id), it updates it. Otherwise, it inserts a new one.
func (r *PostgreSQLInventoryRepository) Save(ctx context.Context, item *domain.InventoryItem) error {
	// Use UPSERT (INSERT ... ON CONFLICT UPDATE) for atomic operations
	query := `
        INSERT INTO inventory_items (product_id, quantity, reserved_quantity, last_updated_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (product_id) DO UPDATE
        SET quantity = EXCLUDED.quantity,
            reserved_quantity = EXCLUDED.reserved_quantity,
            last_updated_at = EXCLUDED.last_updated_at`

	_, err := r.db.ExecContext(ctx, query,
		item.ProductID, item.Quantity, item.ReservedQuantity, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save inventory item: %w", err)
	}
	return nil
}

// FindByProductID retrieves an inventory item by its product ID from PostgreSQL.
func (r *PostgreSQLInventoryRepository) FindByProductID(ctx context.Context, productID string) (*domain.InventoryItem, error) {
	item := &domain.InventoryItem{}
	query := `SELECT product_id, quantity, reserved_quantity, last_updated_at FROM inventory_items WHERE product_id = $1`
	row := r.db.QueryRowContext(ctx, query, productID)

	err := row.Scan(&item.ProductID, &item.Quantity, &item.ReservedQuantity, &item.LastUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: %v", domain.ErrInventoryNotFound, err)
		}
		return nil, fmt.Errorf("failed to find inventory item by product ID: %w", err)
	}
	return item, nil
}
