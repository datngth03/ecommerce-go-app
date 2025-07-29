// internal/inventory/domain/repository.go
package domain

import (
	"context"
	"errors"
)

var ErrInventoryNotFound = errors.New("inventory item not found")

// InventoryRepository defines the interface for inventory data operations.
type InventoryRepository interface {
	// Save stores an inventory item.
	Save(ctx context.Context, item *InventoryItem) error

	// FindByProductID retrieves an inventory item by its product ID.
	FindByProductID(ctx context.Context, productID string) (*InventoryItem, error)
}
