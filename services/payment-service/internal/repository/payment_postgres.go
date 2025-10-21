package repository

import (
	"context"
	"fmt"

	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/models"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{
		db: db,
	}
}

// CreatePayment creates a new payment
func (r *paymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// GetPayment retrieves a payment by ID
func (r *paymentRepository) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Preload("Transactions").
		Preload("Refunds").
		Where("id = ?", paymentID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetPaymentByOrder retrieves a payment by order ID
func (r *paymentRepository) GetPaymentByOrder(ctx context.Context, orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.WithContext(ctx).
		Preload("Transactions").
		Preload("Refunds").
		Where("order_id = ?", orderID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdatePayment updates a payment
func (r *paymentRepository) UpdatePayment(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// GetPaymentHistory retrieves payment history for a user
func (r *paymentRepository) GetPaymentHistory(ctx context.Context, userID string, limit, offset int) ([]*models.Payment, int, error) {
	var payments []*models.Payment
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error
	if err != nil {
		return nil, 0, err
	}

	return payments, int(total), nil
}

// CreateTransaction creates a new transaction
func (r *paymentRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetTransactionsByPayment retrieves all transactions for a payment
func (r *paymentRepository) GetTransactionsByPayment(ctx context.Context, paymentID string) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// CreateRefund creates a new refund
func (r *paymentRepository) CreateRefund(ctx context.Context, refund *models.Refund) error {
	return r.db.WithContext(ctx).Create(refund).Error
}

// GetRefund retrieves a refund by ID
func (r *paymentRepository) GetRefund(ctx context.Context, refundID string) (*models.Refund, error) {
	var refund models.Refund
	err := r.db.WithContext(ctx).Where("id = ?", refundID).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// GetRefundsByPayment retrieves all refunds for a payment
func (r *paymentRepository) GetRefundsByPayment(ctx context.Context, paymentID string) ([]*models.Refund, error) {
	var refunds []*models.Refund
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at DESC").
		Find(&refunds).Error
	if err != nil {
		return nil, err
	}
	return refunds, nil
}

// UpdateRefund updates a refund
func (r *paymentRepository) UpdateRefund(ctx context.Context, refund *models.Refund) error {
	return r.db.WithContext(ctx).Save(refund).Error
}

// SavePaymentMethod saves a payment method
func (r *paymentRepository) SavePaymentMethod(ctx context.Context, method *models.PaymentMethod) error {
	// If this is set as default, unset other defaults
	if method.IsDefault {
		query := r.db.WithContext(ctx).Model(&models.PaymentMethod{}).
			Where("user_id = ?", method.UserID)

		// Only exclude current ID if it's not empty (updating existing method)
		if method.ID != "" {
			query = query.Where("id != ?", method.ID)
		}

		if err := query.Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset other defaults: %w", err)
		}
	}
	return r.db.WithContext(ctx).Create(method).Error
}

// GetPaymentMethods retrieves all payment methods for a user
func (r *paymentRepository) GetPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethod, error) {
	var methods []*models.PaymentMethod
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&methods).Error
	if err != nil {
		return nil, err
	}
	return methods, nil
}

// GetPaymentMethod retrieves a payment method by ID
func (r *paymentRepository) GetPaymentMethod(ctx context.Context, methodID string) (*models.PaymentMethod, error) {
	var method models.PaymentMethod
	err := r.db.WithContext(ctx).Where("id = ?", methodID).First(&method).Error
	if err != nil {
		return nil, err
	}
	return &method, nil
}

// DeletePaymentMethod deletes a payment method
func (r *paymentRepository) DeletePaymentMethod(ctx context.Context, methodID string) error {
	return r.db.WithContext(ctx).Delete(&models.PaymentMethod{}, "id = ?", methodID).Error
}
