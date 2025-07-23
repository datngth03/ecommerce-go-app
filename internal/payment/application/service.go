// internal/payment/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/payment/domain"
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"     // For calling Order Service
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment" // Generated Payment gRPC client
	"github.com/google/uuid"                                                  // For generating UUIDs
)

// PaymentService defines the application service interface for payment-related operations.
type PaymentService interface {
	CreatePayment(ctx context.Context, req *payment_client.CreatePaymentRequest) (*payment_client.PaymentResponse, error)
	GetPaymentById(ctx context.Context, req *payment_client.GetPaymentByIdRequest) (*payment_client.PaymentResponse, error)
	ConfirmPayment(ctx context.Context, req *payment_client.ConfirmPaymentRequest) (*payment_client.PaymentResponse, error)
	RefundPayment(ctx context.Context, req *payment_client.RefundPaymentRequest) (*payment_client.PaymentResponse, error)
	ListPayments(ctx context.Context, req *payment_client.ListPaymentsRequest) (*payment_client.ListPaymentsResponse, error)
}

// paymentService implements the PaymentService interface.
type paymentService struct {
	paymentRepo domain.PaymentRepository
	orderClient order_client.OrderServiceClient // Client to call Order Service
	// TODO: Add other dependencies like payment gateway client, event publisher
}

// NewPaymentService creates a new instance of PaymentService.
func NewPaymentService(paymentRepo domain.PaymentRepository, orderClient order_client.OrderServiceClient) PaymentService {
	return &paymentService{
		paymentRepo: paymentRepo,
		orderClient: orderClient,
	}
}

