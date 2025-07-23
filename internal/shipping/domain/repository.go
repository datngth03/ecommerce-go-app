// internal/shipping/domain/repository.go
package domain

import (
	"context"
)

// ShipmentRepository defines the interface for shipment data operations.
type ShipmentRepository interface {
	// Save creates a new shipment record or updates an existing one.
	Save(ctx context.Context, shipment *Shipment) error

	// FindByID retrieves a shipment by its ID.
	FindByID(ctx context.Context, id string) (*Shipment, error)

	// FindByOrderID retrieves shipments associated with a specific order ID.
	FindByOrderID(ctx context.Context, orderID string) ([]*Shipment, error)

	// FindByTrackingNumber retrieves a shipment by its tracking number.
	FindByTrackingNumber(ctx context.Context, trackingNumber string) (*Shipment, error)

	// FindAll retrieves a list of shipments based on filters and pagination.
	FindAll(ctx context.Context, userID, orderID, status string, limit, offset int32) ([]*Shipment, int32, error)

	// Delete removes a shipment record from the repository.
	Delete(ctx context.Context, id string) error
}
