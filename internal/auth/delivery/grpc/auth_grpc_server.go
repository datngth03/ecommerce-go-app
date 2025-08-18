// internal/auth/delivery/grpc/auth_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/auth/application"
	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth" // Generated Auth gRPC client
)

// AuthGRPCServer implements the auth_client.AuthServiceServer interface.
type AuthGRPCServer struct {
	auth_client.UnimplementedAuthServiceServer // Embedded to satisfy all methods
	authService                                application.AuthService
}

// NewAuthGRPCServer creates a new instance of AuthGRPCServer.
func NewAuthGRPCServer(svc application.AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{
		authService: svc,
	}
}

// Login implements the gRPC Login method. (THÊM PHẦN NÀY)
func (s *AuthGRPCServer) Login(ctx context.Context, req *auth_client.LoginRequest) (*auth_client.AuthResponse, error) {
	log.Printf("Received Login request for email: %s", req.GetEmail())
	resp, err := s.authService.Login(ctx, req)
	if err != nil {
		log.Printf("Error during login: %v", err)
		return nil, err
	}
	return resp, nil
}

// AuthenticateUser implements the gRPC AuthenticateUser method. (THÊM PHẦN NÀY)
func (s *AuthGRPCServer) AuthenticateUser(ctx context.Context, req *auth_client.AuthenticateUserRequest) (*auth_client.AuthenticateUserResponse, error) {
	log.Printf("Received AuthenticateUser request for email: %s", req.GetEmail())
	resp, err := s.authService.AuthenticateUser(ctx, req)
	if err != nil {
		log.Printf("Error during AuthenticateUser: %v", err)
		return nil, err
	}
	return resp, nil
}

// RefreshToken implements the gRPC RefreshToken method.
func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *auth_client.RefreshTokenRequest) (*auth_client.AuthResponse, error) {
	log.Printf("Received RefreshToken request.")
	resp, err := s.authService.RefreshToken(ctx, req)
	if err != nil {
		log.Printf("Error refreshing token: %v", err)
		return nil, err
	}
	return resp, nil
}

// ValidateToken implements the gRPC ValidateToken method.
func (s *AuthGRPCServer) ValidateToken(ctx context.Context, req *auth_client.ValidateTokenRequest) (*auth_client.ValidateTokenResponse, error) {
	log.Printf("Received ValidateToken request.")
	resp, err := s.authService.ValidateToken(ctx, req)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		return nil, err
	}
	return resp, nil
}

// LoginWithGoogle implements the gRPC LoginWithGoogle method.
func (s *AuthGRPCServer) LoginWithGoogle(ctx context.Context, req *auth_client.LoginWithGoogleRequest) (*auth_client.AuthResponse, error) {
	log.Printf("Received LoginWithGoogle request")
	resp, err := s.authService.LoginWithGoogle(ctx, req)
	if err != nil {
		log.Printf("Error during LoginWithGoogle: %v", err)
		return nil, err
	}
	return resp, nil
}
