package handler

import (
	"context"
	"errors"
	"log"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implement gRPC server cho UserService.
type Server struct {
	pb.UnimplementedUserServiceServer
	userService service.UserServiceInterface
	authService service.AuthServiceInterface
}

// NewGRPCServer tạo một instance mới của gRPC server handler.
func NewGRPCServer(userService service.UserServiceInterface, authService service.AuthServiceInterface) *Server {
	return &Server{
		userService: userService,
		authService: authService,
	}
}

// --- User CRUD RPCs ---

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	log.Printf("Handler: CreateUser for %s", req.Email)

	userModel := &models.User{
		Email: req.Email,
		// Password sẽ được hash trong service
		Name:  req.Name,
		Phone: req.Phone,
	}

	createdUser, err := s.userService.CreateUser(ctx, userModel)
	if err != nil {
		log.Printf("Handler: Error CreateUser: %v", err)
		return nil, status.Errorf(codes.Internal, "could not create user: %v", err)
	}

	return &pb.UserResponse{User: convertToPbUser(createdUser)}, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	var user *models.User
	var err error

	// Use a switch to check which field in the 'oneof' is populated
	switch identifier := req.GetIdentifier().(type) {
	case *pb.GetUserRequest_Id:
		log.Printf("Handler: GetUser for ID %d", identifier.Id)
		if identifier.Id <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "User ID must be positive")
		}
		user, err = s.userService.GetUserByID(ctx, identifier.Id)

	case *pb.GetUserRequest_Email:
		log.Printf("Handler: GetUser for Email %s", identifier.Email)
		if identifier.Email == "" {
			return nil, status.Errorf(codes.InvalidArgument, "Email cannot be empty")
		}
		user, err = s.userService.GetUserByEmail(ctx, identifier.Email)

	default:
		return nil, status.Errorf(codes.InvalidArgument, "An identifier (ID or Email) is required")
	}

	// Common error handling after the switch
	if err != nil {
		log.Printf("gRPC Handler: Error getting user: %v", err)
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	return &pb.UserResponse{User: convertToPbUser(user)}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	log.Printf("Handler: UpdateUser for ID %d", req.Id)

	updateData := &models.UserUpdateData{
		ID:    req.Id,
		Name:  req.Name,
		Phone: req.Phone,
	}

	updatedUser, err := s.userService.UpdateUser(ctx, updateData)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found to update")
		}
		return nil, status.Errorf(codes.Internal, "could not update user: %v", err)
	}

	return &pb.UserResponse{User: convertToPbUser(updatedUser)}, nil
}

func (s *Server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Printf("Handler: DeleteUser for ID %d", req.Id)
	if err := s.userService.DeleteUser(ctx, req.Id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found to delete")
		}
		return nil, status.Errorf(codes.Internal, "could not delete user: %v", err)
	}
	return &pb.DeleteUserResponse{Message: "User deleted successfully"}, nil
}

// --- Auth & Session RPCs ---

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Handler: Login attempt for %s", req.Email)

	// Step 1: Validate credentials
	user, err := s.userService.ValidateUserCredentials(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	// Step 2: Generate tokens
	tokenPair, err := s.authService.GenerateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not generate tokens: %v", err)
	}

	// Step 3: Store refresh token
	err = s.authService.StoreRefreshToken(ctx, user.ID, tokenPair.RefreshToken, tokenPair.RefreshExpiresAt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not store refresh token: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         convertToPbUser(user),
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	log.Printf("Handler: RefreshToken request")

	// Step 1: Validate the refresh token
	tokenData, err := s.authService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token: %v", err)
	}

	// Step 2: Get user details
	user, err := s.userService.GetUserByID(ctx, tokenData.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not retrieve user for token")
	}

	// Step 3: Generate a new token pair
	newPair, err := s.authService.GenerateTokenPair(ctx, user.ID, user.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not generate new tokens")
	}

	// Step 4: Update the refresh token in storage (delete old, store new)
	err = s.authService.UpdateRefreshToken(ctx, user.ID, req.RefreshToken, newPair.RefreshToken, newPair.RefreshExpiresAt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update refresh token")
	}

	return &pb.LoginResponse{
		AccessToken:  newPair.AccessToken,
		RefreshToken: newPair.RefreshToken,
		User:         convertToPbUser(user),
	}, nil
}

