package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ecommerce/proto/user_service"
	"github.com/ecommerce/services/user-service/internal/models"
	"github.com/ecommerce/services/user-service/internal/service"
)

type UserServer struct {
	user_service.UnimplementedUserServiceServer
	userService service.UserService
	authService service.AuthService
}

func NewUserServer(userService service.UserService, authService service.AuthService) *UserServer {
	return &UserServer{
		userService: userService,
		authService: authService,
	}
}

func (s *UserServer) CreateUser(ctx context.Context, req *user_service.CreateUserRequest) (*user_service.UserResponse, error) {
	// Convert gRPC request to internal model
	createReq := &models.CreateUserRequest{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		Phone:    req.Phone,
		Role:     req.Role,
	}

	// Call service
	user, err := s.userService.CreateUser(ctx, createReq)
	if err != nil {
		return &user_service.UserResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert to gRPC response
	return &user_service.UserResponse{
		Success: true,
		Message: "User created successfully",
		User:    s.modelToProto(user),
	}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *user_service.GetUserRequest) (*user_service.UserResponse, error) {
	user, err := s.userService.GetUserByID(ctx, req.Id)
	if err != nil {
		code := codes.Internal
		if err.Error() == "user not found" {
			code = codes.NotFound
		}
		return &user_service.UserResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(code, err.Error())
	}

	return &user_service.UserResponse{
		Success: true,
		Message: "User retrieved successfully",
		User:    s.modelToProto(user),
	}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *user_service.UpdateUserRequest) (*user_service.UserResponse, error) {
	updateReq := &models.UpdateUserRequest{
		Name:  req.Name,
		Phone: req.Phone,
		Role:  req.Role,
	}

	user, err := s.userService.UpdateUser(ctx, req.Id, updateReq)
	if err != nil {
		code := codes.Internal
		if err.Error() == "user not found" {
			code = codes.NotFound
		}
		return &user_service.UserResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(code, err.Error())
	}

	return &user_service.UserResponse{
		Success: true,
		Message: "User updated successfully",
		User:    s.modelToProto(user),
	}, nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *user_service.DeleteUserRequest) (*user_service.DeleteUserResponse, error) {
	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		code := codes.Internal
		if err.Error() == "user not found" {
			code = codes.NotFound
		}
		return &user_service.DeleteUserResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(code, err.Error())
	}

	return &user_service.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

func (s *UserServer) ListUsers(ctx context.Context, req *user_service.ListUsersRequest) (*user_service.ListUsersResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	users, total, err := s.userService.ListUsers(ctx, page, limit, req.Role)
	if err != nil {
		return &user_service.ListUsersResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.Internal, err.Error())
	}

	// Convert users to proto
	protoUsers := make([]*user_service.User, len(users))
	for i, user := range users {
		protoUsers[i] = s.modelToProto(&user)
	}

	return &user_service.ListUsersResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Users:   protoUsers,
		Total:   int32(total),
	}, nil
}

func (s *UserServer) Login(ctx context.Context, req *user_service.LoginRequest) (*user_service.LoginResponse, error) {
	loginReq := &models.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	loginResp, err := s.userService.Login(ctx, loginReq)
	if err != nil {
		code := codes.Internal
		if err.Error() == "invalid credentials" {
			code = codes.Unauthenticated
		}
		return &user_service.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(code, err.Error())
	}

	return &user_service.LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   loginResp.Token,
		User:    s.modelToProto(loginResp.User),
	}, nil
}

func (s *UserServer) ValidateToken(ctx context.Context, req *user_service.ValidateTokenRequest) (*user_service.ValidateTokenResponse, error) {
	user, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &user_service.ValidateTokenResponse{
			Success: false,
			Message: err.Error(),
		}, status.Error(codes.Unauthenticated, err.Error())
	}

	return &user_service.ValidateTokenResponse{
		Success: true,
		Message: "Token is valid",
		User:    s.modelToProto(user),
	}, nil
}

// Helper function to convert internal model to proto
func (s *UserServer) modelToProto(user *models.User) *user_service.User {
	if user == nil {
		return nil
	}

	return &user_service.User{
		Id:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Phone:     user.Phone,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
