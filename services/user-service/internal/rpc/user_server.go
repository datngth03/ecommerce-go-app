// internal/rpc/user_server.go
package rpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"
)

// UserServer implements the UserService gRPC server
type UserServer struct {
	pb.UnimplementedUserServiceServer
	userService service.UserServiceInterface
}

// NewUserServer creates a new UserServer instance
func NewUserServer(userService service.UserServiceInterface) *UserServer {
	return &UserServer{
		userService: userService,
	}
}

// =================================
// CRUD Operations Implementation
// =================================

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		metrics.RecordGRPCRequest("CreateUser", statusCode, time.Since(start))
	}()

	log.Printf("CreateUser RPC called with email: %s", req.Email)

	// Validate request
	if err := s.validateCreateUserRequest(req); err != nil {
		statusCode = "validation_error"
		return &pb.UserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Convert proto request to domain model
	user := &models.User{
		Email:    req.Email,
		Name:     req.Name,
		Phone:    req.Phone,
		Password: req.Password, // Will be hashed in service layer
	}

	// Call service layer
	createdUser, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		log.Printf("CreateUser service error: %v", err)
		statusCode = "error"

		// Handle business logic errors
		if err.Error() == "user already exists" {
			return &pb.UserResponse{
				Success: false,
				Message: "User with this email already exists",
			}, nil
		}

		return nil, status.Errorf(codes.Internal, "Failed to create user: %v", err)
	}

	statusCode = "success"
	// Convert domain model to proto response
	pbUser := s.modelToProtoUser(createdUser)

	return &pb.UserResponse{
		Success: true,
		Message: "User created successfully",
		User:    pbUser,
	}, nil
}

// GetUser retrieves a user by ID or email
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	log.Printf("GetUser RPC called")

	var user *models.User
	var err error

	// Determine query type and call appropriate service method
	switch identifier := req.Identifier.(type) {
	case *pb.GetUserRequest_Id:
		log.Printf("Getting user by ID: %d", identifier.Id)
		user, err = s.userService.GetUserByID(ctx, identifier.Id)

	case *pb.GetUserRequest_Email:
		log.Printf("Getting user by email: %s", identifier.Email)
		user, err = s.userService.GetUserByEmail(ctx, identifier.Email)

	default:
		return &pb.UserResponse{
			Success: false,
			Message: "Either ID or email must be provided",
		}, nil
	}

	if err != nil {
		log.Printf("GetUser service error: %v", err)

		if err.Error() == "user not found" {
			return &pb.UserResponse{
				Success: false,
				Message: "User not found",
			}, nil
		}

		return nil, status.Errorf(codes.Internal, "Failed to get user: %v", err)
	}

	// Convert domain model to proto response
	pbUser := s.modelToProtoUser(user)

	return &pb.UserResponse{
		Success: true,
		Message: "User retrieved successfully",
		User:    pbUser,
	}, nil
}

// UpdateUser updates user information
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	log.Printf("UpdateUser RPC called for user ID: %d", req.Id)

	if req.Id == 0 {
		return &pb.UserResponse{
			Success: false,
			Message: "User ID is required",
		}, nil
	}

	// Convert proto request to update data
	updateData := &models.UserUpdateData{
		ID: req.Id,
	}

	// Set optional fields if provided
	if req.Name != nil {
		updateData.Name = req.Name
	}
	if req.Phone != nil {
		updateData.Phone = req.Phone
	}
	if req.IsActive != nil {
		updateData.IsActive = req.IsActive
	}

	// Call service layer
	updatedUser, err := s.userService.UpdateUser(ctx, updateData)
	if err != nil {
		log.Printf("UpdateUser service error: %v", err)

		if err.Error() == "user not found" {
			return &pb.UserResponse{
				Success: false,
				Message: "User not found",
			}, nil
		}

		return nil, status.Errorf(codes.Internal, "Failed to update user: %v", err)
	}

	// Convert domain model to proto response
	pbUser := s.modelToProtoUser(updatedUser)

	return &pb.UserResponse{
		Success: true,
		Message: "User updated successfully",
		User:    pbUser,
	}, nil
}

// DeleteUser deletes a user
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	log.Printf("DeleteUser RPC called for user ID: %d", req.Id)

	if req.Id == 0 {
		return &pb.DeleteUserResponse{
			Success: false,
			Message: "User ID is required",
		}, nil
	}

	// Call service layer
	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		log.Printf("DeleteUser service error: %v", err)

		if err.Error() == "user not found" {
			return &pb.DeleteUserResponse{
				Success: false,
				Message: "User not found",
			}, nil
		}

		return nil, status.Errorf(codes.Internal, "Failed to delete user: %v", err)
	}

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

// =================================
// Helper Methods
// =================================

// validateCreateUserRequest validates the create user request
func (s *UserServer) validateCreateUserRequest(req *pb.CreateUserRequest) error {
	if req.Email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Name == "" {
		return status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Password == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

// modelToProtoUser converts domain model User to protobuf User
func (s *UserServer) modelToProtoUser(user *models.User) *pb.User {
	pbUser := &pb.User{
		Id:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
		IsActive: user.IsActive,
	}

	// Convert timestamps
	if !user.CreatedAt.IsZero() {
		pbUser.CreatedAt = timestamppb.New(user.CreatedAt)
	}
	if !user.UpdatedAt.IsZero() {
		pbUser.UpdatedAt = timestamppb.New(user.UpdatedAt)
	}

	return pbUser
}
