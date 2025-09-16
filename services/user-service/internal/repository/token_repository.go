package repository

import (
	"context"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
)

// TokenRepositoryInterface defines the contract for token storage and management.
type TokenRepositoryInterface interface {
	// --- Refresh Token Management ---

	// StoreRefreshToken saves a refresh token with its expiration.
	StoreRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error

	// GetRefreshToken retrieves refresh token data by the token string.
	GetRefreshToken(ctx context.Context, token string) (*utils.RefreshTokenData, error)

	// DeleteRefreshToken removes a specific refresh token.
	DeleteRefreshToken(ctx context.Context, token string) error

	// DeleteAllUserRefreshTokens removes all refresh tokens associated with a user.
	// This is useful for security events like a password change.
	DeleteAllUserRefreshTokens(ctx context.Context, userID int64) error

	// --- Access Token Blacklist ---

	// BlacklistToken adds an access token to a blacklist until it expires.
	// This is used for logging out users before their token's natural expiration.
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error

	// IsTokenBlacklisted checks if an access token is in the blacklist.
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)

	// --- Password Reset Token Management ---

	// StorePasswordResetToken saves a password reset token.
	StorePasswordResetToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error

	// GetPasswordResetToken retrieves data for a given password reset token.
	GetPasswordResetToken(ctx context.Context, token string) (*utils.PasswordResetTokenData, error)

	// DeletePasswordResetToken removes a password reset token after it has been used.
	DeletePasswordResetToken(ctx context.Context, token string) error
}
