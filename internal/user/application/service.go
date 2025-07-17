// internal/user/application/service.go
package application

import (
	"context"
	"errors"

	"github.com/datngth03/ecommerce-go-app/internal/user/domain"
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user" // Import mã gRPC đã tạo
	"github.com/google/uuid"                                            // Để tạo UUID
)

// UserService defines the application service interface for user-related operations.
// UserService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến người dùng.
type UserService interface {
	RegisterUser(ctx context.Context, req *user_client.RegisterUserRequest) (*user_client.RegisterUserResponse, error)
	LoginUser(ctx context.Context, req *user_client.LoginUserRequest) (*user_client.LoginUserResponse, error)
	GetUserProfile(ctx context.Context, req *user_client.GetUserProfileRequest) (*user_client.UserProfileResponse, error)
	UpdateUserProfile(ctx context.Context, req *user_client.UpdateUserProfileRequest) (*user_client.UserProfileResponse, error)
}

// userService implements the UserService interface.
// userService triển khai interface UserService.
type userService struct {
	userRepo domain.UserRepository
	// TODO: Add other dependencies like token generator, event publisher
	// Thêm các dependency khác như trình tạo token, trình phát sự kiện
}

// NewUserService creates a new instance of UserService.
// NewUserService tạo một thể hiện mới của UserService.
func NewUserService(repo domain.UserRepository) UserService {
	return &userService{
		userRepo: repo,
	}
}

// RegisterUser handles the user registration use case.
// RegisterUser xử lý trường hợp sử dụng đăng ký người dùng.
func (s *userService) RegisterUser(ctx context.Context, req *user_client.RegisterUserRequest) (*user_client.RegisterUserResponse, error) {
	// Basic validation
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetFullName() == "" {
		return nil, errors.New("email, password, and full name are required")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err != nil && err.Error() != "user not found" { // Assuming "user not found" is a specific error from repo
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create new user domain entity
	userID := uuid.New().String()
	user := domain.NewUser(userID, req.GetEmail(), req.GetPassword(), req.GetFullName())

	// Hash password (placeholder)
	if err := user.HashPassword(); err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Save user to repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, errors.New("failed to save user")
	}

	// TODO: Publish UserRegistered event to message queue

	return &user_client.RegisterUserResponse{
		UserId:  user.ID,
		Message: "User registered successfully",
	}, nil
}

// LoginUser handles the user login use case.
// LoginUser xử lý trường hợp sử dụng đăng nhập người dùng.
func (s *userService) LoginUser(ctx context.Context, req *user_client.LoginUserRequest) (*user_client.LoginUserResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, errors.New("invalid credentials") // Avoid revealing if user exists or not
	}

	if !user.CheckPassword(req.GetPassword()) { // Placeholder for password check
		return nil, errors.New("invalid credentials")
	}

	// TODO: Generate JWT access and refresh tokens
	accessToken := "dummy_access_token_" + user.ID
	refreshToken := "dummy_refresh_token_" + user.ID
	expiresIn := int64(3600) // 1 hour

	return &user_client.LoginUserResponse{
		UserId:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// GetUserProfile handles retrieving a user's profile.
// GetUserProfile xử lý việc lấy hồ sơ người dùng.
func (s *userService) GetUserProfile(ctx context.Context, req *user_client.GetUserProfileRequest) (*user_client.UserProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &user_client.UserProfileResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
	}, nil
}

// UpdateUserProfile handles updating a user's profile.
// UpdateUserProfile xử lý việc cập nhật hồ sơ người dùng.
func (s *userService) UpdateUserProfile(ctx context.Context, req *user_client.UpdateUserProfileRequest) (*user_client.UserProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, errors.New("user not found")
	}

	user.UpdateProfile(req.GetFullName(), req.GetPhoneNumber(), req.GetAddress())

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, errors.New("failed to update user profile")
	}

	return &user_client.UserProfileResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Address:     user.Address,
	}, nil
}
