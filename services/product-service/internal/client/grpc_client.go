package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GRPCClients holds all gRPC client connections
type GRPCClients struct {
	UserClient UserServiceClient
}

// UserServiceClient interface for user service operations
type UserServiceClient interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	Close() error
}

// UserInfo represents user information returned from user service
type UserInfo struct {
	ID       string
	Email    string
	Username string
	Role     string
	IsActive bool
}

// userServiceClientImpl implements UserServiceClient
type userServiceClientImpl struct {
	conn *grpc.ClientConn
	// client pb.UserServiceClient // Uncomment when proto files are ready
}

// NewGRPCClients initializes all gRPC clients
func NewGRPCClients(userServiceAddr string) (*GRPCClients, error) {
	if userServiceAddr == "" {
		return nil, fmt.Errorf("user service address is required")
	}

	userClient, err := newUserServiceClient(userServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service client: %w", err)
	}

	return &GRPCClients{
		UserClient: userClient,
	}, nil
}

// newUserServiceClient creates a new user service gRPC client
func newUserServiceClient(addr string) (UserServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// gRPC connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // TODO: Use TLS in production
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Establish connection
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service at %s: %w", addr, err)
	}

	log.Printf("✅ Successfully connected to user service at %s", addr)

	return &userServiceClientImpl{
		conn: conn,
		// client: pb.NewUserServiceClient(conn), // Uncomment when proto files are ready
	}, nil
}

// ValidateToken validates a JWT token with user service
func (c *userServiceClientImpl) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	// TODO: Implement actual gRPC call when proto files are ready
	//
	// Example implementation:
	// req := &pb.ValidateTokenRequest{Token: token}
	// resp, err := c.client.ValidateToken(ctx, req)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to validate token: %w", err)
	// }
	//
	// return &UserInfo{
	//     ID:       resp.UserId,
	//     Email:    resp.Email,
	//     Username: resp.Username,
	//     Role:     resp.Role,
	//     IsActive: resp.IsActive,
	// }, nil

	// Mock implementation for development
	log.Printf("ValidateToken called with token: %s...", truncateString(token, 20))

	// Simulate validation delay
	time.Sleep(50 * time.Millisecond)

	return &UserInfo{
		ID:       "user-123",
		Email:    "admin@example.com",
		Username: "admin",
		Role:     "admin",
		IsActive: true,
	}, nil
}

// GetUserByID retrieves user information by ID from user service
func (c *userServiceClientImpl) GetUserByID(ctx context.Context, userID string) (*UserInfo, error) {
	// TODO: Implement actual gRPC call when proto files are ready
	//
	// Example implementation:
	// req := &pb.GetUserRequest{UserId: userID}
	// resp, err := c.client.GetUser(ctx, req)
	// if err != nil {
	//     if status.Code(err) == codes.NotFound {
	//         return nil, fmt.Errorf("user not found")
	//     }
	//     return nil, fmt.Errorf("failed to get user: %w", err)
	// }
	//
	// return &UserInfo{
	//     ID:       resp.UserId,
	//     Email:    resp.Email,
	//     Username: resp.Username,
	//     Role:     resp.Role,
	//     IsActive: resp.IsActive,
	// }, nil

	// Mock implementation for development
	log.Printf("GetUserByID called with userID: %s", userID)

	// Simulate lookup delay
	time.Sleep(30 * time.Millisecond)

	return &UserInfo{
		ID:       userID,
		Email:    "user@example.com",
		Username: "user",
		Role:     "user",
		IsActive: true,
	}, nil
}

// CheckPermission checks if a user has a specific permission
func (c *userServiceClientImpl) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	// TODO: Implement actual gRPC call when proto files are ready
	//
	// Example implementation:
	// req := &pb.CheckPermissionRequest{
	//     UserId:     userID,
	//     Permission: permission,
	// }
	// resp, err := c.client.CheckPermission(ctx, req)
	// if err != nil {
	//     return false, fmt.Errorf("failed to check permission: %w", err)
	// }
	//
	// return resp.HasPermission, nil

	// Mock implementation for development
	log.Printf("CheckPermission called - userID: %s, permission: %s", userID, permission)

	// Simulate permission check
	time.Sleep(20 * time.Millisecond)

	// Mock: admins have all permissions
	return true, nil
}

// Close closes the gRPC connection
func (c *userServiceClientImpl) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close user service connection: %w", err)
		}
		log.Println("✅ User service client connection closed")
	}
	return nil
}

// CloseAll closes all gRPC client connections
func (g *GRPCClients) CloseAll() error {
	var errs []error

	if g.UserClient != nil {
		if err := g.UserClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close user client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing gRPC clients: %v", errs)
	}

	log.Println("✅ All gRPC client connections closed")
	return nil
}

// Helper function to truncate string for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
