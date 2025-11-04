// internal/models/user.go
package models

import (
	"time"
)

// User represents the user domain model
type User struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string    `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Name      string    `json:"name" gorm:"type:varchar(100);not null"`
	Phone     string    `json:"phone" gorm:"type:varchar(20)"`
	Password  string    `json:"password_hash" gorm:"column:password_hash;type:varchar(255);not null"` // Changed from json:"-" to allow Redis cache serialization
	IsActive  bool      `json:"is_active" gorm:"default:true;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// UserUpdateData represents data for updating a user
type UserUpdateData struct {
	ID       int64   `json:"id"`
	Name     *string `json:"name,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Phone    string `json:"phone" validate:"omitempty,min=10,max=20"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// ResetPasswordRequest represents the reset password request
type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
