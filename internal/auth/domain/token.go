// internal/auth/domain/token.go
package domain

import (
	"time"
)

// RefreshToken represents a refresh token entity.
// RefreshToken đại diện cho một thực thể refresh token.
type RefreshToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewRefreshToken creates a new RefreshToken instance.
// NewRefreshToken tạo một thể hiện RefreshToken mới.
func NewRefreshToken(token, userID string, expiresAt time.Time) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}
}
