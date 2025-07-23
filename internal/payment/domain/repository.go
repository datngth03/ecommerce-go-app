// internal/payment/domain/repository.go

package domain

import (
	"context"
)

// PaymentRepository defines the interface for payment data operations.
type PaymentRepository interface {
	// Save creates a new payment record or updates an existing one.
	Save(ctx context.Context, payment *Payment) error

	// FindByID retrieves a payment by its ID.
	FindByID(ctx context.Context, id string) (*Payment, error)

	// FindByOrderID retrieves payments associated with a specific order ID.
	FindByOrderID(ctx context.Context, orderID string) ([]*Payment, error)

	// FindByTransactionID retrieves a payment by its transaction ID from the gateway.
	FindByTransactionID(ctx context.Context, transactionID string) (*Payment, error)

	// FindAll retrieves a list of payments based on filters and pagination.
	FindAll(ctx context.Context, userID, orderID, status string, limit, offset int32) ([]*Payment, int32, error)

	// Delete removes a payment record from the repository.
	Delete(ctx context.Context, id string) error
}
