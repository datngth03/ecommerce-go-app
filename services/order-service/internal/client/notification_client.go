package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/notification_service"
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	conn   *grpc.ClientConn
	client pb.NotificationServiceClient
	pool   *grpcpool.ConnectionPool // Connection pool support
}

func NewNotificationClient(endpoint sharedConfig.ServiceEndpoint) (*NotificationClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(endpoint.GRPCAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}

	return &NotificationClient{
		conn:   conn,
		client: pb.NewNotificationServiceClient(conn),
	}, nil
}

// NewNotificationClientWithPool creates a new notification client with connection pooling support
func NewNotificationClientWithPool(endpoint sharedConfig.ServiceEndpoint, poolManager *grpcpool.Manager) (*NotificationClient, error) {
	pool, exists := poolManager.Get("notification")
	if !exists {
		poolConfig := grpcpool.DefaultPoolConfig(endpoint.GRPCAddr)
		var err error
		pool, err = poolManager.GetOrCreate("notification", poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create notification service pool: %w", err)
		}
	}

	return &NotificationClient{
		pool: pool,
	}, nil
}

func (c *NotificationClient) Close() error {
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
func (c *NotificationClient) getClient() (pb.NotificationServiceClient, error) {
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewNotificationServiceClient(conn), nil
	}
	return c.client, nil
}

// SendEmail sends an email notification
func (c *NotificationClient) SendEmail(ctx context.Context, userID, recipient, subject, body string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	resp, err := client.SendEmail(ctx, &pb.SendEmailRequest{
		UserId:    userID,
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
	})
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to send email: %s", resp.Message)
	}

	return nil
}

// SendSMS sends an SMS notification
func (c *NotificationClient) SendSMS(ctx context.Context, userID, recipient, message string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	resp, err := client.SendSMS(ctx, &pb.SendSMSRequest{
		UserId:    userID,
		Recipient: recipient,
		Message:   message,
	})
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to send SMS: %s", resp.Message)
	}

	return nil
}

// SendOrderConfirmation sends order confirmation email/SMS
func (c *NotificationClient) SendOrderConfirmation(ctx context.Context, userID, userEmail, orderID string) error {
	subject := "Order Confirmation"
	body := fmt.Sprintf("Thank you for your order! Your order ID is: %s", orderID)

	return c.SendEmail(ctx, userID, userEmail, subject, body)
}

// SendOrderStatusUpdate sends order status update notification
func (c *NotificationClient) SendOrderStatusUpdate(ctx context.Context, userID, userEmail, orderID, status string) error {
	subject := "Order Status Update"
	body := fmt.Sprintf("Your order %s status has been updated to: %s", orderID, status)

	return c.SendEmail(ctx, userID, userEmail, subject, body)
}

// SendPaymentConfirmation sends payment confirmation notification
func (c *NotificationClient) SendPaymentConfirmation(ctx context.Context, userID, userEmail, paymentID string, amount float32) error {
	subject := "Payment Confirmation"
	body := fmt.Sprintf("Your payment of $%.2f has been confirmed. Payment ID: %s", amount, paymentID)

	return c.SendEmail(ctx, userID, userEmail, subject, body)
}
