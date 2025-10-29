package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// UserClient wraps the gRPC client for user-service with connection pooling
type UserClient struct {
	conn    *grpc.ClientConn         // Legacy: single connection
	pool    *grpcpool.ConnectionPool // New: connection pool
	client  pb.UserServiceClient
	timeout time.Duration
}

// NewUserClient creates a new user service gRPC client (legacy method)
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
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
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

// NewUserClientWithPool creates a new user service gRPC client with connection pooling
func NewUserClientWithPool(pool *grpcpool.ConnectionPool, timeout time.Duration) (*UserClient, error) {
	// Get a connection from the pool to create the client
	conn := pool.Get()

	return &UserClient{
		pool:    pool,
		client:  pb.NewUserServiceClient(conn),
		timeout: timeout,
	}, nil
}

// Close closes the gRPC connection (no-op for pooled connections)
func (c *UserClient) Close() error {
	// If using connection pool, connections are managed by the pool
	if c.pool != nil {
		return nil
	}

	// Legacy: close single connection
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getClient returns a client using either pooled or direct connection
func (c *UserClient) getClient() pb.UserServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewUserServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.client
}

// CreateUser creates a new user (equivalent to Register)
func (c *UserClient) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getClient()
	return client.CreateUser(ctx, req)
}

// Login authenticates a user
func (c *UserClient) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getClient()
	return client.Login(ctx, req)
}

// ValidateToken validates a JWT token
func (c *UserClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getClient()
	resp, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUser retrieves a user by ID
func (c *UserClient) GetUser(ctx context.Context, id int64) (*pb.User, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getClient()
	resp, err := client.GetUser(ctx, &pb.GetUserRequest{
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

	client := c.getClient()
	resp, err := client.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.User, nil
}

// DeleteUser deletes a user
func (c *UserClient) DeleteUser(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getClient()
	_, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
	return err
}

// RefreshToken refreshes the access token
func (c *UserClient) RefreshToken(ctx context.Context, refreshToken string) (*pb.LoginResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getClient()
	return client.RefreshToken(ctx, &pb.RefreshTokenRequest{RefreshToken: refreshToken})
}
