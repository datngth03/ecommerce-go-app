package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/payment_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	conn   *grpc.ClientConn
	client pb.PaymentServiceClient
}

func NewPaymentClient(address string, timeout time.Duration) (*PaymentClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	return &PaymentClient{
		conn:   conn,
		client: pb.NewPaymentServiceClient(conn),
	}, nil
}

func (c *PaymentClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// ProcessPayment processes a payment
func (c *PaymentClient) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	return c.client.ProcessPayment(ctx, req)
}

// ConfirmPayment confirms a pending payment
func (c *PaymentClient) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	return c.client.ConfirmPayment(ctx, req)
}

// RefundPayment refunds a payment
func (c *PaymentClient) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	return c.client.RefundPayment(ctx, req)
}

// GetPayment retrieves payment details
func (c *PaymentClient) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	return c.client.GetPayment(ctx, req)
}

// GetPaymentByOrder retrieves payment by order ID
func (c *PaymentClient) GetPaymentByOrder(ctx context.Context, req *pb.GetPaymentByOrderRequest) (*pb.GetPaymentByOrderResponse, error) {
	return c.client.GetPaymentByOrder(ctx, req)
}

// GetPaymentHistory retrieves payment history
func (c *PaymentClient) GetPaymentHistory(ctx context.Context, req *pb.GetPaymentHistoryRequest) (*pb.GetPaymentHistoryResponse, error) {
	return c.client.GetPaymentHistory(ctx, req)
}

// SavePaymentMethod saves a payment method
func (c *PaymentClient) SavePaymentMethod(ctx context.Context, req *pb.SavePaymentMethodRequest) (*pb.SavePaymentMethodResponse, error) {
	return c.client.SavePaymentMethod(ctx, req)
}

// GetPaymentMethods retrieves saved payment methods
func (c *PaymentClient) GetPaymentMethods(ctx context.Context, req *pb.GetPaymentMethodsRequest) (*pb.GetPaymentMethodsResponse, error) {
	return c.client.GetPaymentMethods(ctx, req)
}
