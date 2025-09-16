// internal/models/user.go
package models

import (
	"time"
)

// User represents the user domain model
type User struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	Password  string    `json:"-" db:"password_hash"` // Never expose password in JSON
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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
