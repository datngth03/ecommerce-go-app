// internal/inventory/domain/inventory.go
package domain

import (
	"errors"
	"time"
)

// InventoryItem represents the stock quantity of a product.
type InventoryItem struct {
	ProductID        string    `json:"product_id"`
	Quantity         int32     `json:"quantity"`          // Available quantity
	ReservedQuantity int32     `json:"reserved_quantity"` // Quantity reserved for pending operations
	LastUpdatedAt    time.Time `json:"last_updated_at"`
}

// NewInventoryItem creates a new InventoryItem instance.
func NewInventoryItem(productID string, initialQuantity int32) *InventoryItem {
	now := time.Now()
	return &InventoryItem{
		ProductID:        productID,
		Quantity:         initialQuantity,
		ReservedQuantity: 0,
		LastUpdatedAt:    now,
	}
}

// IncreaseQuantity increases the available stock.
func (i *InventoryItem) IncreaseQuantity(amount int32) {
	i.Quantity += amount
	i.LastUpdatedAt = time.Now()
}

// DecreaseQuantity decreases the available stock.
func (i *InventoryItem) DecreaseQuantity(amount int32) error {
	if i.Quantity < amount {
		return errors.New("not enough stock available")
	}
	i.Quantity -= amount
	i.LastUpdatedAt = time.Now()
	return nil
}

// ReserveQuantity reserves a certain amount of stock.
func (i *InventoryItem) ReserveQuantity(amount int32) error {
	if i.Quantity < amount {
		return errors.New("not enough stock to reserve")
	}
	i.Quantity -= amount
	i.ReservedQuantity += amount
	i.LastUpdatedAt = time.Now()
	return nil
}

// ReleaseReservedQuantity releases a certain amount of reserved stock back to available.
func (i *InventoryItem) ReleaseReservedQuantity(amount int32) error {
	if i.ReservedQuantity < amount {
		return errors.New("not enough reserved stock to release")
	}
	i.ReservedQuantity -= amount
	i.Quantity += amount
	i.LastUpdatedAt = time.Now()
	return nil
}

// SetQuantity sets the total available quantity directly.
func (i *InventoryItem) SetQuantity(newQuantity int32) {
	i.Quantity = newQuantity
	i.LastUpdatedAt = time.Now()
}
