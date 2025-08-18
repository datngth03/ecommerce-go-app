// internal/user/application/service.go
package application

import (
	"context"
	"errors"

	// "fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"
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
	GetUserByEmail(ctx context.Context, req *user_client.GetUserByEmailRequest) (*user_client.GetUserByEmailResponse, error)
	ChangePassword(ctx context.Context, req *user_client.ChangePasswordRequest) (*user_client.ChangePasswordResponse, error)
	ResetPassword(ctx context.Context, req *user_client.ResetPasswordRequest) (*user_client.ResetPasswordResponse, error)
	AddAddress(ctx context.Context, req *user_client.AddAddressRequest) (*user_client.AddressResponse, error)
	UpdateAddress(ctx context.Context, req *user_client.UpdateAddressRequest) (*user_client.AddressResponse, error)
	DeleteAddress(ctx context.Context, req *user_client.DeleteAddressRequest) (*user_client.DeleteAddressResponse, error)
	ListAddresses(ctx context.Context, req *user_client.ListAddressesRequest) (*user_client.ListAddressesResponse, error)
	GetOrderHistory(ctx context.Context, req *user_client.GetOrderHistoryRequest) (*user_client.GetOrderHistoryResponse, error)
}

// userService implements the UserService interface.
// userService triển khai interface UserService.
type userService struct {
	userRepo domain.UserRepository
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
		return nil, status.Error(codes.InvalidArgument, "Email, password, and full name are required")
	}

	// Check if user already exists
	_, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "User with this email already exists")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, status.Errorf(codes.Internal, "Failed to check for existing user: %v", err)
	}

	// Create new user domain entity
	userID := uuid.New().String()
	user := domain.NewUser(userID, req.GetEmail(), req.GetPassword(), req.GetFullName())

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, status.Error(codes.Internal, "Failed to hash password")
	}

	// Save user to repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user: %v", err)
	}

	logger.Logger.Info("User registered successfully", zap.String("userID", user.ID))

	return &user_client.RegisterUserResponse{
		UserId:  user.ID,
		Message: "User registered successfully",
	}, nil
}

// LoginUser handles the user login use case.
// LoginUser xử lý trường hợp sử dụng đăng nhập người dùng.
func (s *userService) LoginUser(ctx context.Context, req *user_client.LoginUserRequest) (*user_client.LoginUserResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
	}

	if !user.CheckPassword(req.GetPassword()) { // Placeholder for password check
		return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
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
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	addresses := make([]*user_client.Address, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = &user_client.Address{
			StreetAddress: addr.StreetAddress,
			City:          addr.City,
			Country:       addr.Country,
			PostalCode:    addr.PostalCode,
			IsDefault:     addr.IsDefault,
		}
	}

	return &user_client.UserProfileResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Addresses:   addresses,
	}, nil
}

// UpdateUserProfile handles updating a user's profile.
// UpdateUserProfile xử lý việc cập nhật hồ sơ người dùng.
func (s *userService) UpdateUserProfile(ctx context.Context, req *user_client.UpdateUserProfileRequest) (*user_client.UserProfileResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	user.UpdateProfile(req.GetFullName(), req.GetPhoneNumber(), req.GetAddress())

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update user profile: %v", err)
	}

	addresses := make([]*user_client.Address, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = &user_client.Address{
			StreetAddress: addr.StreetAddress,
			City:          addr.City,
			Country:       addr.Country,
			PostalCode:    addr.PostalCode,
			IsDefault:     addr.IsDefault,
		}
	}

	return &user_client.UserProfileResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Addresses:   addresses,
	}, nil
}

// GetUserByEmail retrieves a user by their email address.
func (s *userService) GetUserByEmail(ctx context.Context, req *user_client.GetUserByEmailRequest) (*user_client.GetUserByEmailResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	addresses := make([]*user_client.Address, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = &user_client.Address{
			StreetAddress: addr.StreetAddress,
			City:          addr.City,
			Country:       addr.Country,
			PostalCode:    addr.PostalCode,
			IsDefault:     addr.IsDefault,
		}
	}

	return &user_client.GetUserByEmailResponse{
		User: &user_client.UserProfileResponse{
			UserId:      user.ID,
			Email:       user.Email,
			FullName:    user.FullName,
			PhoneNumber: user.PhoneNumber,
			Addresses:   addresses,
		},
	}, nil
}

// ChangePassword handles a user changing their password.
func (s *userService) ChangePassword(ctx context.Context, req *user_client.ChangePasswordRequest) (*user_client.ChangePasswordResponse, error) {
	if req.GetUserId() == "" || req.GetOldPassword() == "" || req.GetNewPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID, old password, and new password are required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	if !user.CheckPassword(req.GetOldPassword()) {
		return nil, status.Error(codes.Unauthenticated, "Invalid old password")
	}

	user.UpdatePassword(req.GetNewPassword())
	if err := user.HashPassword(); err != nil {
		return nil, status.Error(codes.Internal, "Failed to hash new password")
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to change password: %v", err)
	}

	return &user_client.ChangePasswordResponse{Message: "Password changed successfully"}, nil
}

