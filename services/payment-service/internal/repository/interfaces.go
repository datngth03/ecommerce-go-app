package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/models"
)

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	// Payment operations
	CreatePayment(ctx context.Context, payment *models.Payment) error
	GetPayment(ctx context.Context, paymentID string) (*models.Payment, error)
	GetPaymentByOrder(ctx context.Context, orderID string) (*models.Payment, error)
	UpdatePayment(ctx context.Context, payment *models.Payment) error
	GetPaymentHistory(ctx context.Context, userID string, limit, offset int) ([]*models.Payment, int, error)

	// Transaction operations
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error
	GetTransactionsByPayment(ctx context.Context, paymentID string) ([]*models.Transaction, error)

	// Refund operations
	CreateRefund(ctx context.Context, refund *models.Refund) error
	GetRefund(ctx context.Context, refundID string) (*models.Refund, error)
	GetRefundsByPayment(ctx context.Context, paymentID string) ([]*models.Refund, error)
	UpdateRefund(ctx context.Context, refund *models.Refund) error

	// Payment method operations
	SavePaymentMethod(ctx context.Context, method *models.PaymentMethod) error
	GetPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethod, error)
	GetPaymentMethod(ctx context.Context, methodID string) (*models.PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, methodID string) error
}
