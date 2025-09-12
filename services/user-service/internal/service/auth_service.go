package service

import (
	"context"
	"fmt"

	"github.com/ecommerce/services/user-service/internal/models"
	"github.com/ecommerce/services/user-service/pkg/utils"
)

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	RefreshToken(ctx context.Context, token string) (string, error)
	Logout(ctx context.Context, token string) error
}

type authService struct {
	userService UserService
}

func NewAuthService(userService UserService) AuthService {
	return &authService{
		userService: userService,
	}
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	return s.userService.ValidateToken(ctx, token)
}

func (s *authService) RefreshToken(ctx context.Context, token string) (string, error) {
	// Validate current token
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// Extract user information
	userID := int64(claims["user_id"].(float64))
	email := claims["email"].(string)
	role := claims["role"].(string)

	// Generate new token
	newToken, err := utils.GenerateToken(userID, email, role)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	// In a production system, you might want to:
	// 1. Add token to blacklist in Redis
	// 2. Log the logout event
	// For now, we'll just validate that the token is valid
	_, err := utils.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// TODO: Add token to blacklist if needed
	return nil
}