// ResetPassword handles the password reset flow.
func (s *userService) ResetPassword(ctx context.Context, req *user_client.ResetPasswordRequest) (*user_client.ResetPasswordResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	// TODO: Logic for a real password reset flow would involve sending a token to the user's email.
	// We'll simulate a successful response for now.
	logger.Logger.Info("Simulating password reset request", zap.String("email", req.GetEmail()))

	return &user_client.ResetPasswordResponse{Message: "If the email is registered, a password reset link has been sent."}, nil
}

// AddAddress adds a new address for a user.
func (s *userService) AddAddress(ctx context.Context, req *user_client.AddAddressRequest) (*user_client.AddressResponse, error) {
	if req.GetUserId() == "" || req.GetAddress() == nil {
		return nil, status.Error(codes.InvalidArgument, "User ID and address are required")
	}
	if req.GetAddress().GetStreetAddress() == "" {
		return nil, status.Error(codes.InvalidArgument, "Street address is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// Placeholder logic to add address to the user's domain model
	// Assuming the user domain model has a method like AddAddress
	newAddressID := uuid.New().String()
	req.GetAddress().Id = newAddressID

	// Simulate adding the address to the user object
	// This would require a `AddAddress` method on the `domain.User` struct
	// For this example, we'll just return the provided address with an ID

	addr := req.GetAddress()

	if err := user.AddAddress(domain.Address{
		ID:            addr.Id,
		FullName:      addr.FullName,
		PhoneNumber:   addr.PhoneNumber,
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		PostalCode:    addr.PostalCode,
		Country:       addr.Country,
		IsDefault:     addr.IsDefault,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to add address: %v", err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user with new address: %v", err)
	}

	return &user_client.AddressResponse{Address: req.GetAddress()}, nil
}

// UpdateAddress updates an existing user address.
func (s *userService) UpdateAddress(ctx context.Context, req *user_client.UpdateAddressRequest) (*user_client.AddressResponse, error) {
	if req.GetUserId() == "" || req.GetAddress() == nil || req.GetAddress().GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID, address ID, and address are required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// Placeholder logic to update the address
	// This would require a `UpdateAddress` method on the `domain.User` struct

	addr := req.GetAddress()
	if err := user.UpdateAddress(domain.Address{
		ID:            addr.Id,
		FullName:      addr.FullName,
		PhoneNumber:   addr.PhoneNumber,
		StreetAddress: addr.StreetAddress,
		City:          addr.City,
		PostalCode:    addr.PostalCode,
		Country:       addr.Country,
		IsDefault:     addr.IsDefault,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update address: %v", err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user with updated address: %v", err)
	}

	return &user_client.AddressResponse{Address: req.GetAddress()}, nil
}

// DeleteAddress removes a user address.
func (s *userService) DeleteAddress(ctx context.Context, req *user_client.DeleteAddressRequest) (*user_client.DeleteAddressResponse, error) {
	if req.GetUserId() == "" || req.GetAddressId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID and address ID are required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// Placeholder logic to delete the address
	// This would require a `DeleteAddress` method on the `domain.User` struct

	if err := user.DeleteAddress(req.GetAddressId()); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete address: %v", err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user after deleting address: %v", err)
	}

	return &user_client.DeleteAddressResponse{Message: "Address deleted successfully"}, nil
}

// ListAddresses retrieves all addresses for a user.
func (s *userService) ListAddresses(ctx context.Context, req *user_client.ListAddressesRequest) (*user_client.ListAddressesResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	// Placeholder logic to get addresses
	// Assuming the user domain model has a `GetAddresses` method that returns a slice of addresses
	// For this example, we'll return an empty list

	addresses := user.GetAddress()
	protoAddresses := []*user_client.Address{}
	for _, addr := range addresses {
		protoAddresses = append(protoAddresses, &user_client.Address{
			Id: addr.ID,
			// Map other fields
		})
	}

	return &user_client.ListAddressesResponse{Addresses: []*user_client.Address{}}, nil
}

// GetOrderHistory retrieves the order history for a user.
func (s *userService) GetOrderHistory(ctx context.Context, req *user_client.GetOrderHistoryRequest) (*user_client.GetOrderHistoryResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// TODO: This would typically call an OrderRepository or another service
	// to fetch the user's order history.

	// For this example, we'll return a dummy order history.
	orders := []*user_client.Order{
		{
			OrderId:     "order-123",
			OrderDate:   1672531200, // Jan 1, 2023
			Status:      "Shipped",
			TotalAmount: 99.99,
		},
		{
			OrderId:     "order-456",
			OrderDate:   1677619200, // Mar 1, 2023
			Status:      "Delivered",
			TotalAmount: 250.50,
		},
	}

	return &user_client.GetOrderHistoryResponse{Orders: orders}, nil
}
