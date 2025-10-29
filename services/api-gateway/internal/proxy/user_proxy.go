package proxy

import (
	"context"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/metrics"
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
	start := time.Now()
	resp, err := p.client.CreateUser(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("user-service", "CreateUser", status, time.Since(start))
	metrics.RecordProxyRequest("user-service", status, time.Since(start))

	return resp, err
}

// Login authenticates a user
func (p *UserProxy) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	start := time.Now()
	resp, err := p.client.Login(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordAuthFailure()
	}
	metrics.RecordGRPCClientRequest("user-service", "Login", status, time.Since(start))
	metrics.RecordAuthRequest("login", status)

	return resp, err
}

// ValidateToken validates a JWT token
func (p *UserProxy) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	start := time.Now()
	resp, err := p.client.ValidateToken(ctx, token)

	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordAuthFailure()
	}
	metrics.RecordGRPCClientRequest("user-service", "ValidateToken", status, time.Since(start))
	metrics.RecordAuthRequest("validate", status)

	return resp, err
}

// GetUser retrieves a user by ID
func (p *UserProxy) GetUser(ctx context.Context, id int64) (*pb.User, error) {
	start := time.Now()
	resp, err := p.client.GetUser(ctx, id)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("user-service", "GetUser", status, time.Since(start))
	metrics.RecordProxyRequest("user-service", status, time.Since(start))

	return resp, err
}

// UpdateUser updates user information
func (p *UserProxy) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	start := time.Now()
	resp, err := p.client.UpdateUser(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("user-service", "UpdateUser", status, time.Since(start))
	metrics.RecordProxyRequest("user-service", status, time.Since(start))

	return resp, err
}

// DeleteUser deletes a user
func (p *UserProxy) DeleteUser(ctx context.Context, id int64) error {
	start := time.Now()
	err := p.client.DeleteUser(ctx, id)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("user-service", "DeleteUser", status, time.Since(start))
	metrics.RecordProxyRequest("user-service", status, time.Since(start))

	return err
}

// RefreshToken refreshes the access token
func (p *UserProxy) RefreshToken(ctx context.Context, refreshToken string) (*pb.LoginResponse, error) {
	start := time.Now()
	resp, err := p.client.RefreshToken(ctx, refreshToken)

	status := "success"
	if err != nil {
		status = "error"
		metrics.RecordAuthFailure()
	}
	metrics.RecordGRPCClientRequest("user-service", "RefreshToken", status, time.Since(start))
	metrics.RecordAuthRequest("refresh", status)

	return resp, err
}
