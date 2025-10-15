package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// UserClient wraps the gRPC client for user-service
type UserClient struct {
	conn    *grpc.ClientConn
	client  pb.UserServiceClient
	timeout time.Duration
}

// NewUserClient creates a new user service gRPC client
func NewUserClient(addr string, timeout time.Duration) (*UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service at %s: %w", addr, err)
	}

	return &UserClient{
		conn:    conn,
		client:  pb.NewUserServiceClient(conn),
		timeout: timeout,
	}, nil
}

// Close closes the gRPC connection
func (c *UserClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateUser creates a new user (equivalent to Register)
func (c *UserClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.client.CreateUser(ctx, req)
}

// Login authenticates a user
func (c *UserClient) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.client.Login(ctx, req)
}

// ValidateToken validates a JWT token
func (c *UserClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUser retrieves a user by ID
func (c *UserClient) GetUser(ctx context.Context, id int64) (*pb.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{
		Identifier: &pb.GetUserRequest_Id{Id: id},
	})
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

// UpdateUser updates user information
func (c *UserClient) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

// DeleteUser deletes a user
func (c *UserClient) DeleteUser(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	return err
}

// RefreshToken refreshes the access token
func (c *UserClient) RefreshToken(ctx context.Context, refreshToken string) (*pb.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: refreshToken})
}
