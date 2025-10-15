package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/notification-service/internal/models"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

// CreateNotification creates a new notification
func (r *notificationRepository) CreateNotification(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetNotification retrieves a notification by ID
func (r *notificationRepository) GetNotification(ctx context.Context, notificationID string) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).Where("id = ?", notificationID).First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// UpdateNotification updates a notification
func (r *notificationRepository) UpdateNotification(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// GetNotificationHistory retrieves notification history
func (r *notificationRepository) GetNotificationHistory(ctx context.Context, userID, notifType string, limit, offset int) ([]*models.Notification, int, error) {
	var notifications []*models.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Notification{})

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, int(total), nil
}

// CreateTemplate creates a new template
func (r *notificationRepository) CreateTemplate(ctx context.Context, template *models.Template) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// GetTemplate retrieves a template by ID
func (r *notificationRepository) GetTemplate(ctx context.Context, templateID string) (*models.Template, error) {
	var template models.Template
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", templateID, true).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetTemplateByName retrieves a template by name
func (r *notificationRepository) GetTemplateByName(ctx context.Context, name string) (*models.Template, error) {
	var template models.Template
	err := r.db.WithContext(ctx).Where("name = ? AND is_active = ?", name, true).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// ListTemplates retrieves all templates
func (r *notificationRepository) ListTemplates(ctx context.Context, notifType string) ([]*models.Template, error) {
	var templates []*models.Template
	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if notifType != "" {
		query = query.Where("type = ?", notifType)
	}

	err := query.Order("created_at DESC").Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// UpdateTemplate updates a template
func (r *notificationRepository) UpdateTemplate(ctx context.Context, template *models.Template) error {
	return r.db.WithContext(ctx).Save(template).Error
}
