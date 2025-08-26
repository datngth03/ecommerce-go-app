// internal/user/application/service.go
package application

import (
	"context"
	"errors"
	"time"

	// "fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	// Kiểm tra xem người dùng đã tồn tại hay chưa.
	_, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "User with this email already exists")
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, status.Errorf(codes.Internal, "Failed to check for existing user: %v", err)
	}

	// Băm mật khẩu trước khi tạo đối tượng domain.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to hash password")
	}

	// Tạo đối tượng domain người dùng.
	userID := uuid.New().String()
	user := domain.NewUser(
		userID,
		req.GetEmail(),
		string(hashedPassword),
		req.GetFirstName(),
		req.GetLastName(),
	)

	// Lưu người dùng vào repository.
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user: %v", err)
	}

	// TODO: Thay thế logger bằng một logger thích hợp.
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

	// So sánh mật khẩu đã nhập với mật khẩu đã băm trong database
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetPassword())); err != nil {
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
		// Dựa vào domain.ErrUserNotFound để trả về lỗi 404
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user profile: %v", err)
	}

	// Chuyển đổi slice domain.Address sang slice user_client.Address
	addresses := make([]*user_client.Address, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = &user_client.Address{
			Id:          addr.ID,
			Name:        addr.FullName,      // Ánh xạ FullName -> Name
			AddressLine: addr.StreetAddress, // Ánh xạ StreetAddress -> AddressLine
			City:        addr.City,
			Country:     addr.Country,
			ZipCode:     addr.PostalCode, // Ánh xạ PostalCode -> ZipCode
			IsDefault:   addr.IsDefault,
			// TODO: Add other fields as needed from the domain.Address
			CreatedAt: timestamppb.New(addr.CreatedAt),
			UpdatedAt: timestamppb.New(addr.UpdatedAt),
		}
	}

	return &user_client.UserProfileResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
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
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// Cập nhật các trường dựa trên yêu cầu
	if req.GetFirstName() != "" {
		user.FirstName = req.GetFirstName()
	}
	if req.GetLastName() != "" {
		user.LastName = req.GetLastName()
	}
	if req.GetPhoneNumber() != "" {
		user.PhoneNumber = req.GetPhoneNumber()
	}
	user.UpdatedAt = time.Now()

	// Lưu người dùng đã cập nhật vào repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update user profile: %v", err)
	}

	// Chuyển đổi slice domain.Address sang slice user_client.Address
	addresses := make([]*user_client.Address, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = &user_client.Address{
			Id:          addr.ID,
			Name:        addr.FullName,      // Ánh xạ FullName -> Name
			AddressLine: addr.StreetAddress, // Ánh xạ StreetAddress -> AddressLine
			City:        addr.City,
			Country:     addr.Country,
			ZipCode:     addr.PostalCode, // Ánh xạ PostalCode -> ZipCode
			IsDefault:   addr.IsDefault,
			CreatedAt:   timestamppb.New(addr.CreatedAt),
			UpdatedAt:   timestamppb.New(addr.UpdatedAt),
		}
	}

	return &user_client.UserProfileResponse{
		UserId:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
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

	// Chuyển đổi slice domain.Address sang slice user_client.Address
	addresses := make([]*user_client.Address, len(user.Addresses))
	for i, addr := range user.Addresses {
		addresses[i] = &user_client.Address{
			Id:          addr.ID,
			Name:        addr.FullName,
			AddressLine: addr.StreetAddress,
			City:        addr.City,
			Country:     addr.Country,
			ZipCode:     addr.PostalCode,
			IsDefault:   addr.IsDefault,
			CreatedAt:   timestamppb.New(addr.CreatedAt),
			UpdatedAt:   timestamppb.New(addr.UpdatedAt),
		}
	}

	return &user_client.GetUserByEmailResponse{
		User: &user_client.UserProfileResponse{
			UserId:      user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
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
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// Xác thực mật khẩu cũ bằng cách so sánh với mật khẩu đã băm trong database.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.GetOldPassword())); err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid old password")
	}

	// Băm mật khẩu mới.
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to hash new password")
	}

	// Cập nhật mật khẩu đã băm và UpdatedAt.
	user.PasswordHash = string(newPasswordHash)
	user.UpdatedAt = time.Now()

	// Lưu người dùng đã cập nhật vào repository.
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

	// TODO: A real password reset flow would involve generating a secure, time-limited token
	// and sending it to the user's email address. The user would then use this token
	// to set a new password. The logic below is a placeholder.
	_, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err != nil {
		// Dựa vào domain.ErrUserNotFound để trả về lỗi 404
		if errors.Is(err, domain.ErrUserNotFound) {
			// Trả về phản hồi thành công ngay cả khi email không được tìm thấy
			// để tránh tiết lộ thông tin người dùng.
			logger.Logger.Info("Simulating password reset request for non-existent ", zap.String("email", req.GetEmail()))
			return &user_client.ResetPasswordResponse{Message: "If the email is registered, a password reset link has been sent."}, nil
		}
		return nil, status.Errorf(codes.Internal, "Failed to find user by email: %v", err)
	}

	// Gửi email hoặc tạo mã token ở đây
	logger.Logger.Info("Simulating password reset request for ", zap.String("email", req.GetEmail()))

	return &user_client.ResetPasswordResponse{Message: "If the email is registered, a password reset link has been sent."}, nil
}

