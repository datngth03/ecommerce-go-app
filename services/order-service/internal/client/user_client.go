package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	sharedConfig "github.com/ecommerce-go-app/shared/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn   *grpc.ClientConn
	client pb.UserServiceClient
}

func NewUserClient(endpoint sharedConfig.ServiceEndpoint) (*UserClient, error) {
	conn, err := grpc.Dial(endpoint.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	return &UserClient{
		conn:   conn,
		client: pb.NewUserServiceClient(conn),
	}, nil
}

func (c *UserClient) Close() error {
	return c.conn.Close()
}

// GetUser retrieves user details by ID
func (c *UserClient) GetUser(ctx context.Context, userID int64) (*pb.User, error) {
	resp, err := c.client.GetUser(ctx, &pb.GetUserRequest{
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
