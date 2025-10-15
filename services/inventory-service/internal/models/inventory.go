package models

import (
	"time"
)

// Stock represents product inventory
type Stock struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProductID   string    `json:"product_id" gorm:"uniqueIndex;not null"`
	Available   int32     `json:"available" gorm:"not null;default:0"` // Available for sale
	Reserved    int32     `json:"reserved" gorm:"not null;default:0"`  // Reserved for pending orders
	Total       int32     `json:"total" gorm:"not null;default:0"`     // Total physical stock
	WarehouseID string    `json:"warehouse_id" gorm:"default:'default'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Stock
func (Stock) TableName() string {
	return "stocks"
}

// MovementType constants
const (
	MovementTypeInbound    = "INBOUND"    // Stock added (purchase, return)
	MovementTypeOutbound   = "OUTBOUND"   // Stock removed (sale, damage)
	MovementTypeReserved   = "RESERVED"   // Stock reserved for order
	MovementTypeReleased   = "RELEASED"   // Reserved stock released
	MovementTypeCommitted  = "COMMITTED"  // Reserved stock committed (sold)
	MovementTypeAdjustment = "ADJUSTMENT" // Manual adjustment
)

// ReferenceType constants
const (
	ReferenceTypeOrder      = "ORDER"
	ReferenceTypePurchase   = "PURCHASE"
	ReferenceTypeAdjustment = "ADJUSTMENT"
	ReferenceTypeReturn     = "RETURN"
)

// StockMovement represents a stock transaction history
type StockMovement struct {
	ID             string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProductID      string    `json:"product_id" gorm:"index;not null"`
	MovementType   string    `json:"movement_type" gorm:"not null"` // INBOUND, OUTBOUND, RESERVED, etc.
	Quantity       int32     `json:"quantity" gorm:"not null"`
	BeforeQuantity int32     `json:"before_quantity" gorm:"not null"`
	AfterQuantity  int32     `json:"after_quantity" gorm:"not null"`
	ReferenceType  string    `json:"reference_type"` // ORDER, PURCHASE, ADJUSTMENT
	ReferenceID    string    `json:"reference_id" gorm:"index"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime;index"`
}

// TableName specifies the table name for StockMovement
func (StockMovement) TableName() string {
	return "stock_movements"
}

// Reservation represents a stock reservation for pending orders
type Reservation struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID     string    `json:"order_id" gorm:"uniqueIndex;not null"`
	ProductID   string    `json:"product_id" gorm:"index;not null"`
	Quantity    int32     `json:"quantity" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null;default:'PENDING'"` // PENDING, COMMITTED, RELEASED
	WarehouseID string    `json:"warehouse_id" gorm:"default:'default'"`
	ExpiresAt   time.Time `json:"expires_at"` // Auto-release if not committed
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for Reservation
func (Reservation) TableName() string {
	return "reservations"
}

// ReservationStatus constants
const (
	ReservationStatusPending   = "PENDING"
	ReservationStatusCommitted = "COMMITTED"
	ReservationStatusReleased  = "RELEASED"
	ReservationStatusExpired   = "EXPIRED"
)
