// internal/cart/domain/repository.go
package domain

import (
	"context"
)

// CartRepository defines the interface for shopping cart data operations.
type CartRepository interface {
	// SaveCart saves or updates a user's shopping cart.
	// It will typically store the entire Cart object.
	SaveCart(ctx context.Context, cart *Cart) error

	// GetCart retrieves a user's shopping cart by user ID.
	// Returns nil and no error if cart is not found.
	GetCart(ctx context.Context, userID string) (*Cart, error)

	// DeleteCart removes a user's shopping cart.
	DeleteCart(ctx context.Context, userID string) error
}
