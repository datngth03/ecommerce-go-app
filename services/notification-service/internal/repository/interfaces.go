package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/models"
)

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	// Notification operations
	CreateNotification(ctx context.Context, notification *models.Notification) error
	GetNotification(ctx context.Context, notificationID string) (*models.Notification, error)
	UpdateNotification(ctx context.Context, notification *models.Notification) error
	GetNotificationHistory(ctx context.Context, userID, notifType string, limit, offset int) ([]*models.Notification, int, error)

	// Template operations
	CreateTemplate(ctx context.Context, template *models.Template) error
	GetTemplate(ctx context.Context, templateID string) (*models.Template, error)
	GetTemplateByName(ctx context.Context, name string) (*models.Template, error)
	ListTemplates(ctx context.Context, notifType string) ([]*models.Template, error)
	UpdateTemplate(ctx context.Context, template *models.Template) error
}
