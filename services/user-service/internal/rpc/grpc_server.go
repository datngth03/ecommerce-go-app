// file: internal/rpc/server.go

package rpc

import (
	"context"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	// "github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"
)

// Server là struct tổng hợp, chứa tất cả các implementation của gRPC server.
type GRPCServer struct {
	pb.UnimplementedUserServiceServer // Nhúng để đảm bảo tương thích
	*UserServer                       // Nhúng UserServer
	*AuthServer                       // Nhúng AuthServer
}

// NewServer tạo một instance của server tổng hợp.
func NewGRPCServer(userService service.UserServiceInterface, authService service.AuthServiceInterface) *GRPCServer {
	return &GRPCServer{
		UserServer: NewUserServer(userService),
		AuthServer: NewAuthServer(userService, authService),
	}
}
func (s *GRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	// Gọi một cách tường minh đến phiên bản "thật" mà bạn đã viết trong UserServer
	return s.UserServer.CreateUser(ctx, req)
}
func (s *GRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {

	return s.UserServer.DeleteUser(ctx, req)
}
func (s *GRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	return s.UserServer.GetUser(ctx, req)
}
func (s *GRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	return s.UserServer.UpdateUser(ctx, req)
}

// func (s *GRPCServer) validateCreateUserRequest(req *pb.CreateUserRequest) error {
// 	return s.UserServer.validateCreateUserRequest(req)
// }
// func (s *GRPCServer) modelToProtoUser(user *models.User) *pb.User {
// 	return s.UserServer.modelToProtoUser(user)
// }

func (s *GRPCServer) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	// Gọi một cách tường minh đến phiên bản "thật" mà bạn đã viết trong UserServer
	return s.AuthServer.ChangePassword(ctx, req)
}

func (s *GRPCServer) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	return s.AuthServer.ForgotPassword(ctx, req)
}
func (s *GRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.AuthServer.Login(ctx, req)
}
func (s *GRPCServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	return s.AuthServer.ValidateToken(ctx, req)
}

func (s *GRPCServer) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	return s.AuthServer.ResetPassword(ctx, req)
}
func (s *GRPCServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return s.AuthServer.Logout(ctx, req)
}
func (s *GRPCServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	return s.AuthServer.RefreshToken(ctx, req)
}
