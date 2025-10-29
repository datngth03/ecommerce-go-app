package rpc

import (
	"context"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/notification_service"
	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NotificationServer implements the gRPC notification service
type NotificationServer struct {
	pb.UnimplementedNotificationServiceServer
	service *service.NotificationService
}

// NewNotificationServer creates a new gRPC notification server
func NewNotificationServer(svc *service.NotificationService) *NotificationServer {
	return &NotificationServer{
		service: svc,
	}
}

// SendEmail sends an email notification
func (s *NotificationServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	start := time.Now()

	notification, err := s.service.SendEmail(
		ctx,
		req.UserId,
		req.Recipient,
		req.Subject,
		req.Body,
		req.TemplateId,
		req.Variables,
	)

	duration := time.Since(start)
	grpcStatus := "success"
	notifStatus := "sent"

	if err != nil {
		grpcStatus = "error"
		notifStatus = "failed"
		metrics.RecordGRPCRequest("SendEmail", grpcStatus, duration)
		metrics.RecordNotificationSent("email", notifStatus, duration)
		metrics.RecordEmailSent(notifStatus)
		return &pb.SendEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	metrics.RecordGRPCRequest("SendEmail", grpcStatus, duration)
	metrics.RecordNotificationSent("email", notification.Status, duration)
	metrics.RecordEmailSent(notification.Status)

	return &pb.SendEmailResponse{
		Notification: &pb.Notification{
			Id:           notification.ID,
			UserId:       notification.UserID,
			Type:         notification.Type,
			Channel:      notification.Channel,
			Recipient:    notification.Recipient,
			Subject:      notification.Subject,
			Content:      notification.Content,
			Status:       notification.Status,
			ErrorMessage: notification.ErrorMessage,
			TemplateId:   notification.TemplateID,
			Metadata:     notification.Metadata,
			CreatedAt:    notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		Success: true,
		Message: "Email sent successfully",
	}, nil
}

// SendSMS sends an SMS notification
func (s *NotificationServer) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	start := time.Now()

	notification, err := s.service.SendSMS(
		ctx,
		req.UserId,
		req.Recipient,
		req.Message,
		req.TemplateId,
		req.Variables,
	)

	duration := time.Since(start)
	grpcStatus := "success"
	notifStatus := "sent"

	if err != nil {
		grpcStatus = "error"
		notifStatus = "failed"
		metrics.RecordGRPCRequest("SendSMS", grpcStatus, duration)
		metrics.RecordNotificationSent("sms", notifStatus, duration)
		metrics.RecordSMSSent(notifStatus)
		return &pb.SendSMSResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	metrics.RecordGRPCRequest("SendSMS", grpcStatus, duration)
	metrics.RecordNotificationSent("sms", notification.Status, duration)
	metrics.RecordSMSSent(notification.Status)

	return &pb.SendSMSResponse{
		Notification: &pb.Notification{
			Id:        notification.ID,
			UserId:    notification.UserID,
			Type:      notification.Type,
			Channel:   notification.Channel,
			Recipient: notification.Recipient,
			Content:   notification.Content,
			Status:    notification.Status,
		},
		Success: true,
		Message: "SMS sent successfully",
	}, nil
}

// GetNotification retrieves a notification
func (s *NotificationServer) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.GetNotificationResponse, error) {
	notification, err := s.service.GetNotification(ctx, req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	sentAt := ""
	if notification.SentAt != nil {
		sentAt = notification.SentAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return &pb.GetNotificationResponse{
		Notification: &pb.Notification{
			Id:           notification.ID,
			UserId:       notification.UserID,
			Type:         notification.Type,
			Channel:      notification.Channel,
			Recipient:    notification.Recipient,
			Subject:      notification.Subject,
			Content:      notification.Content,
			Status:       notification.Status,
			ErrorMessage: notification.ErrorMessage,
			TemplateId:   notification.TemplateID,
			Metadata:     notification.Metadata,
			CreatedAt:    notification.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SentAt:       sentAt,
		},
	}, nil
}

// GetNotificationHistory retrieves notification history
func (s *NotificationServer) GetNotificationHistory(ctx context.Context, req *pb.GetNotificationHistoryRequest) (*pb.GetNotificationHistoryResponse, error) {
	notifications, total, err := s.service.GetNotificationHistory(ctx, req.UserId, req.Type, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbNotifications []*pb.Notification
	for _, n := range notifications {
		sentAt := ""
		if n.SentAt != nil {
			sentAt = n.SentAt.Format("2006-01-02T15:04:05Z07:00")
		}

		pbNotifications = append(pbNotifications, &pb.Notification{
			Id:        n.ID,
			UserId:    n.UserID,
			Type:      n.Type,
			Channel:   n.Channel,
			Recipient: n.Recipient,
			Subject:   n.Subject,
			Status:    n.Status,
			CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			SentAt:    sentAt,
		})
	}

	return &pb.GetNotificationHistoryResponse{
		Notifications: pbNotifications,
		Total:         int32(total),
	}, nil
}