// AddAddress adds a new address for a user.
func (s *userService) AddAddress(ctx context.Context, req *user_client.AddAddressRequest) (*user_client.AddressResponse, error) {
	if req.GetUserId() == "" || req.GetAddress() == nil {
		return nil, status.Error(codes.InvalidArgument, "User ID and address are required")
	}
	if req.GetAddress().GetAddressLine() == "" {
		return nil, status.Error(codes.InvalidArgument, "Street address is required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// Tạo một đối tượng domain.Address từ yêu cầu
	newAddress := domain.NewAddress(
		uuid.New().String(),
		req.GetAddress().GetName(),
		req.GetAddress().GetAddressLine(),
		req.GetAddress().GetCity(),
		req.GetAddress().GetZipCode(),
		req.GetAddress().GetCountry(),
		req.GetAddress().GetIsDefault(),
	)

	// Thêm địa chỉ mới vào người dùng
	user.AddAddress(newAddress)

	// Lưu người dùng đã cập nhật vào repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user with new address: %v", err)
	}

	// Chuyển đổi địa chỉ đã thêm sang loại gRPC client để trả về
	return &user_client.AddressResponse{
		Address: &user_client.Address{
			Id:          newAddress.ID,
			Name:        newAddress.FullName,
			AddressLine: newAddress.StreetAddress,
			City:        newAddress.City,
			ZipCode:     newAddress.PostalCode,
			Country:     newAddress.Country,
			IsDefault:   newAddress.IsDefault,
			CreatedAt:   timestamppb.New(newAddress.CreatedAt),
			UpdatedAt:   timestamppb.New(newAddress.UpdatedAt),
		},
	}, nil
}

// UpdateAddress updates an existing user address.
func (s *userService) UpdateAddress(ctx context.Context, req *user_client.UpdateAddressRequest) (*user_client.AddressResponse, error) {
	if req.GetUserId() == "" || req.GetAddress() == nil || req.GetAddress().GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID, address ID, and address are required")
	}

	user, err := s.userRepo.FindByID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// Tạo một đối tượng domain.Address từ yêu cầu
	updatedAddress := domain.Address{
		ID:            req.GetAddress().GetId(),
		FullName:      req.GetAddress().GetName(),
		StreetAddress: req.GetAddress().GetAddressLine(),
		City:          req.GetAddress().GetCity(),
		PostalCode:    req.GetAddress().GetZipCode(),
		Country:       req.GetAddress().GetCountry(),
		IsDefault:     req.GetAddress().GetIsDefault(),
	}

	// Cập nhật địa chỉ trong user domain model
	if err := user.UpdateAddress(updatedAddress); err != nil {
		if errors.Is(err, domain.ErrAddressNotFound) {
			return nil, status.Error(codes.NotFound, "Address not found for this user")
		}
		return nil, status.Errorf(codes.Internal, "Failed to update address: %v", err)
	}

	// Lưu người dùng đã cập nhật vào repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save user with updated address: %v", err)
	}

	// Tìm địa chỉ đã cập nhật trong slice của người dùng để trả về
	var returnedAddress *user_client.Address
	for _, addr := range user.Addresses {
		if addr.ID == updatedAddress.ID {
			returnedAddress = &user_client.Address{
				Id:          addr.ID,
				Name:        addr.FullName,
				AddressLine: addr.StreetAddress,
				City:        addr.City,
				ZipCode:     addr.PostalCode,
				Country:     addr.Country,
				IsDefault:   addr.IsDefault,
				CreatedAt:   timestamppb.New(addr.CreatedAt),
				UpdatedAt:   timestamppb.New(addr.UpdatedAt),
			}
			break
		}
	}
	if returnedAddress == nil {
		// Đây là một lỗi hiếm gặp, nhưng nên xử lý để đảm bảo an toàn
		return nil, status.Error(codes.Internal, "Failed to find updated address after saving")
	}

	return &user_client.AddressResponse{Address: returnedAddress}, nil
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

	if err := user.RemoveAddress(req.GetAddressId()); err != nil {
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
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// Chuyển đổi slice domain.Address sang slice user_client.Address
	protoAddresses := make([]*user_client.Address, 0, len(user.Addresses))
	for _, addr := range user.Addresses {
		protoAddresses = append(protoAddresses, &user_client.Address{
			Id:          addr.ID,
			Name:        addr.FullName,
			AddressLine: addr.StreetAddress,
			City:        addr.City,
			Country:     addr.Country,
			ZipCode:     addr.PostalCode,
			IsDefault:   addr.IsDefault,
			CreatedAt:   timestamppb.New(addr.CreatedAt),
			UpdatedAt:   timestamppb.New(addr.UpdatedAt),
		})
	}

	return &user_client.ListAddressesResponse{
		Addresses: protoAddresses,
	}, nil
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
