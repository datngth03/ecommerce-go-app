// internal/notification/delivery/grpc/notification_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/notification/application"
	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification" // Generated Notification gRPC client
)

// NotificationGRPCServer implements the notification_client.NotificationServiceServer interface.
type NotificationGRPCServer struct {
	notification_client.UnimplementedNotificationServiceServer // Embedded to satisfy all methods
	notificationService                                        application.NotificationService
}

// NewNotificationGRPCServer creates a new instance of NotificationGRPCServer.
func NewNotificationGRPCServer(svc application.NotificationService) *NotificationGRPCServer {
	return &NotificationGRPCServer{
		notificationService: svc,
	}
}

// SendEmail implements the gRPC SendEmail method.
func (s *NotificationGRPCServer) SendEmail(ctx context.Context, req *notification_client.SendEmailRequest) (*notification_client.SendNotificationResponse, error) {
	log.Printf("Received SendEmail request for recipient: %s, subject: %s", req.GetRecipientEmail(), req.GetSubject())
	resp, err := s.notificationService.SendEmail(ctx, req)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return nil, err
	}
	return resp, nil
}

// SendSMS implements the gRPC SendSMS method.
func (s *NotificationGRPCServer) SendSMS(ctx context.Context, req *notification_client.SendSMSRequest) (*notification_client.SendNotificationResponse, error) {
	log.Printf("Received SendSMS request for recipient: %s", req.GetRecipientPhoneNumber())
	resp, err := s.notificationService.SendSMS(ctx, req)
	if err != nil {
		log.Printf("Error sending SMS: %v", err)
		return nil, err
	}
	return resp, nil
}

// SendPushNotification implements the gRPC SendPushNotification method.
func (s *NotificationGRPCServer) SendPushNotification(ctx context.Context, req *notification_client.SendPushNotificationRequest) (*notification_client.SendNotificationResponse, error) {
	log.Printf("Received SendPushNotification request for user: %s, title: %s", req.GetUserId(), req.GetTitle())
	resp, err := s.notificationService.SendPushNotification(ctx, req)
	if err != nil {
		log.Printf("Error sending push notification: %v", err)
		return nil, err
	}
	return resp, nil
}
