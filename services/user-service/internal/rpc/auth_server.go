// internal/rpc/auth_server.go
package rpc

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"
	// "github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
)

// AuthServer implements the authentication-related RPC methods
type AuthServer struct {
	pb.UnimplementedUserServiceServer
	userService service.UserServiceInterface
	authService service.AuthServiceInterface
}

// NewAuthServer creates a new AuthServer instance
func NewAuthServer(userService service.UserServiceInterface, authService service.AuthServiceInterface) *AuthServer {
	return &AuthServer{
		userService: userService,
		authService: authService,
	}
}

// =================================
// Authentication & Session Methods
// =================================

// Login authenticates user and returns tokens
func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Login RPC called for email: %s", req.Email)

	// Validate request
	if req.Email == "" || req.Password == "" {
		return &pb.LoginResponse{
			Success: false,
			Message: "Email and password are required",
		}, nil
	}

	// Validate credentials
	user, err := s.userService.ValidateUserCredentials(ctx, req.Email, req.Password)
	if err != nil {
		log.Printf("Login failed for email %s: %v", req.Email, err)
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Generate tokens
	tokenPair, err := s.authService.GenerateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to generate tokens for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "Failed to generate authentication tokens")
	}

	// Store refresh token
	err = s.authService.StoreRefreshToken(ctx, user.ID, tokenPair.RefreshToken, tokenPair.RefreshExpiresAt)
	if err != nil {
		log.Printf("Failed to store refresh token for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "Failed to complete login process")
	}

	// Convert user model to proto
	pbUser := &pb.User{
		Id:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
		IsActive: user.IsActive,
	}

	if !user.CreatedAt.IsZero() {
		pbUser.CreatedAt = timestamppb.New(user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		pbUser.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}

	log.Printf("Login successful for user %d", user.ID)
	return &pb.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         pbUser,
		ExpiresAt:    timestamppb.New(tokenPair.AccessExpiresAt),
	}, nil
}

// ValidateToken validates an access token
func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	log.Printf("ValidateToken RPC called")

	if req.Token == "" {
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "Token is required",
		}, nil
	}

	// Validate and parse token
	claims, err := s.authService.ValidateAccessToken(ctx, req.Token)
	if err != nil {
		log.Printf("Token validation failed: %v", err)
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid or expired token",
		}, nil
	}

	log.Printf("Token validated successfully for user %d", claims.UserID)
	return &pb.ValidateTokenResponse{
		Valid:     true,
		Message:   "Token is valid",
		UserId:    claims.UserID,
		Email:     claims.Email,
		ExpiresAt: timestamppb.New(claims.ExpiresAt.Time),
	}, nil
}

// RefreshToken refreshes an access token using refresh token
func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	log.Printf("RefreshToken RPC called")

	if req.RefreshToken == "" {
		return &pb.LoginResponse{
			Success: false,
			Message: "Refresh token is required",
		}, nil
	}

	// Validate refresh token
	tokenData, err := s.authService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("Refresh token validation failed: %v", err)
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid or expired refresh token",
		}, nil
	}

	// Get user info
	user, err := s.userService.GetUserByID(ctx, tokenData.UserID)
	if err != nil {
		log.Printf("User not found during token refresh: %d", tokenData.UserID)
		return &pb.LoginResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	// Generate new token pair
	newTokenPair, err := s.authService.GenerateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		log.Printf("Failed to generate new tokens for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "Failed to refresh tokens")
	}

	// Update refresh token in storage
	err = s.authService.UpdateRefreshToken(ctx, tokenData.UserID, req.RefreshToken, newTokenPair.RefreshToken, newTokenPair.RefreshExpiresAt)
	if err != nil {
		log.Printf("Failed to update refresh token for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "Failed to complete token refresh")
	}

	// Convert user model to proto
	pbUser := &pb.User{
		Id:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
		IsActive: user.IsActive,
	}

	if !user.CreatedAt.IsZero() {
		pbUser.CreatedAt = timestamppb.New(user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		pbUser.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}

	log.Printf("Token refreshed successfully for user %d", user.ID)
	return &pb.LoginResponse{
		Success:      true,
		Message:      "Tokens refreshed successfully",
		AccessToken:  newTokenPair.AccessToken,
		RefreshToken: newTokenPair.RefreshToken,
		User:         pbUser,
		ExpiresAt:    timestamppb.New(newTokenPair.AccessExpiresAt),
	}, nil
}

// Logout invalidates user tokens
func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	log.Printf("Logout RPC called")

	if req.AccessToken == "" {
		return &pb.LogoutResponse{
			Success: false,
			Message: "Access token is required",
		}, nil
	}

	// Extract user ID from access token
	claims, err := s.authService.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		// Even if token is invalid/expired, we still try to logout
		log.Printf("Access token validation failed during logout: %v", err)
	}

	var userID int64
	if claims != nil {
		userID = claims.UserID
	}

	// Invalidate tokens
	err = s.authService.InvalidateUserTokens(ctx, req.AccessToken, req.RefreshToken)
	if err != nil {
		log.Printf("Failed to invalidate tokens: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to complete logout process")
	}

	log.Printf("Logout successful for user %d", userID)
	return &pb.LogoutResponse{
		Success: true,
		Message: "Logout successful",
	}, nil
}

