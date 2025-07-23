// internal/shipping/domain/shipment.go
package domain

import (
	"time"
)

// Shipment represents the core shipment entity in the domain.
type Shipment struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"order_id"`
	UserID          string    `json:"user_id"`
	ShippingCost    float64   `json:"shipping_cost"`
	TrackingNumber  string    `json:"tracking_number,omitempty"`
	Carrier         string    `json:"carrier"` // e.g., "FedEx", "UPS", "LocalPost"
	Status          string    `json:"status"`  // e.g., "pending", "in_transit", "delivered", "failed"
	ShippingAddress string    `json:"shipping_address"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// NewShipment creates a new Shipment instance.
func NewShipment(id, orderID, userID string, shippingCost float64, shippingAddress, carrier string) *Shipment {
	now := time.Now()
	return &Shipment{
		ID:              id,
		OrderID:         orderID,
		UserID:          userID,
		ShippingCost:    shippingCost,
		Carrier:         carrier,
		Status:          "pending", // Default status
		ShippingAddress: shippingAddress,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// UpdateStatus updates the shipment's status.
func (s *Shipment) UpdateStatus(newStatus string) {
	s.Status = newStatus
	s.UpdatedAt = time.Now()
}

// SetTrackingNumber sets the tracking number for the shipment.
func (s *Shipment) SetTrackingNumber(trackingNumber string) {
	s.TrackingNumber = trackingNumber
	s.UpdatedAt = time.Now()
}