// CreatePayment handles the initiation of a new payment for an order.
func (s *paymentService) CreatePayment(ctx context.Context, req *payment_client.CreatePaymentRequest) (*payment_client.PaymentResponse, error) {
	if req.GetOrderId() == "" || req.GetUserId() == "" || req.GetAmount() <= 0 || req.GetCurrency() == "" || req.GetPaymentMethod() == "" {
		return nil, errors.New("order ID, user ID, amount, currency, and payment method are required")
	}

	// Validate order existence and amount with Order Service
	orderResp, err := s.orderClient.GetOrderById(ctx, &order_client.GetOrderByIdRequest{Id: req.GetOrderId()})
	if err != nil {
		return nil, fmt.Errorf("failed to get order details for ID %s: %w", req.GetOrderId(), err)
	}
	if orderResp.GetOrder() == nil {
		return nil, fmt.Errorf("order with ID %s not found", req.GetOrderId())
	}
	if orderResp.GetOrder().GetTotalAmount() != req.GetAmount() {
		return nil, fmt.Errorf("amount mismatch for order %s. Expected %.2f, got %.2f",
			req.GetOrderId(), orderResp.GetOrder().GetTotalAmount(), req.GetAmount())
	}
	if orderResp.GetOrder().GetStatus() != "pending" {
		return nil, fmt.Errorf("order %s is not in pending status for payment", req.GetOrderId())
	}

	paymentID := uuid.New().String()
	payment := domain.NewPayment(
		paymentID,
		req.GetOrderId(),
		req.GetUserId(),
		req.GetAmount(),
		req.GetCurrency(),
		req.GetPaymentMethod(),
	)

	if err := s.paymentRepo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// TODO: Integrate with a real payment gateway here (e.g., Stripe, PayPal)
	// This would involve sending payment details to the gateway and getting a transaction ID.
	// For now, we'll just simulate a pending payment.

	return &payment_client.PaymentResponse{
		Payment: &payment_client.Payment{
			Id:            payment.ID,
			OrderId:       payment.OrderID,
			UserId:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Status:        payment.Status,
			PaymentMethod: payment.PaymentMethod,
			CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetPaymentById retrieves payment details by ID.
func (s *paymentService) GetPaymentById(ctx context.Context, req *payment_client.GetPaymentByIdRequest) (*payment_client.PaymentResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("payment ID is required")
	}

	payment, err := s.paymentRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, errors.New("payment not found")) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to retrieve payment: %w", err)
	}

	return &payment_client.PaymentResponse{
		Payment: &payment_client.Payment{
			Id:            payment.ID,
			OrderId:       payment.OrderID,
			UserId:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Status:        payment.Status,
			PaymentMethod: payment.PaymentMethod,
			TransactionId: payment.TransactionID,
			CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ConfirmPayment handles confirming a payment after a successful transaction with the gateway.
func (s *paymentService) ConfirmPayment(ctx context.Context, req *payment_client.ConfirmPaymentRequest) (*payment_client.PaymentResponse, error) {
	if req.GetPaymentId() == "" || req.GetTransactionId() == "" || req.GetStatus() == "" {
		return nil, errors.New("payment ID, transaction ID, and status are required")
	}

	payment, err := s.paymentRepo.FindByID(ctx, req.GetPaymentId())
	if err != nil {
		if errors.Is(err, errors.New("payment not found")) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment for confirmation: %w", err)
	}

	if payment.Status == "completed" || payment.Status == "failed" || payment.Status == "refunded" {
		return nil, errors.New("payment already finalized")
	}

	payment.SetTransactionID(req.GetTransactionId())
	payment.UpdateStatus(req.GetStatus()) // Update status based on gateway's final status

	if err := s.paymentRepo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save confirmed payment: %w", err)
	}

	// Update order status in Order Service if payment is successful
	if req.GetStatus() == "completed" {
		_, err := s.orderClient.UpdateOrderStatus(ctx, &order_client.UpdateOrderStatusRequest{
			OrderId:   payment.OrderID,
			NewStatus: "paid", // Or "completed"
		})
		if err != nil {
			log.Printf("Warning: Failed to update order status to 'paid' for order %s after payment %s: %v", payment.OrderID, payment.ID, err)
			// Decide how to handle this: retry, manual intervention, dead letter queue
		}
	}
	// TODO: Publish PaymentConfirmed event (for other services like Inventory, Shipping)

	return &payment_client.PaymentResponse{
		Payment: &payment_client.Payment{
			Id:            payment.ID,
			OrderId:       payment.OrderID,
			UserId:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Status:        payment.Status,
			PaymentMethod: payment.PaymentMethod,
			TransactionId: payment.TransactionID,
			CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// RefundPayment handles refunding a payment.
func (s *paymentService) RefundPayment(ctx context.Context, req *payment_client.RefundPaymentRequest) (*payment_client.PaymentResponse, error) {
	if req.GetPaymentId() == "" {
		return nil, errors.New("payment ID is required for refund")
	}

	payment, err := s.paymentRepo.FindByID(ctx, req.GetPaymentId())
	if err != nil {
		if errors.Is(err, errors.New("payment not found")) {
			return nil, errors.New("payment not found")
		}
		return nil, fmt.Errorf("failed to find payment for refund: %w", err)
	}

	if payment.Status != "completed" {
		return nil, errors.New("only completed payments can be refunded")
	}
	// TODO: Integrate with payment gateway to process actual refund
	// For now, just update status
	payment.UpdateStatus("refunded")

	if err := s.paymentRepo.Save(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save refunded payment: %w", err)
	}

	// TODO: Publish PaymentRefunded event (for Order Service to update status, Inventory to restock)

	return &payment_client.PaymentResponse{
		Payment: &payment_client.Payment{
			Id:            payment.ID,
			OrderId:       payment.OrderID,
			UserId:        payment.UserID,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Status:        payment.Status,
			PaymentMethod: payment.PaymentMethod,
			TransactionId: payment.TransactionID,
			CreatedAt:     payment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     payment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ListPayments handles listing payments with pagination and filters.
func (s *paymentService) ListPayments(ctx context.Context, req *payment_client.ListPaymentsRequest) (*payment_client.ListPaymentsResponse, error) {
	payments, totalCount, err := s.paymentRepo.FindAll(ctx, req.GetUserId(), req.GetOrderId(), req.GetStatus(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	paymentResponses := make([]*payment_client.Payment, len(payments))
	for i, p := range payments {
		paymentResponses[i] = &payment_client.Payment{
			Id:            p.ID,
			OrderId:       p.OrderID,
			UserId:        p.UserID,
			Amount:        p.Amount,
			Currency:      p.Currency,
			Status:        p.Status,
			PaymentMethod: p.PaymentMethod,
			TransactionId: p.TransactionID,
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &payment_client.ListPaymentsResponse{
		Payments:   paymentResponses,
		TotalCount: totalCount,
	}, nil
}
