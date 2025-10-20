package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/notification_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NotificationClient struct {
	conn   *grpc.ClientConn
	client pb.NotificationServiceClient
}

func NewNotificationClient(address string, timeout time.Duration) (*NotificationClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
	}

	return &NotificationClient{
		conn:   conn,
		client: pb.NewNotificationServiceClient(conn),
	}, nil
}

func (c *NotificationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendEmail sends an email notification
func (c *NotificationClient) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	return c.client.SendEmail(ctx, req)
}

// SendSMS sends an SMS notification
func (c *NotificationClient) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	return c.client.SendSMS(ctx, req)
}

// GetNotification retrieves notification details
func (c *NotificationClient) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.GetNotificationResponse, error) {
	return c.client.GetNotification(ctx, req)
}

// GetNotificationHistory retrieves notification history
func (c *NotificationClient) GetNotificationHistory(ctx context.Context, req *pb.GetNotificationHistoryRequest) (*pb.GetNotificationHistoryResponse, error) {
	return c.client.GetNotificationHistory(ctx, req)
}
