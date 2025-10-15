package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/email"
	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	repo         repository.NotificationRepository
	emailService *email.EmailService
}

// NewNotificationService creates a new notification service
func NewNotificationService(repo repository.NotificationRepository, emailService *email.EmailService) *NotificationService {
	return &NotificationService{
		repo:         repo,
		emailService: emailService,
	}
}

// SendEmail sends an email notification
func (s *NotificationService) SendEmail(ctx context.Context, userID, recipient, subject, body, templateID string, variables map[string]string) (*models.Notification, error) {
	// If template is specified, use it
	if templateID != "" {
		template, err := s.repo.GetTemplate(ctx, templateID)
		if err != nil {
			return nil, fmt.Errorf("template not found: %w", err)
		}

		subject = s.emailService.RenderTemplate(template.Subject, variables)
		body = s.emailService.RenderTemplate(template.Body, variables)
	}

	// Create notification record
	metadataJSON, _ := json.Marshal(variables)
	notification := &models.Notification{
		UserID:     userID,
		Type:       models.NotificationTypeEmail,
		Channel:    models.NotificationChannelSMTP,
		Recipient:  recipient,
		Subject:    subject,
		Content:    body,
		Status:     models.NotificationStatusPending,
		TemplateID: templateID,
		Metadata:   string(metadataJSON),
	}

	err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send email
	err = s.emailService.SendEmail(recipient, subject, body)
	if err != nil {
		// Update status to failed
		notification.Status = models.NotificationStatusFailed
		notification.ErrorMessage = err.Error()
		s.repo.UpdateNotification(ctx, notification)
		return notification, fmt.Errorf("failed to send email: %w", err)
	}

	// Update status to sent
	now := time.Now()
	notification.Status = models.NotificationStatusSent
	notification.SentAt = &now
	s.repo.UpdateNotification(ctx, notification)

	return notification, nil
}

// SendSMS sends an SMS notification (stub - not implemented)
func (s *NotificationService) SendSMS(ctx context.Context, userID, recipient, message, templateID string, variables map[string]string) (*models.Notification, error) {
	// Create notification record
	notification := &models.Notification{
		UserID:       userID,
		Type:         models.NotificationTypeSMS,
		Channel:      models.NotificationChannelTwilio,
		Recipient:    recipient,
		Content:      message,
		Status:       models.NotificationStatusFailed,
		ErrorMessage: "SMS sending not implemented",
	}

	err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// TODO: Implement Twilio integration
	return notification, fmt.Errorf("SMS sending not implemented")
}

// SendBulkEmail sends email to multiple recipients
func (s *NotificationService) SendBulkEmail(ctx context.Context, recipients []string, subject, body, templateID string, variables map[string]string) (int, int, error) {
	// If template is specified, use it
	if templateID != "" {
		template, err := s.repo.GetTemplate(ctx, templateID)
		if err != nil {
			return 0, 0, fmt.Errorf("template not found: %w", err)
		}

		subject = s.emailService.RenderTemplate(template.Subject, variables)
		body = s.emailService.RenderTemplate(template.Body, variables)
	}

	return s.emailService.SendBulkEmail(recipients, subject, body)
}

// GetNotification retrieves a notification
func (s *NotificationService) GetNotification(ctx context.Context, notificationID string) (*models.Notification, error) {
	return s.repo.GetNotification(ctx, notificationID)
}

// GetNotificationHistory retrieves notification history
func (s *NotificationService) GetNotificationHistory(ctx context.Context, userID, notifType string, limit, offset int) ([]*models.Notification, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetNotificationHistory(ctx, userID, notifType, limit, offset)
}

// CreateTemplate creates a notification template
func (s *NotificationService) CreateTemplate(ctx context.Context, name, notifType, subject, body string, variables map[string]string) (*models.Template, error) {
	variablesJSON, _ := json.Marshal(variables)
	template := &models.Template{
		Name:      name,
		Type:      notifType,
		Subject:   subject,
		Body:      body,
		Variables: string(variablesJSON),
		IsActive:  true,
	}

	err := s.repo.CreateTemplate(ctx, template)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return template, nil
}

// GetTemplate retrieves a template
func (s *NotificationService) GetTemplate(ctx context.Context, templateID string) (*models.Template, error) {
	return s.repo.GetTemplate(ctx, templateID)
}

// ListTemplates lists all templates
func (s *NotificationService) ListTemplates(ctx context.Context, notifType string) ([]*models.Template, error) {
	return s.repo.ListTemplates(ctx, notifType)
}
