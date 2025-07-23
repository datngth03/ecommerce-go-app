// internal/payment/domain/payment.go
package domain

import (
	"time"
)

// Payment represents the core payment entity in the domain.
type Payment struct {
	ID            string    `json:"id"`
	OrderID       string    `json:"order_id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`                   // e.g., "pending", "completed", "failed", "refunded"
	PaymentMethod string    `json:"payment_method"`           // e.g., "credit_card", "paypal", "bank_transfer"
	TransactionID string    `json:"transaction_id,omitempty"` // ID from the payment gateway
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewPayment creates a new Payment instance.
func NewPayment(id, orderID, userID string, amount float64, currency, paymentMethod string) *Payment {
	now := time.Now()
	return &Payment{
		ID:            id,
		OrderID:       orderID,
		UserID:        userID,
		Amount:        amount,
		Currency:      currency,
		Status:        "pending", // Default status
		PaymentMethod: paymentMethod,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// UpdateStatus updates the payment's status.
func (p *Payment) UpdateStatus(newStatus string) {
	p.Status = newStatus
	p.UpdatedAt = time.Now()
}

// SetTransactionID sets the transaction ID from the payment gateway.
func (p *Payment) SetTransactionID(transactionID string) {
	p.TransactionID = transactionID
	p.UpdatedAt = time.Now()
}
