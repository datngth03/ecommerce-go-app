// internal/notification/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"log" // Temporarily using log for errors

	// "time"

	"github.com/google/uuid" // For generating UUIDs

	"github.com/datngth03/ecommerce-go-app/internal/notification/domain"
	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification" // Generated Notification gRPC client
)

// NotificationService defines the application service interface for sending notifications.
// NotificationService định nghĩa interface dịch vụ ứng dụng để gửi thông báo.
type NotificationService interface {
	SendEmail(ctx context.Context, req *notification_client.SendEmailRequest) (*notification_client.SendNotificationResponse, error)
	SendSMS(ctx context.Context, req *notification_client.SendSMSRequest) (*notification_client.SendNotificationResponse, error)
	SendPushNotification(ctx context.Context, req *notification_client.SendPushNotificationRequest) (*notification_client.SendNotificationResponse, error)
}

// notificationService implements the NotificationService interface.
// notificationService triển khai interface NotificationService.
type notificationService struct {
	notificationRepo domain.NotificationRepository
	// TODO: Add clients for third-party email/SMS/push services (e.g., SendGrid, Twilio, FCM)
	// Thêm các client cho dịch vụ email/SMS/push bên thứ ba (ví dụ: SendGrid, Twilio, FCM)
}

// NewNotificationService creates a new instance of NotificationService.
// NewNotificationService tạo một thể hiện mới của NotificationService.
func NewNotificationService(repo domain.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: repo,
	}
}

// SendEmail handles sending an email notification.
// SendEmail xử lý việc gửi thông báo email.
func (s *notificationService) SendEmail(ctx context.Context, req *notification_client.SendEmailRequest) (*notification_client.SendNotificationResponse, error) {
	if req.GetRecipientEmail() == "" || (req.GetBodyHtml() == "" && req.GetBodyPlain() == "") {
		return nil, errors.New("recipient email and at least one body (html or plain) are required")
	}

	notificationID := uuid.New().String()
	record := domain.NewNotificationRecord(
		notificationID,
		req.GetUserId(),
		"email",
		req.GetRecipientEmail(),
		req.GetSubject(),
		req.GetBodyPlain(), // Store plain text body for record, or choose html
	)

	// TODO: Integrate with a real email sending service (e.g., SendGrid, Mailgun)
	// For now, simulate sending.
	log.Printf("Simulating sending email to %s with subject: %s", req.GetRecipientEmail(), req.GetSubject())

	// Simulate success or failure
	var sendErr error
	// if some_condition_fails {
	// 	sendErr = errors.New("simulated email sending failure")
	// }

	if sendErr != nil {
		record.UpdateStatus("failed", sendErr.Error())
		s.notificationRepo.Save(ctx, record) // Save failure record
		return nil, fmt.Errorf("failed to send email: %w", sendErr)
	}

	record.UpdateStatus("sent", "")
	if err := s.notificationRepo.Save(ctx, record); err != nil {
		log.Printf("Warning: Failed to save notification record for email %s: %v", notificationID, err)
		// Decide how to handle this: retry, manual intervention
	}

	return &notification_client.SendNotificationResponse{
		Success:        true,
		Message:        "Email sent successfully (simulated)",
		NotificationId: notificationID,
	}, nil
}

// SendSMS handles sending an SMS notification.
// SendSMS xử lý việc gửi thông báo SMS.
func (s *notificationService) SendSMS(ctx context.Context, req *notification_client.SendSMSRequest) (*notification_client.SendNotificationResponse, error) {
	if req.GetRecipientPhoneNumber() == "" || req.GetMessage() == "" {
		return nil, errors.New("recipient phone number and message are required")
	}

	notificationID := uuid.New().String()
	record := domain.NewNotificationRecord(
		notificationID,
		req.GetUserId(),
		"sms",
		req.GetRecipientPhoneNumber(),
		"", // SMS usually doesn't have a subject
		req.GetMessage(),
	)

	// TODO: Integrate with a real SMS sending service (e.g., Twilio, Nexmo)
	// For now, simulate sending.
	log.Printf("Simulating sending SMS to %s with message: %s", req.GetRecipientPhoneNumber(), req.GetMessage())

	var sendErr error
	// if some_condition_fails {
	// 	sendErr = errors.New("simulated SMS sending failure")
	// }

	if sendErr != nil {
		record.UpdateStatus("failed", sendErr.Error())
		s.notificationRepo.Save(ctx, record)
		return nil, fmt.Errorf("failed to send SMS: %w", sendErr)
	}

	record.UpdateStatus("sent", "")
	if err := s.notificationRepo.Save(ctx, record); err != nil {
		log.Printf("Warning: Failed to save notification record for SMS %s: %v", notificationID, err)
	}

	return &notification_client.SendNotificationResponse{
		Success:        true,
		Message:        "SMS sent successfully (simulated)",
		NotificationId: notificationID,
	}, nil
}

// SendPushNotification handles sending a push notification. (Placeholder)
// SendPushNotification xử lý việc gửi thông báo đẩy. (Placeholder)
func (s *notificationService) SendPushNotification(ctx context.Context, req *notification_client.SendPushNotificationRequest) (*notification_client.SendNotificationResponse, error) {
	if req.GetUserId() == "" || req.GetDeviceToken() == "" || req.GetTitle() == "" || req.GetBody() == "" {
		return nil, errors.New("user ID, device token, title, and body are required for push notification")
	}

	notificationID := uuid.New().String()
	record := domain.NewNotificationRecord(
		notificationID,
		req.GetUserId(),
		"push",
		req.GetDeviceToken(), // Recipient is device token for push
		req.GetTitle(),
		req.GetBody(),
	)

	// TODO: Integrate with a real push notification service (e.g., Firebase Cloud Messaging, Apple Push Notification Service)
	// For now, simulate sending.
	log.Printf("Simulating sending push notification to User %s (Device: %s) with title: %s", req.GetUserId(), req.GetDeviceToken(), req.GetTitle())

	var sendErr error
	// if some_condition_fails {
	// 	sendErr = errors.New("simulated push notification sending failure")
	// }

	if sendErr != nil {
		record.UpdateStatus("failed", sendErr.Error())
		s.notificationRepo.Save(ctx, record)
		return nil, fmt.Errorf("failed to send push notification: %w", sendErr)
	}

	record.UpdateStatus("sent", "")
	if err := s.notificationRepo.Save(ctx, record); err != nil {
		log.Printf("Warning: Failed to save notification record for push %s: %v", notificationID, err)
	}

	return &notification_client.SendNotificationResponse{
		Success:        true,
		Message:        "Push notification sent successfully (simulated)",
		NotificationId: notificationID,
	}, nil
}
