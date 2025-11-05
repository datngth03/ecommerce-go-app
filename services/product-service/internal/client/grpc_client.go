package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

// GRPCClients holds all gRPC client connections
type GRPCClients struct {
	UserClient UserServiceClient
}

// UserServiceClient interface for user service operations
type UserServiceClient interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
	GetUserByID(ctx context.Context, userID string) (*UserInfo, error)
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
	conn   *grpc.ClientConn
	client pb.UserServiceClient
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
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
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

	log.Printf("Successfully connected to user service at %s", addr)

	return &userServiceClientImpl{
		conn:   conn,
		client: pb.NewUserServiceClient(conn),
	}, nil
}

// ValidateToken validates a JWT token with user service
func (c *userServiceClientImpl) ValidateToken(ctx context.Context, token string) (*UserInfo, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	req := &pb.ValidateTokenRequest{
		Token: token,
	}

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return nil, fmt.Errorf("invalid or expired token")
			case codes.NotFound:
				return nil, fmt.Errorf("user not found")
			default:
				return nil, fmt.Errorf("failed to validate token: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("token validation failed: %s", resp.Message)
	}

	return &UserInfo{
		ID:       fmt.Sprintf("%d", resp.UserId),
		Email:    resp.Email,
		Username: resp.Email, // Use email as username if not provided
		Role:     "user",     // Default role, can be enhanced
		IsActive: true,
	}, nil
}

// GetUserByID retrieves user information by ID from user service
func (c *userServiceClientImpl) GetUserByID(ctx context.Context, userID string) (*UserInfo, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Convert string ID to int64
	var id int64
	_, err := fmt.Sscanf(userID, "%d", &id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	req := &pb.GetUserRequest{
		Identifier: &pb.GetUserRequest_Id{Id: id},
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user not found")
			default:
				return nil, fmt.Errorf("failed to get user: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get user: %s", resp.Message)
	}

	if resp.User == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &UserInfo{
		ID:       fmt.Sprintf("%d", resp.User.Id),
		Email:    resp.User.Email,
		Username: resp.User.Name,
		Role:     "user", // Can be enhanced with role from database
		IsActive: resp.User.IsActive,
	}, nil
}

// CheckPermission checks if a user has a specific permission
// Note: This is a mock implementation as CheckPermission RPC is not defined in user_service.proto
// You may need to add this RPC to the proto file or implement permission checking differently

// Close closes the gRPC connection
func (c *userServiceClientImpl) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close user service connection: %w", err)
		}
		log.Println("User service client connection closed")
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

	log.Println("All gRPC client connections closed")
	return nil
}

// Helper function to truncate string for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
