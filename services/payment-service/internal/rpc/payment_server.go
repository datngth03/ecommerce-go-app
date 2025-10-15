package rpc

import (
	"context"

	pb "github.com/datngth03/ecommerce-go-app/proto/payment_service"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentServer implements the gRPC payment service
type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

// NewPaymentServer creates a new gRPC payment server
func NewPaymentServer(svc *service.PaymentService) *PaymentServer {
	return &PaymentServer{
		service: svc,
	}
}

// ProcessPayment processes a new payment
func (s *PaymentServer) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	payment, clientSecret, err := s.service.ProcessPayment(
		ctx,
		req.OrderId,
		req.UserId,
		req.Amount,
		req.Currency,
		req.Method,
		req.Metadata,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ProcessPaymentResponse{
		Payment: &pb.Payment{
			Id:                payment.ID,
			OrderId:           payment.OrderID,
			UserId:            payment.UserID,
			Amount:            payment.Amount,
			Currency:          payment.Currency,
			Status:            payment.Status,
			Method:            payment.Method,
			GatewayPaymentId:  payment.GatewayPaymentID,
			GatewayCustomerId: payment.GatewayCustomerID,
			FailureReason:     payment.FailureReason,
			Metadata:          payment.Metadata,
			CreatedAt:         payment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         payment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Success:      true,
		Message:      "Payment processed successfully",
		ClientSecret: clientSecret,
	}, nil
}

// ConfirmPayment confirms a pending payment
func (s *PaymentServer) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	payment, err := s.service.ConfirmPayment(ctx, req.PaymentId, req.PaymentIntentId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.ConfirmPaymentResponse{
		Payment: &pb.Payment{
			Id:       payment.ID,
			OrderId:  payment.OrderID,
			UserId:   payment.UserID,
			Amount:   payment.Amount,
			Currency: payment.Currency,
			Status:   payment.Status,
			Method:   payment.Method,
		},
		Success: true,
		Message: "Payment confirmed successfully",
	}, nil
}

// RefundPayment processes a refund
func (s *PaymentServer) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	refund, err := s.service.RefundPayment(ctx, req.PaymentId, req.Amount, req.Reason)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RefundPaymentResponse{
		Refund: &pb.Refund{
			Id:              refund.ID,
			PaymentId:       refund.PaymentID,
			Amount:          refund.Amount,
			Reason:          refund.Reason,
			Status:          refund.Status,
			GatewayRefundId: refund.GatewayRefundID,
			CreatedAt:       refund.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       refund.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Success: true,
		Message: "Refund processed successfully",
	}, nil
}

// GetPayment retrieves payment details
func (s *PaymentServer) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	payment, err := s.service.GetPayment(ctx, req.PaymentId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convert transactions
	var transactions []*pb.Transaction
	for _, t := range payment.Transactions {
		transactions = append(transactions, &pb.Transaction{
			Id:              t.ID,
			PaymentId:       t.PaymentID,
			TransactionType: t.TransactionType,
			Amount:          t.Amount,
			Status:          t.Status,
			GatewayResponse: t.GatewayResponse,
			CreatedAt:       t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Convert refunds
	var refunds []*pb.Refund
	for _, r := range payment.Refunds {
		refunds = append(refunds, &pb.Refund{
			Id:              r.ID,
			PaymentId:       r.PaymentID,
			Amount:          r.Amount,
			Reason:          r.Reason,
			Status:          r.Status,
			GatewayRefundId: r.GatewayRefundID,
			CreatedAt:       r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetPaymentResponse{
		Payment: &pb.Payment{
			Id:                payment.ID,
			OrderId:           payment.OrderID,
			UserId:            payment.UserID,
			Amount:            payment.Amount,
			Currency:          payment.Currency,
			Status:            payment.Status,
			Method:            payment.Method,
			GatewayPaymentId:  payment.GatewayPaymentID,
			GatewayCustomerId: payment.GatewayCustomerID,
			FailureReason:     payment.FailureReason,
			Metadata:          payment.Metadata,
			CreatedAt:         payment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         payment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Transactions: transactions,
		Refunds:      refunds,
	}, nil
}

// GetPaymentByOrder retrieves payment by order ID
func (s *PaymentServer) GetPaymentByOrder(ctx context.Context, req *pb.GetPaymentByOrderRequest) (*pb.GetPaymentByOrderResponse, error) {
	payment, err := s.service.GetPaymentByOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.GetPaymentByOrderResponse{
		Payment: &pb.Payment{
			Id:       payment.ID,
			OrderId:  payment.OrderID,
			UserId:   payment.UserID,
			Amount:   payment.Amount,
			Currency: payment.Currency,
			Status:   payment.Status,
			Method:   payment.Method,
		},
	}, nil
}

// GetPaymentHistory retrieves user payment history
func (s *PaymentServer) GetPaymentHistory(ctx context.Context, req *pb.GetPaymentHistoryRequest) (*pb.GetPaymentHistoryResponse, error) {
	payments, total, err := s.service.GetPaymentHistory(ctx, req.UserId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbPayments []*pb.Payment
	for _, p := range payments {
		pbPayments = append(pbPayments, &pb.Payment{
			Id:        p.ID,
			OrderId:   p.OrderID,
			UserId:    p.UserID,
			Amount:    p.Amount,
			Currency:  p.Currency,
			Status:    p.Status,
			Method:    p.Method,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetPaymentHistoryResponse{
		Payments: pbPayments,
		Total:    int32(total),
	}, nil
}

// SavePaymentMethod saves a payment method
func (s *PaymentServer) SavePaymentMethod(ctx context.Context, req *pb.SavePaymentMethodRequest) (*pb.SavePaymentMethodResponse, error) {
	method, err := s.service.SavePaymentMethod(ctx, req.UserId, req.MethodType, req.GatewayMethodId, req.IsDefault)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SavePaymentMethodResponse{
		PaymentMethod: &pb.PaymentMethod{
			Id:              method.ID,
			UserId:          method.UserID,
			MethodType:      method.MethodType,
			Last4:           method.Last4,
			Brand:           method.Brand,
			GatewayMethodId: method.GatewayMethodID,
			IsDefault:       method.IsDefault,
			CreatedAt:       method.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Success: true,
		Message: "Payment method saved successfully",
	}, nil
}

// GetPaymentMethods retrieves user's payment methods
func (s *PaymentServer) GetPaymentMethods(ctx context.Context, req *pb.GetPaymentMethodsRequest) (*pb.GetPaymentMethodsResponse, error) {
	methods, err := s.service.GetPaymentMethods(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbMethods []*pb.PaymentMethod
	for _, m := range methods {
		pbMethods = append(pbMethods, &pb.PaymentMethod{
			Id:              m.ID,
			UserId:          m.UserID,
			MethodType:      m.MethodType,
			Last4:           m.Last4,
			Brand:           m.Brand,
			GatewayMethodId: m.GatewayMethodID,
			IsDefault:       m.IsDefault,
			CreatedAt:       m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetPaymentMethodsResponse{
		PaymentMethods: pbMethods,
	}, nil
}

// HandleWebhook handles payment gateway webhooks
func (s *PaymentServer) HandleWebhook(ctx context.Context, req *pb.WebhookEventRequest) (*pb.WebhookEventResponse, error) {
	err := s.service.HandleWebhook(ctx, req.Gateway, req.EventType, req.EventData)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.WebhookEventResponse{
		Success: true,
		Message: "Webhook processed successfully",
	}, nil
}
