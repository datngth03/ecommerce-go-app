// internal/auth/domain/repository.go
package domain

import (
	"context"
)

// RefreshTokenRepository defines the interface for refresh token data operations.
type RefreshTokenRepository interface {
	// Save stores a refresh token.
	Save(ctx context.Context, token *RefreshToken) error

	// FindByToken retrieves a refresh token by its value.
	FindByToken(ctx context.Context, token string) (*RefreshToken, error)

	// Delete removes a refresh token.
	Delete(ctx context.Context, token string) error

	// DeleteByUserID removes all refresh tokens for a given user.
	DeleteByUserID(ctx context.Context, userID string) error
}
