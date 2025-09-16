package service

import (
	"context"
	// "fmt"
	"log"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceInterface defines the auth service contract
type AuthServiceInterface interface {
	// Token management
	GenerateTokenPair(ctx context.Context, userID int64, email string) (*utils.TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (*utils.JWTClaims, error)
	ValidateRefreshToken(ctx context.Context, refreshToken string) (*utils.RefreshTokenData, error)
	UpdateRefreshToken(ctx context.Context, userID int64, oldToken, newToken string, newExpiresAt time.Time) error

	// Token storage & invalidation
	StoreRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	InvalidateUserTokens(ctx context.Context, accessToken string, refreshToken *string) error
	InvalidateAllUserTokens(ctx context.Context, userID int64) error

	// Password reset
	GeneratePasswordResetToken(ctx context.Context, userID int64) (string, time.Time, error)
	StorePasswordResetToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error
	ValidatePasswordResetToken(ctx context.Context, token string) (*utils.PasswordResetTokenData, error)
	InvalidatePasswordResetToken(ctx context.Context, token string) error
}

// AuthService implements the AuthServiceInterface
type AuthService struct {
	userRepo        repository.UserRepositoryInterface
	tokenRepo       repository.TokenRepositoryInterface
	jwtSecret       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	resetTokenTTL   time.Duration
}

// NewAuthService creates a new AuthService instance
func NewAuthService(
	userRepo repository.UserRepositoryInterface,
	tokenRepo repository.TokenRepositoryInterface,
	jwtSecret string,
	accessTokenTTL, refreshTokenTTL, resetTokenTTL time.Duration,
) AuthServiceInterface {
	return &AuthService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		jwtSecret:       jwtSecret,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		resetTokenTTL:   resetTokenTTL,
	}
}

// =================================
// Token Management Implementation
// =================================

// GenerateTokenPair generates access and refresh token pair by calling the utility function.
func (s *AuthService) GenerateTokenPair(ctx context.Context, userID int64, email string) (*utils.TokenPair, error) {
	log.Printf("AuthService: Generating token pair for user %d", userID)

	tokenPair, err := utils.GenerateTokenPair(userID, email, s.jwtSecret, s.accessTokenTTL, s.refreshTokenTTL)
	if err != nil {
		log.Printf("AuthService: Failed to generate token pair for user %d: %v", userID, err)
		return nil, status.Error(codes.Internal, "could not generate token pair")
	}

	return tokenPair, nil
}

// ValidateAccessToken validates and parses an access token.
func (s *AuthService) ValidateAccessToken(ctx context.Context, token string) (*utils.JWTClaims, error) {
	log.Printf("AuthService: Validating access token")

	claims, err := utils.ValidateJWT(token, s.jwtSecret)
	if err != nil {
		log.Printf("AuthService: Access token validation failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired access token")
	}

	// Check if token is blacklisted (e.g., after logout)
	isBlacklisted, err := s.tokenRepo.IsTokenBlacklisted(ctx, token)
	if err != nil {
		log.Printf("AuthService: Failed to check token blacklist: %v", err)
		return nil, status.Error(codes.Internal, "token validation error")
	}

	if isBlacklisted {
		log.Printf("AuthService: Token is blacklisted")
		return nil, status.Error(codes.Unauthenticated, "token has been invalidated")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token from storage.
func (s *AuthService) ValidateRefreshToken(ctx context.Context, refreshToken string) (*utils.RefreshTokenData, error) {
	log.Printf("AuthService: Validating refresh token")

	tokenData, err := s.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Printf("AuthService: Refresh token validation failed: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
	}

	// Double-check expiration, although Redis TTL should handle this.
	if time.Now().After(tokenData.ExpiresAt) {
		log.Printf("AuthService: Refresh token is expired (logic check)")
		// Clean up expired token just in case
		_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
		return nil, status.Error(codes.Unauthenticated, "refresh token has expired")
	}

	return tokenData, nil
}

// UpdateRefreshToken atomically deletes an old refresh token and stores a new one.
func (s *AuthService) UpdateRefreshToken(ctx context.Context, userID int64, oldToken, newToken string, newExpiresAt time.Time) error {
	log.Printf("AuthService: Updating refresh token for user %d", userID)

	// Xóa token cũ
	// Chúng ta có thể bỏ qua lỗi ở đây vì nếu token cũ không tồn tại, đó không phải là vấn đề.
	if err := s.tokenRepo.DeleteRefreshToken(ctx, oldToken); err != nil {
		log.Printf("AuthService: Could not delete old refresh token '%s' during update (this may be okay): %v", oldToken, err)
	}

	// Lưu token mới
	if err := s.tokenRepo.StoreRefreshToken(ctx, userID, newToken, newExpiresAt); err != nil {
		log.Printf("AuthService: Failed to store new refresh token for user %d: %v", userID, err)
		return status.Error(codes.Internal, "failed to store new refresh token")
	}

	log.Printf("AuthService: Successfully updated refresh token for user %d", userID)
	return nil
}

// =================================
// Token Storage & Invalidation Implementation
// =================================

// StoreRefreshToken stores a refresh token.
func (s *AuthService) StoreRefreshToken(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	log.Printf("AuthService: Storing refresh token for user %d", userID)

	if err := s.tokenRepo.StoreRefreshToken(ctx, userID, refreshToken, expiresAt); err != nil {
		log.Printf("AuthService: Failed to store refresh token: %v", err)
		return status.Error(codes.Internal, "failed to store refresh token")
	}
	return nil
}

// InvalidateUserTokens invalidates a user's specific access/refresh token pair (e.g., on logout).
func (s *AuthService) InvalidateUserTokens(ctx context.Context, accessToken string, refreshToken *string) error {
	log.Printf("AuthService: Invalidating tokens")

	// Blacklist the access token until it expires naturally.
	claims, err := utils.ValidateJWT(accessToken, s.jwtSecret)
	if err == nil {
		// CORRECTED LINE: Access .ExpiresAt.Time from the embedded RegisteredClaims
		if err_blacklist := s.tokenRepo.BlacklistToken(ctx, accessToken, claims.ExpiresAt.Time); err_blacklist != nil {
			log.Printf("AuthService: Failed to blacklist access token: %v", err_blacklist)
			// Non-critical, but should be monitored.
		}
	}

	// Delete the refresh token from storage.
	if refreshToken != nil && *refreshToken != "" {
		if err_delete := s.tokenRepo.DeleteRefreshToken(ctx, *refreshToken); err_delete != nil {
			log.Printf("AuthService: Failed to delete refresh token: %v", err_delete)
			return status.Error(codes.Internal, "failed to invalidate refresh token")
		}
	}
	return nil
}

// InvalidateAllUserTokens invalidates all refresh tokens for a user (e.g., on password change).
func (s *AuthService) InvalidateAllUserTokens(ctx context.Context, userID int64) error {
	log.Printf("AuthService: Invalidating all tokens for user %d", userID)
	if err := s.tokenRepo.DeleteAllUserRefreshTokens(ctx, userID); err != nil {
		log.Printf("AuthService: Failed to delete all refresh tokens for user %d: %v", userID, err)
		return status.Error(codes.Internal, "failed to invalidate all tokens")
	}
	return nil
}

// =================================
// Password Reset Implementation
// =================================

// GeneratePasswordResetToken generates a new password reset token.
func (s *AuthService) GeneratePasswordResetToken(ctx context.Context, userID int64) (string, time.Time, error) {
	log.Printf("AuthService: Generating password reset token for user %d", userID)

	// Reset tokens are typically secure random strings, not JWTs.
	token, err := utils.GenerateRefreshToken() // Re-using the same random string generator
	if err != nil {
		log.Printf("AuthService: Failed to generate reset token: %v", err)
		return "", time.Time{}, status.Error(codes.Internal, "failed to generate reset token")
	}

	expiresAt := time.Now().Add(s.resetTokenTTL)
	return token, expiresAt, nil
}

// StorePasswordResetToken stores a password reset token.
func (s *AuthService) StorePasswordResetToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	log.Printf("AuthService: Storing password reset token for user %d", userID)
	if err := s.tokenRepo.StorePasswordResetToken(ctx, userID, token, expiresAt); err != nil {
		log.Printf("AuthService: Failed to store reset token: %v", err)
		return status.Error(codes.Internal, "failed to store reset token")
	}
	return nil
}

// ValidatePasswordResetToken validates a password reset token.
func (s *AuthService) ValidatePasswordResetToken(ctx context.Context, token string) (*utils.PasswordResetTokenData, error) {
	log.Printf("AuthService: Validating password reset token")

	tokenData, err := s.tokenRepo.GetPasswordResetToken(ctx, token)
	if err != nil {
		log.Printf("AuthService: Password reset token validation failed: %v", err)
		return nil, status.Error(codes.NotFound, "invalid or expired reset token")
	}

	// Final check on expiration
	if time.Now().After(tokenData.ExpiresAt) {
		log.Printf("AuthService: Reset token is expired (logic check)")
		_ = s.tokenRepo.DeletePasswordResetToken(ctx, token)
		return nil, status.Error(codes.NotFound, "reset token has expired")
	}

	return tokenData, nil
}

// InvalidatePasswordResetToken deletes a password reset token after it's been used.
func (s *AuthService) InvalidatePasswordResetToken(ctx context.Context, token string) error {
	log.Printf("AuthService: Invalidating password reset token")
	if err := s.tokenRepo.DeletePasswordResetToken(ctx, token); err != nil {
		log.Printf("AuthService: Failed to invalidate reset token: %v", err)
		// This is often not a critical error to return to the user, but should be logged.
	}
	return nil
}
