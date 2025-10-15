package proxy

import (
	"context"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
)

// UserProxy adapts user client for HTTP handlers
type UserProxy struct {
	client *clients.UserClient
}

// NewUserProxy creates a new user proxy
func NewUserProxy(client *clients.UserClient) *UserProxy {
	return &UserProxy{client: client}
}

// CreateUser creates a new user (equivalent to Register)
func (p *UserProxy) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	return p.client.CreateUser(ctx, req)
}

// Login authenticates a user
func (p *UserProxy) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return p.client.Login(ctx, req)
}

// ValidateToken validates a JWT token
func (p *UserProxy) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	return p.client.ValidateToken(ctx, token)
}

// GetUser retrieves a user by ID
func (p *UserProxy) GetUser(ctx context.Context, id int64) (*pb.User, error) {
	return p.client.GetUser(ctx, id)
}

// UpdateUser updates user information
func (p *UserProxy) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	return p.client.UpdateUser(ctx, req)
}

// DeleteUser deletes a user
func (p *UserProxy) DeleteUser(ctx context.Context, id int64) error {
	return p.client.DeleteUser(ctx, id)
}

// RefreshToken refreshes the access token
func (p *UserProxy) RefreshToken(ctx context.Context, refreshToken string) (*pb.LoginResponse, error) {
	return p.client.RefreshToken(ctx, refreshToken)
}
