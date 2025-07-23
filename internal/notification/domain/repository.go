// internal/notification/domain/repository.go
package domain

import (
	"context"
)

// NotificationRepository defines the interface for storing notification records.
type NotificationRepository interface {
	// Save stores a notification record.
	Save(ctx context.Context, record *NotificationRecord) error

	// FindByID retrieves a notification record by its ID.
	FindByID(ctx context.Context, id string) (*NotificationRecord, error)

	// FindAll retrieves a list of notification records based on filters.
	FindAll(ctx context.Context, userID, notificationType, status string, limit, offset int32) ([]*NotificationRecord, int32, error)
}