// ... Các RPC khác như ValidateToken, Logout giữ nguyên logic tương tự ...

// --- Password Management RPCs ---

func (s *Server) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	// Step 1: Securely get user claims from the context.
	claims, ok := GetUserFromContext(ctx) // Assuming this helper function exists
	if !ok || claims == nil {
		return nil, status.Errorf(codes.Unauthenticated, "missing authentication details")
	}

	userID := claims.UserID // Use the ID from the trusted context.
	log.Printf("Handler: ChangePassword request for user ID %d", userID)

	// Step 2: Call the service with the trusted userID.
	err := s.userService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		// Service layer should return specific errors for "wrong password" vs "other error"
		return nil, status.Errorf(codes.InvalidArgument, "could not change password: %v", err)
	}

	// Step 3 (Best practice): Invalidate all other sessions on password change.
	_ = s.authService.InvalidateAllUserTokens(ctx, userID)

	return &pb.ChangePasswordResponse{Message: "Password changed successfully"}, nil
}

// You should have this helper function available from your middleware or in this handler file.
// It retrieves the claims that the authentication interceptor/middleware added.

func GetUserFromContext(ctx context.Context) (*utils.JWTClaims, bool) {
	value := ctx.Value("claims") // Make sure the key matches what your interceptor uses
	if value == nil {
		return nil, false
	}

	claims, ok := value.(*utils.JWTClaims)
	return claims, ok
}

func (s *Server) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	log.Printf("Handler: ForgotPassword request for email %s", req.Email)

	// Step 1: Check if user exists
	user, err := s.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Do not reveal if the email exists or not for security reasons.
		log.Printf("Handler: Attempted ForgotPassword for non-existent email %s", req.Email)
		return &pb.ForgotPasswordResponse{Message: "If a user with that email exists, a password reset link has been sent."}, nil
	}

	// Step 2: Generate and store reset token
	token, expiresAt, err := s.authService.GeneratePasswordResetToken(ctx, user.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not generate reset token")
	}
	if err := s.authService.StorePasswordResetToken(ctx, user.ID, token, expiresAt); err != nil {
		return nil, status.Errorf(codes.Internal, "could not store reset token")
	}

	// Step 3: (In a real app) Send the token to the user's email
	log.Printf("Handler: Generated password reset token for %s. Token: %s", req.Email, token)

	return &pb.ForgotPasswordResponse{Message: "If a user with that email exists, a password reset link has been sent."}, nil
}

func (s *Server) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	log.Printf("Handler: ResetPassword request with token %s", req.ResetToken)

	// Step 1: Validate reset token
	tokenData, err := s.authService.ValidatePasswordResetToken(ctx, req.ResetToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid or expired reset token")
	}

	// Step 2: Get user from token data
	user, err := s.userService.GetUserByID(ctx, tokenData.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not find user for valid token")
	}

	// Step 3: Update user's password
	err = s.userService.UpdatePasswordByEmail(ctx, user.Email, req.NewPassword)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not update password")
	}

	// Step 4: Invalidate the reset token so it can't be used again
	_ = s.authService.InvalidatePasswordResetToken(ctx, req.ResetToken)

	// Step 5: Invalidate all active sessions for security
	_ = s.authService.InvalidateAllUserTokens(ctx, user.ID)

	return &pb.ResetPasswordResponse{Message: "Password has been reset successfully"}, nil
}

// --- Helper Functions ---
func convertToPbUser(user *models.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		IsActive:  user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
