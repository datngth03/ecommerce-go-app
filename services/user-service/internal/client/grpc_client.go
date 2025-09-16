// internal/client/grpc_client.go
package client

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// GRPCClient wraps the user service gRPC client
type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

// NewGRPCClient creates a new gRPC client connection to user service
func NewGRPCClient(serverAddress string) (*GRPCClient, error) {
	log.Printf("Connecting to user service at %s", serverAddress)

	// Set up connection options
	opts := []grpc.DialOption{
		// Use insecure connection (for development)
		grpc.WithTransportCredentials(insecure.NewCredentials()),

		// Keep alive settings
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Send keepalive pings every 10 seconds
			Timeout:             5 * time.Second,  // Wait 5 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // Send pings even without active streams
		}),

		// Connection timeout
		grpc.WithTimeout(10 * time.Second),

		// Retry settings
		grpc.WithDefaultCallOptions(
			grpc.WaitForReady(true), // Wait for connection to be ready
		),
	}

	// Establish connection
	conn, err := grpc.Dial(serverAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create client
	client := pb.NewUserServiceClient(conn)

	_, err = client.GetUser(ctx, &pb.GetUserRequest{Identifier: &pb.GetUserRequest_Id{Id: -1}})
	if err != nil {
		log.Printf("Connection test failed (expected): %v", err)
	}
	// Test connection by calling a health check if available
	// You can add a health check call here if your proto has one
	log.Printf("Successfully connected to user service at %s", serverAddress)

	return &GRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		log.Println("Closing gRPC connection to user service")
		return c.conn.Close()
	}
	return nil
}

// GetClient returns the underlying gRPC client
func (c *GRPCClient) GetClient() pb.UserServiceClient {
	return c.client
}

// CreateUser creates a new user via gRPC
func (c *GRPCClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Printf("Creating user with email: %s", req.Email)

	resp, err := c.client.CreateUser(ctx, req)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return nil, fmt.Errorf("create user failed: %w", err)
	}

	log.Printf("User created successfully with ID: %s", resp.User.Id)
	return resp, nil
}

// GetUserByID retrieves a user by ID via gRPC
func (c *GRPCClient) GetUserByID(ctx context.Context, userID int64) (*pb.UserResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Getting user with ID: %d", userID)

	req := &pb.GetUserRequest{
		Identifier: &pb.GetUserRequest_Id{Id: userID},
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		return nil, fmt.Errorf("get user by ID failed: %w", err)
	}

	log.Printf("User retrieved successfully by ID: %s", resp.User.Email)
	return resp, nil
}

// GetUserByEmail retrieves a user by email via gRPC
func (c *GRPCClient) GetUserByEmail(ctx context.Context, email string) (*pb.UserResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Getting user with email: %s", email)

	req := &pb.GetUserRequest{
		Identifier: &pb.GetUserRequest_Email{Email: email},
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		log.Printf("Failed to get user by email: %v", err)
		return nil, fmt.Errorf("get user by email failed: %w", err)
	}

	log.Printf("User retrieved successfully by email: %s", resp.User.Email)
	return resp, nil
}

// GetUser is a generic method that accepts a GetUserRequest directly
func (c *GRPCClient) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Log based on what identifier is set
	switch identifier := req.Identifier.(type) {
	case *pb.GetUserRequest_Id:
		log.Printf("Getting user with ID: %d", identifier.Id)
	case *pb.GetUserRequest_Email:
		log.Printf("Getting user with email: %s", identifier.Email)
	default:
		log.Printf("Getting user with unknown identifier type")
	}

	resp, err := c.client.GetUser(ctx, req)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	log.Printf("User retrieved successfully: %s", resp.User.Email)
	return resp, nil
}

// UpdateUser updates a user via gRPC
func (c *GRPCClient) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Printf("Updating user with ID: %s", req.Id)

	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return nil, fmt.Errorf("update user failed: %w", err)
	}

	log.Printf("User updated successfully: %s", resp.User.Email)
	return resp, nil
}

// DeleteUser deletes a user via gRPC
func (c *GRPCClient) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Deleting user with ID: %s", req.Id)

	resp, err := c.client.DeleteUser(ctx, req)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return nil, fmt.Errorf("delete user failed: %w", err)
	}

	log.Printf("User deleted successfully with ID: %s", req.Id)
	return resp, nil
}

// LoginUser authenticates a user via gRPC
func (c *GRPCClient) LoginUser(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	log.Printf("Logging in user with email: %s", req.Email)

	resp, err := c.client.Login(ctx, req)
	if err != nil {
		log.Printf("Failed to login user: %v", err)
		return nil, fmt.Errorf("login failed: %w", err)
	}

	log.Printf("User logged in successfully: %s", req.Email)
	return resp, nil
}

// ChangePassword changes user password via gRPC
func (c *GRPCClient) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	log.Printf("Changing password for authenticated user")

	resp, err := c.client.ChangePassword(ctx, req)
	if err != nil {
		log.Printf("Failed to change password: %v", err)
		return nil, fmt.Errorf("change password failed: %w", err)
	}

	log.Printf("Password changed successfully for authenticated user")
	return resp, nil
}

// ChangePasswordWithAuth is a convenience method that creates the request with old and new passwords
func (c *GRPCClient) ChangePasswordWithAuth(ctx context.Context, oldPassword, newPassword string) (*pb.ChangePasswordResponse, error) {
	req := &pb.ChangePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	return c.ChangePassword(ctx, req)
}

// RefreshToken refreshes user token via gRPC
func (c *GRPCClient) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.LoginResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Refreshing token for user")

	resp, err := c.client.RefreshToken(ctx, req)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		return nil, fmt.Errorf("refresh token failed: %w", err)
	}

	log.Printf("Token refreshed successfully")
	return resp, nil
}

// ValidateToken validates user token via gRPC
func (c *GRPCClient) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	ctx = c.addMetadata(ctx)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Printf("Validating token")

	resp, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		log.Printf("Failed to validate token: %v", err)
		return nil, fmt.Errorf("validate token failed: %w", err)
	}

	log.Printf("Token validated successfully")
	return resp, nil
}

// addMetadata adds common metadata to the context
func (c *GRPCClient) addMetadata(ctx context.Context) context.Context {
	md := metadata.Pairs(
		"client", "api-gateway",
		"version", "1.0.0",
		"timestamp", time.Now().UTC().Format(time.RFC3339),
	)
	return metadata.NewOutgoingContext(ctx, md)
}

// HealthCheck performs a health check on the gRPC connection
func (c *GRPCClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// If your proto has a health check method, use it here
	// For now, we'll just try to get a non-existent user to test connectivity
	req := &pb.GetUserRequest{
		Identifier: &pb.GetUserRequest_Id{Id: -1}, // Non-existent ID for health check
	}

	_, err := c.client.GetUser(ctx, req)

	// We expect this to fail with "not found" rather than connection errors
	if err != nil {
		// Check if it's a connection error vs. application error
		log.Printf("Health check error (expected): %v", err)
	}

	return nil
}

// IsConnected checks if the gRPC connection is active
func (c *GRPCClient) IsConnected() bool {
	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Connecting || state == connectivity.Ready
}
