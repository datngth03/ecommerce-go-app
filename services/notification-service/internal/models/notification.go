package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification statuses
const (
	NotificationStatusPending   = "PENDING"
	NotificationStatusSent      = "SENT"
	NotificationStatusDelivered = "DELIVERED"
	NotificationStatusFailed    = "FAILED"
)

// Notification types
const (
	NotificationTypeEmail = "EMAIL"
	NotificationTypeSMS   = "SMS"
	NotificationTypePush  = "PUSH"
)

// Notification channels
const (
	NotificationChannelSMTP   = "SMTP"
	NotificationChannelTwilio = "TWILIO"
	NotificationChannelFCM    = "FCM"
)

// Notification represents a notification record
type Notification struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string         `gorm:"type:varchar(255);index" json:"user_id"`
	Type         string         `gorm:"type:varchar(50);not null;index" json:"type"`
	Channel      string         `gorm:"type:varchar(50);not null" json:"channel"`
	Recipient    string         `gorm:"type:varchar(255);not null" json:"recipient"`
	Subject      string         `gorm:"type:varchar(500)" json:"subject"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	Status       string         `gorm:"type:varchar(50);not null;index" json:"status"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	TemplateID   string         `gorm:"type:uuid" json:"template_id,omitempty"`
	Metadata     string         `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	SentAt       *time.Time     `json:"sent_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Template represents a notification template
type Template struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Type      string         `gorm:"type:varchar(50);not null" json:"type"`
	Subject   string         `gorm:"type:varchar(500)" json:"subject"`
	Body      string         `gorm:"type:text;not null" json:"body"`
	Variables string         `gorm:"type:jsonb" json:"variables"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// TableName specifies the table name for Template
func (Template) TableName() string {
	return "templates"
}
