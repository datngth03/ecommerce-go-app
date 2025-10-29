package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/payment_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PaymentClient wraps the gRPC client for payment-service with connection pooling
type PaymentClient struct {
	conn   *grpc.ClientConn         // Legacy: single connection
	pool   *grpcpool.ConnectionPool // New: connection pool
	client pb.PaymentServiceClient
}

// NewPaymentClient creates a new payment service gRPC client (legacy method)
func NewPaymentClient(address string, timeout time.Duration) (*PaymentClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	return &PaymentClient{
		conn:   conn,
		client: pb.NewPaymentServiceClient(conn),
	}, nil
}

// NewPaymentClientWithPool creates a new payment service gRPC client with connection pooling
func NewPaymentClientWithPool(pool *grpcpool.ConnectionPool, timeout time.Duration) (*PaymentClient, error) {
	// Get a connection from the pool to create the client
	conn := pool.Get()

	return &PaymentClient{
		pool:   pool,
		client: pb.NewPaymentServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection (no-op for pooled connections)
func (c *PaymentClient) Close() error {
	// If using connection pool, connections are managed by the pool
	if c.pool != nil {
		return nil
	}

	// Legacy: close single connection
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getClient returns a client using either pooled or direct connection
func (c *PaymentClient) getClient() pb.PaymentServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewPaymentServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.client
}

// ProcessPayment processes a payment
func (c *PaymentClient) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	client := c.getClient()
	return client.ProcessPayment(ctx, req)
}

// ConfirmPayment confirms a pending payment
func (c *PaymentClient) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	client := c.getClient()
	return client.ConfirmPayment(ctx, req)
}

// RefundPayment refunds a payment
func (c *PaymentClient) RefundPayment(ctx context.Context, req *pb.RefundPaymentRequest) (*pb.RefundPaymentResponse, error) {
	client := c.getClient()
	return client.RefundPayment(ctx, req)
}

// GetPayment retrieves payment details
func (c *PaymentClient) GetPayment(ctx context.Context, req *pb.GetPaymentRequest) (*pb.GetPaymentResponse, error) {
	client := c.getClient()
	return client.GetPayment(ctx, req)
}

// GetPaymentByOrder retrieves payment by order ID
func (c *PaymentClient) GetPaymentByOrder(ctx context.Context, req *pb.GetPaymentByOrderRequest) (*pb.GetPaymentByOrderResponse, error) {
	client := c.getClient()
	return client.GetPaymentByOrder(ctx, req)
}

// GetPaymentHistory retrieves payment history
func (c *PaymentClient) GetPaymentHistory(ctx context.Context, req *pb.GetPaymentHistoryRequest) (*pb.GetPaymentHistoryResponse, error) {
	client := c.getClient()
	return client.GetPaymentHistory(ctx, req)
}

// SavePaymentMethod saves a payment method
func (c *PaymentClient) SavePaymentMethod(ctx context.Context, req *pb.SavePaymentMethodRequest) (*pb.SavePaymentMethodResponse, error) {
	client := c.getClient()
	return client.SavePaymentMethod(ctx, req)
}

// GetPaymentMethods retrieves saved payment methods
func (c *PaymentClient) GetPaymentMethods(ctx context.Context, req *pb.GetPaymentMethodsRequest) (*pb.GetPaymentMethodsResponse, error) {
	client := c.getClient()
	return client.GetPaymentMethods(ctx, req)
}
