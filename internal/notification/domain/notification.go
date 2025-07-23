// internal/notification/domain/notification.go
package domain

import (
	"time"
)

// NotificationRecord represents a record of a sent notification.
// NotificationRecord đại diện cho một bản ghi của một thông báo đã gửi.
type NotificationRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"` // Optional: User associated with the notification
	Type      string    `json:"type"`              // e.g., "email", "sms", "push"
	Recipient string    `json:"recipient"`         // e.g., email address or phone number
	Subject   string    `json:"subject,omitempty"` // For emails
	Message   string    `json:"message"`           // Content of the notification
	Status    string    `json:"status"`            // e.g., "pending", "sent", "failed"
	SentAt    time.Time `json:"sent_at"`
	UpdatedAt time.Time `json:"update_at"`       // Time when the record was last updated
	Error     string    `json:"error,omitempty"` // Error message if sending failed
}

// NewNotificationRecord creates a new NotificationRecord instance.
// NewNotificationRecord tạo một thể hiện NotificationRecord mới.
func NewNotificationRecord(id, userID, notificationType, recipient, subject, message string) *NotificationRecord {
	now := time.Now()
	return &NotificationRecord{
		ID:        id,
		UserID:    userID,
		Type:      notificationType,
		Recipient: recipient,
		Subject:   subject,
		Message:   message,
		Status:    "pending", // Default status
		SentAt:    now,
		UpdatedAt: now, // Add UpdatedAt for consistency, even if not explicitly used in this domain
	}
}

// UpdateStatus updates the notification record's status.
func (nr *NotificationRecord) UpdateStatus(newStatus string, errorMessage string) {
	nr.Status = newStatus
	nr.Error = errorMessage
	nr.UpdatedAt = time.Now()
}
