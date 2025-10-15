package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment statuses
const (
	PaymentStatusPending    = "PENDING"
	PaymentStatusProcessing = "PROCESSING"
	PaymentStatusCompleted  = "COMPLETED"
	PaymentStatusFailed     = "FAILED"
	PaymentStatusRefunded   = "REFUNDED"
	PaymentStatusCancelled  = "CANCELLED"
)

// Payment methods
const (
	PaymentMethodStripe       = "STRIPE"
	PaymentMethodPayPal       = "PAYPAL"
	PaymentMethodCreditCard   = "CREDIT_CARD"
	PaymentMethodBankTransfer = "BANK_TRANSFER"
)

// Transaction types
const (
	TransactionTypeCharge        = "CHARGE"
	TransactionTypeRefund        = "REFUND"
	TransactionTypeAuthorization = "AUTHORIZATION"
	TransactionTypeCapture       = "CAPTURE"
)

// Refund statuses
const (
	RefundStatusPending    = "PENDING"
	RefundStatusProcessing = "PROCESSING"
	RefundStatusCompleted  = "COMPLETED"
	RefundStatusFailed     = "FAILED"
)

// Payment represents a payment transaction
type Payment struct {
	ID                string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrderID           string         `gorm:"type:varchar(255);not null;index" json:"order_id"`
	UserID            string         `gorm:"type:varchar(255);not null;index" json:"user_id"`
	Amount            float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency          string         `gorm:"type:varchar(3);not null;default:'USD'" json:"currency"`
	Status            string         `gorm:"type:varchar(50);not null;index" json:"status"`
	Method            string         `gorm:"type:varchar(50);not null" json:"method"`
	GatewayPaymentID  string         `gorm:"type:varchar(255);index" json:"gateway_payment_id"`
	GatewayCustomerID string         `gorm:"type:varchar(255)" json:"gateway_customer_id"`
	FailureReason     string         `gorm:"type:text" json:"failure_reason,omitempty"`
	Metadata          string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt         time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Transactions []Transaction `gorm:"foreignKey:PaymentID" json:"transactions,omitempty"`
	Refunds      []Refund      `gorm:"foreignKey:PaymentID" json:"refunds,omitempty"`
}

// Transaction represents a payment transaction log
type Transaction struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentID       string         `gorm:"type:uuid;not null;index" json:"payment_id"`
	TransactionType string         `gorm:"type:varchar(50);not null" json:"transaction_type"`
	Amount          float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status          string         `gorm:"type:varchar(50);not null" json:"status"`
	GatewayResponse string         `gorm:"type:jsonb" json:"gateway_response"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Payment Payment `gorm:"foreignKey:PaymentID" json:"-"`
}

// Refund represents a payment refund
type Refund struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentID       string         `gorm:"type:uuid;not null;index" json:"payment_id"`
	Amount          float64        `gorm:"type:decimal(10,2);not null" json:"amount"`
	Reason          string         `gorm:"type:text" json:"reason"`
	Status          string         `gorm:"type:varchar(50);not null;index" json:"status"`
	GatewayRefundID string         `gorm:"type:varchar(255);index" json:"gateway_refund_id"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Payment Payment `gorm:"foreignKey:PaymentID" json:"-"`
}

// PaymentMethod represents a saved payment method
type PaymentMethod struct {
	ID              string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          string         `gorm:"type:varchar(255);not null;index" json:"user_id"`
	MethodType      string         `gorm:"type:varchar(50);not null" json:"method_type"`
	Last4           string         `gorm:"type:varchar(4)" json:"last4"`
	Brand           string         `gorm:"type:varchar(50)" json:"brand"`
	GatewayMethodID string         `gorm:"type:varchar(255);not null" json:"gateway_method_id"`
	IsDefault       bool           `gorm:"default:false" json:"is_default"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// TableName specifies the table name for Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// TableName specifies the table name for Refund
func (Refund) TableName() string {
	return "refunds"
}

// TableName specifies the table name for PaymentMethod
func (PaymentMethod) TableName() string {
	return "payment_methods"
}
