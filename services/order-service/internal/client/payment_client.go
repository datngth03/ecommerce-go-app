package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/payment_service"
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PaymentClient struct {
	conn   *grpc.ClientConn
	client pb.PaymentServiceClient
	pool   *grpcpool.ConnectionPool // Connection pool support
}

func NewPaymentClient(endpoint sharedConfig.ServiceEndpoint) (*PaymentClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(endpoint.GRPCAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	return &PaymentClient{
		conn:   conn,
		client: pb.NewPaymentServiceClient(conn),
	}, nil
}

// NewPaymentClientWithPool creates a new payment client with connection pooling support
func NewPaymentClientWithPool(endpoint sharedConfig.ServiceEndpoint, poolManager *grpcpool.Manager) (*PaymentClient, error) {
	pool, exists := poolManager.Get("payment")
	if !exists {
		poolConfig := grpcpool.DefaultPoolConfig(endpoint.GRPCAddr)
		var err error
		pool, err = poolManager.GetOrCreate("payment", poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment service pool: %w", err)
		}
	}

	return &PaymentClient{
		pool: pool,
	}, nil
}

func (c *PaymentClient) Close() error {
	// If using pool, don't close individual connections
	if c.pool != nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getClient returns a gRPC client, using pool if available
func (c *PaymentClient) getClient() (pb.PaymentServiceClient, error) {
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewPaymentServiceClient(conn), nil
	}
	return c.client, nil
}

// ProcessPayment initiates a payment for an order
func (c *PaymentClient) ProcessPayment(ctx context.Context, req *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.ProcessPayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	return resp, nil
}

// ConfirmPayment confirms a payment (webhook/callback)
func (c *PaymentClient) ConfirmPayment(ctx context.Context, paymentID string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	resp, err := client.ConfirmPayment(ctx, &pb.ConfirmPaymentRequest{
		PaymentId: paymentID,
	})
	if err != nil {
		return fmt.Errorf("failed to confirm payment: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to confirm payment: %s", resp.Message)
	}

	return nil
}

// GetPayment retrieves payment details
func (c *PaymentClient) GetPayment(ctx context.Context, paymentID string) (*pb.Payment, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetPayment(ctx, &pb.GetPaymentRequest{
		PaymentId: paymentID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return resp.Payment, nil
}

// GetPaymentByOrderID retrieves payment by order ID
func (c *PaymentClient) GetPaymentByOrderID(ctx context.Context, orderID string) (*pb.Payment, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetPaymentByOrder(ctx, &pb.GetPaymentByOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get payment by order ID: %w", err)
	}

	return resp.Payment, nil
}

// RefundPayment initiates a refund
func (c *PaymentClient) RefundPayment(ctx context.Context, paymentID string, amount float64, reason string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	resp, err := client.RefundPayment(ctx, &pb.RefundPaymentRequest{
		PaymentId: paymentID,
		Amount:    amount,
		Reason:    reason,
	})
	if err != nil {
		return fmt.Errorf("failed to refund payment: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to refund payment: %s", resp.Message)
	}

	return nil
}
