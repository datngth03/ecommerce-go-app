package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/models"
)

// InventoryRepository defines methods for inventory data access
type InventoryRepository interface {
	// Stock operations
	GetStock(ctx context.Context, productID string) (*models.Stock, error)
	UpdateStock(ctx context.Context, productID string, quantity int32, reason string) (*models.Stock, error)
	CheckAvailability(ctx context.Context, productID string, quantity int32) (bool, error)

	// Reservation operations
	CreateReservation(ctx context.Context, orderID, productID string, quantity int32) (*models.Reservation, error)
	GetReservation(ctx context.Context, orderID string) ([]*models.Reservation, error)
	CommitReservation(ctx context.Context, orderID string) error
	ReleaseReservation(ctx context.Context, orderID string, reason string) error

	// Stock movement operations
	CreateMovement(ctx context.Context, movement *models.StockMovement) error
	GetMovementHistory(ctx context.Context, productID string, limit, offset int) ([]*models.StockMovement, int, error)
}
