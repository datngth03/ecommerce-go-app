// internal/payment/delivery/grpc/payment_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/payment/application"
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment" // Generated Payment gRPC client
)

// PaymentGRPCServer implements the payment_client.PaymentServiceServer interface.
type PaymentGRPCServer struct {
	payment_client.UnimplementedPaymentServiceServer // Embedded to satisfy all methods
	paymentService                                   application.PaymentService
}

// NewPaymentGRPCServer creates a new instance of PaymentGRPCServer.
func NewPaymentGRPCServer(svc application.PaymentService) *PaymentGRPCServer {
	return &PaymentGRPCServer{
		paymentService: svc,
	}
}

// CreatePayment implements the gRPC CreatePayment method.
func (s *PaymentGRPCServer) CreatePayment(ctx context.Context, req *payment_client.CreatePaymentRequest) (*payment_client.PaymentResponse, error) {
	log.Printf("Received CreatePayment request for Order ID: %s, Amount: %.2f", req.GetOrderId(), req.GetAmount())
	resp, err := s.paymentService.CreatePayment(ctx, req)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetPaymentById implements the gRPC GetPaymentById method.
func (s *PaymentGRPCServer) GetPaymentById(ctx context.Context, req *payment_client.GetPaymentByIdRequest) (*payment_client.PaymentResponse, error) {
	log.Printf("Received GetPaymentById request for ID: %s", req.GetId())
	resp, err := s.paymentService.GetPaymentById(ctx, req)
	if err != nil {
		log.Printf("Error getting payment by ID: %v", err)
		return nil, err
	}
	return resp, nil
}

// ConfirmPayment implements the gRPC ConfirmPayment method.
func (s *PaymentGRPCServer) ConfirmPayment(ctx context.Context, req *payment_client.ConfirmPaymentRequest) (*payment_client.PaymentResponse, error) {
	log.Printf("Received ConfirmPayment request for Payment ID: %s, Transaction ID: %s, Status: %s", req.GetPaymentId(), req.GetTransactionId(), req.GetStatus())
	resp, err := s.paymentService.ConfirmPayment(ctx, req)
	if err != nil {
		log.Printf("Error confirming payment: %v", err)
		return nil, err
	}
	return resp, nil
}

// RefundPayment implements the gRPC RefundPayment method.
func (s *PaymentGRPCServer) RefundPayment(ctx context.Context, req *payment_client.RefundPaymentRequest) (*payment_client.PaymentResponse, error) {
	log.Printf("Received RefundPayment request for Payment ID: %s, Refund Amount: %.2f", req.GetPaymentId(), req.GetRefundAmount())
	resp, err := s.paymentService.RefundPayment(ctx, req)
	if err != nil {
		log.Printf("Error refunding payment: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListPayments implements the gRPC ListPayments method.
func (s *PaymentGRPCServer) ListPayments(ctx context.Context, req *payment_client.ListPaymentsRequest) (*payment_client.ListPaymentsResponse, error) {
	log.Printf("Received ListPayments request (User ID: %s, Order ID: %s, Status: %s)", req.GetUserId(), req.GetOrderId(), req.GetStatus())
	resp, err := s.paymentService.ListPayments(ctx, req)
	if err != nil {
		log.Printf("Error listing payments: %v", err)
		return nil, err
	}
	return resp, nil
}
