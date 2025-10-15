package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/repository"
)

// PaymentService handles payment business logic
type PaymentService struct {
	repo repository.PaymentRepository
}

// NewPaymentService creates a new payment service
func NewPaymentService(repo repository.PaymentRepository) *PaymentService {
	return &PaymentService{
		repo: repo,
	}
}

// ProcessPayment processes a new payment
func (s *PaymentService) ProcessPayment(ctx context.Context, orderID, userID string, amount float64, currency, method string, metadata map[string]string) (*models.Payment, string, error) {
	// Validate input
	if orderID == "" || userID == "" {
		return nil, "", fmt.Errorf("order_id and user_id are required")
	}
	if amount <= 0 {
		return nil, "", fmt.Errorf("amount must be positive")
	}

	// Convert metadata to JSON
	metadataJSON, _ := json.Marshal(metadata)

	// Create payment record
	payment := &models.Payment{
		OrderID:  orderID,
		UserID:   userID,
		Amount:   amount,
		Currency: currency,
		Status:   models.PaymentStatusPending,
		Method:   method,
		Metadata: string(metadataJSON),
	}

	// TODO: Integrate with payment gateway (Stripe/PayPal)
	// For now, we'll simulate payment processing
	payment.Status = models.PaymentStatusProcessing
	payment.GatewayPaymentID = fmt.Sprintf("sim_%s", orderID) // Simulated gateway ID

	err := s.repo.CreatePayment(ctx, payment)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create payment: %w", err)
	}

	// Create transaction log
	transaction := &models.Transaction{
		PaymentID:       payment.ID,
		TransactionType: models.TransactionTypeCharge,
		Amount:          amount,
		Status:          models.PaymentStatusProcessing,
		GatewayResponse: `{"simulated": true}`,
	}
	s.repo.CreateTransaction(ctx, transaction)

	// Simulate successful payment (in production, this would be async via webhook)
	payment.Status = models.PaymentStatusCompleted
	s.repo.UpdatePayment(ctx, payment)

	return payment, "", nil // client_secret for 3D Secure (not implemented)
}

// ConfirmPayment confirms a pending payment (for 3D Secure)
func (s *PaymentService) ConfirmPayment(ctx context.Context, paymentID, paymentIntentID string) (*models.Payment, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// TODO: Confirm with payment gateway
	payment.Status = models.PaymentStatusCompleted
	err = s.repo.UpdatePayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return payment, nil
}

// RefundPayment processes a refund
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string, amount float64, reason string) (*models.Refund, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	if payment.Status != models.PaymentStatusCompleted {
		return nil, fmt.Errorf("can only refund completed payments")
	}

	// Create refund record
	refund := &models.Refund{
		PaymentID: paymentID,
		Amount:    amount,
		Reason:    reason,
		Status:    models.RefundStatusPending,
	}

	// TODO: Process refund with payment gateway
	refund.Status = models.RefundStatusCompleted
	refund.GatewayRefundID = fmt.Sprintf("rfnd_%s", payment.ID)

	err = s.repo.CreateRefund(ctx, refund)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	// Update payment status if fully refunded
	if amount >= payment.Amount {
		payment.Status = models.PaymentStatusRefunded
		s.repo.UpdatePayment(ctx, payment)
	}

	return refund, nil
}

// GetPayment retrieves payment details
func (s *PaymentService) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	return s.repo.GetPayment(ctx, paymentID)
}

// GetPaymentByOrder retrieves payment by order ID
func (s *PaymentService) GetPaymentByOrder(ctx context.Context, orderID string) (*models.Payment, error) {
	return s.repo.GetPaymentByOrder(ctx, orderID)
}

// GetPaymentHistory retrieves user payment history
func (s *PaymentService) GetPaymentHistory(ctx context.Context, userID string, limit, offset int) ([]*models.Payment, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetPaymentHistory(ctx, userID, limit, offset)
}

// SavePaymentMethod saves a payment method
func (s *PaymentService) SavePaymentMethod(ctx context.Context, userID, methodType, gatewayMethodID string, isDefault bool) (*models.PaymentMethod, error) {
	// TODO: Validate with payment gateway and get card details
	method := &models.PaymentMethod{
		UserID:          userID,
		MethodType:      methodType,
		GatewayMethodID: gatewayMethodID,
		IsDefault:       isDefault,
		Last4:           "4242", // Would come from gateway
		Brand:           "VISA", // Would come from gateway
	}

	err := s.repo.SavePaymentMethod(ctx, method)
	if err != nil {
		return nil, fmt.Errorf("failed to save payment method: %w", err)
	}

	return method, nil
}

// GetPaymentMethods retrieves user's payment methods
func (s *PaymentService) GetPaymentMethods(ctx context.Context, userID string) ([]*models.PaymentMethod, error) {
	return s.repo.GetPaymentMethods(ctx, userID)
}

// HandleWebhook handles payment gateway webhooks
func (s *PaymentService) HandleWebhook(ctx context.Context, gateway, eventType, eventData string) error {
	// TODO: Implement webhook handling for Stripe/PayPal
	// Parse event, verify signature, update payment status
	return nil
}
