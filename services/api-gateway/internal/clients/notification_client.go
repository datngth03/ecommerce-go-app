package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/notification_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NotificationClient wraps the gRPC client for notification-service with connection pooling
type NotificationClient struct {
	conn   *grpc.ClientConn         // Legacy: single connection
	pool   *grpcpool.ConnectionPool // New: connection pool
	client pb.NotificationServiceClient
}

// NewNotificationClient creates a new notification service gRPC client (legacy method)
func NewNotificationClient(address string, timeout time.Duration) (*NotificationClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}

	return &NotificationClient{
		conn:   conn,
		client: pb.NewNotificationServiceClient(conn),
	}, nil
}

// NewNotificationClientWithPool creates a new notification service gRPC client with connection pooling
func NewNotificationClientWithPool(pool *grpcpool.ConnectionPool, timeout time.Duration) (*NotificationClient, error) {
	// Get a connection from the pool to create the client
	conn := pool.Get()

	return &NotificationClient{
		pool:   pool,
		client: pb.NewNotificationServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection (no-op for pooled connections)
func (c *NotificationClient) Close() error {
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
func (c *NotificationClient) getClient() pb.NotificationServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewNotificationServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.client
}

// SendEmail sends an email notification
func (c *NotificationClient) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	client := c.getClient()
	return client.SendEmail(ctx, req)
}

// SendSMS sends an SMS notification
func (c *NotificationClient) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	client := c.getClient()
	return client.SendSMS(ctx, req)
}

// GetNotification retrieves notification details
func (c *NotificationClient) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.GetNotificationResponse, error) {
	client := c.getClient()
	return client.GetNotification(ctx, req)
}

// GetNotificationHistory retrieves notification history
func (c *NotificationClient) GetNotificationHistory(ctx context.Context, req *pb.GetNotificationHistoryRequest) (*pb.GetNotificationHistoryResponse, error) {
	client := c.getClient()
	return client.GetNotificationHistory(ctx, req)
}
