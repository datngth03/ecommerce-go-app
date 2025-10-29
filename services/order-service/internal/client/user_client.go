package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
	pool   *grpcpool.ConnectionPool // Connection pool support
}

func NewUserClient(endpoint sharedConfig.ServiceEndpoint) (*UserClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(endpoint.GRPCAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	return &UserClient{
		conn:   conn,
		client: pb.NewUserServiceClient(conn),
	}, nil
}

// NewUserClientWithPool creates a new user client with connection pooling support
func NewUserClientWithPool(endpoint sharedConfig.ServiceEndpoint, poolManager *grpcpool.Manager) (*UserClient, error) {
	pool, exists := poolManager.Get("user")
	if !exists {
		poolConfig := grpcpool.DefaultPoolConfig(endpoint.GRPCAddr)
		var err error
		pool, err = poolManager.GetOrCreate("user", poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create user service pool: %w", err)
		}
	}

	return &UserClient{
		pool: pool,
	}, nil
}

func (c *UserClient) Close() error {
	// If using pool, don't close individual connections
	if c.pool != nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getClient returns a gRPC client, using pool if available
func (c *UserClient) getClient() (pb.UserServiceClient, error) {
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewUserServiceClient(conn), nil
	}
	return c.client, nil
}

// GetUser retrieves user details by ID
func (c *UserClient) GetUser(ctx context.Context, userID int64) (*pb.User, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetUser(ctx, &pb.GetUserRequest{
		Identifier: &pb.GetUserRequest_Id{Id: userID},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("user not found: %s", resp.Message)
	}

	return resp.User, nil
}

// ValidateUser checks if user exists and is active
func (c *UserClient) ValidateUser(ctx context.Context, userID int64) (bool, error) {
	user, err := c.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}