// =================================
// Password Management Methods
// =================================

// ChangePassword changes user password
func (s *AuthServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	log.Printf("ChangePassword RPC called")

	// Validate request
	if req.OldPassword == "" || req.NewPassword == "" {
		return &pb.ChangePasswordResponse{
			Success: false,
			Message: "Old password and new password are required",
		}, nil
	}

	// Extract user ID from authentication context (from JWT token in metadata)
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		log.Printf("Failed to get user ID from context: %v", err)
		return &pb.ChangePasswordResponse{
			Success: false,
			Message: "Authentication required",
		}, nil
	}

	// Change password
	err = s.userService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Printf("Failed to change password for user %d: %v", userID, err)

		if err.Error() == "invalid old password" {
			return &pb.ChangePasswordResponse{
				Success: false,
				Message: "Current password is incorrect",
			}, nil
		}

		return &pb.ChangePasswordResponse{
			Success: false,
			Message: "Failed to change password",
		}, nil
	}

	// Optionally invalidate all existing tokens to force re-login
	err = s.authService.InvalidateAllUserTokens(ctx, userID)
	if err != nil {
		log.Printf("Warning: Failed to invalidate tokens after password change for user %d: %v", userID, err)
	}

	log.Printf("Password changed successfully for user %d", userID)
	return &pb.ChangePasswordResponse{
		Success: true,
		Message: "Password changed successfully. Please login again.",
	}, nil
}

// ForgotPassword initiates password reset process
func (s *AuthServer) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	log.Printf("ForgotPassword RPC called for email: %s", req.Email)

	if req.Email == "" {
		return &pb.ForgotPasswordResponse{
			Success: false,
			Message: "Email is required",
		}, nil
	}

	// Check if user exists
	user, err := s.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// For security reasons, don't reveal if email exists or not
		log.Printf("User not found for password reset: %s", req.Email)
		return &pb.ForgotPasswordResponse{
			Success: true,
			Message: "If the email exists, password reset instructions have been sent",
		}, nil
	}

	// Generate reset token and expiry
	resetToken, expiresAt, err := s.authService.GeneratePasswordResetToken(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to generate reset token for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "Failed to process password reset request")
	}

	// Store reset token
	err = s.authService.StorePasswordResetToken(ctx, user.ID, resetToken, expiresAt)
	if err != nil {
		log.Printf("Failed to store reset token for user %d: %v", user.ID, err)
		return nil, status.Errorf(codes.Internal, "Failed to process password reset request")
	}

	// TODO: Send email with reset token (integrate with email service)
	// For now, we just log it (remove in production)
	log.Printf("Password reset token for %s: %s (expires: %v)", req.Email, resetToken, expiresAt)

	return &pb.ForgotPasswordResponse{
		Success:             true,
		Message:             "Password reset instructions have been sent to your email",
		ResetTokenExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// ResetPassword resets password using reset token
func (s *AuthServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	log.Printf("ResetPassword RPC called for email: %s", req.Email)

	// Validate request
	if req.Email == "" || req.ResetToken == "" || req.NewPassword == "" {
		return &pb.ResetPasswordResponse{
			Success: false,
			Message: "Email, reset token, and new password are required",
		}, nil
	}

	// Validate reset token
	res, err := s.authService.ValidatePasswordResetToken(ctx, req.ResetToken)
	if err != nil {
		log.Printf("Invalid reset token for email %s: %v", req.Email, err)
		return &pb.ResetPasswordResponse{
			Success: false,
			Message: "Invalid or expired reset token",
		}, nil
	}

	// Update password
	err = s.userService.UpdatePasswordByEmail(ctx, req.Email, req.NewPassword)
	if err != nil {
		log.Printf("Failed to update password for email %s: %v", req.Email, err)
		return nil, status.Errorf(codes.Internal, "Failed to reset password")
	}

	// Invalidate the reset token and all user tokens
	err = s.authService.InvalidatePasswordResetToken(ctx, req.ResetToken)
	if err != nil {
		log.Printf("Warning: Failed to invalidate reset token for user %d: %v", res, err)
	}

	err = s.authService.InvalidateAllUserTokens(ctx, res.UserID)
	if err != nil {
		log.Printf("Warning: Failed to invalidate all tokens for user %d: %v", res, err)
	}

	log.Printf("Password reset successful for user %d", res)
	return &pb.ResetPasswordResponse{
		Success: true,
		Message: "Password reset successful. Please login with your new password.",
	}, nil
}

// =================================
// Helper Methods
// =================================

// getUserIDFromContext extracts user ID from gRPC metadata (JWT token)
func (s *AuthServer) getUserIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	// Get authorization header
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return 0, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	// Extract token (assuming "Bearer <token>" format)
	token := authHeader[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validate token and extract claims
	claims, err := s.authService.ValidateAccessToken(ctx, token)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return claims.UserID, nil
}
