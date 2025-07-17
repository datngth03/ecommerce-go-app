// internal/user/delivery/grpc/user_grpc_server.go
package grpc

import (
	"context"
	"log" // Tạm thời dùng log để in lỗi

	"github.com/datngth03/ecommerce-go-app/internal/user/application"
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user" // Import mã gRPC đã tạo
)

// UserGRPCServer implements the user_client.UserServiceServer interface.
// UserGRPCServer triển khai interface user_client.UserServiceServer.
type UserGRPCServer struct {
	user_client.UnimplementedUserServiceServer // Embedded to satisfy all methods, allows future additions
	userService                                application.UserService
}

// NewUserGRPCServer creates a new instance of UserGRPCServer.
// NewUserGRPCServer tạo một thể hiện mới của UserGRPCServer.
func NewUserGRPCServer(svc application.UserService) *UserGRPCServer {
	return &UserGRPCServer{
		userService: svc,
	}
}

// RegisterUser implements the gRPC RegisterUser method.
// RegisterUser triển khai phương thức gRPC RegisterUser.
func (s *UserGRPCServer) RegisterUser(ctx context.Context, req *user_client.RegisterUserRequest) (*user_client.RegisterUserResponse, error) {
	log.Printf("Nhận yêu cầu RegisterUser cho email: %s", req.GetEmail())
	resp, err := s.userService.RegisterUser(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi đăng ký người dùng: %v", err)
		return nil, err // gRPC sẽ tự động chuyển đổi lỗi Go sang mã lỗi gRPC
	}
	return resp, nil
}

// LoginUser implements the gRPC LoginUser method.
// LoginUser triển khai phương thức gRPC LoginUser.
func (s *UserGRPCServer) LoginUser(ctx context.Context, req *user_client.LoginUserRequest) (*user_client.LoginUserResponse, error) {
	log.Printf("Nhận yêu cầu LoginUser cho email: %s", req.GetEmail())
	resp, err := s.userService.LoginUser(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi đăng nhập người dùng: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetUserProfile implements the gRPC GetUserProfile method.
// GetUserProfile triển khai phương thức gRPC GetUserProfile.
func (s *UserGRPCServer) GetUserProfile(ctx context.Context, req *user_client.GetUserProfileRequest) (*user_client.UserProfileResponse, error) {
	log.Printf("Nhận yêu cầu GetUserProfile cho User ID: %s", req.GetUserId())
	resp, err := s.userService.GetUserProfile(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi lấy hồ sơ người dùng: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateUserProfile implements the gRPC UpdateUserProfile method.
// UpdateUserProfile triển khai phương thức gRPC UpdateUserProfile.
func (s *UserGRPCServer) UpdateUserProfile(ctx context.Context, req *user_client.UpdateUserProfileRequest) (*user_client.UserProfileResponse, error) {
	log.Printf("Nhận yêu cầu UpdateUserProfile cho User ID: %s", req.GetUserId())
	resp, err := s.userService.UpdateUserProfile(ctx, req)
	if err != nil {
		log.Printf("Lỗi khi cập nhật hồ sơ người dùng: %v", err)
		return nil, err
	}
	return resp, nil
}
