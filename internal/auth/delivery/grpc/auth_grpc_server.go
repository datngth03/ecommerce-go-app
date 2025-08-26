package grpc

import (
	"context"
	"log"

	"github.com/datngth03/ecommerce-go-app/internal/auth/application"
	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth"
)

// AuthGRPCServer implements the auth_client.AuthServiceServer interface.
type AuthGRPCServer struct {
	auth_client.UnimplementedAuthServiceServer
	authService application.AuthService
}

// NewAuthGRPCServer creates a new instance of AuthGRPCServer.
func NewAuthGRPCServer(svc application.AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{
		authService: svc,
	}
}

// Register implements the gRPC Register method.
func (s *AuthGRPCServer) Register(ctx context.Context, req *auth_client.RegisterRequest) (*auth_client.RegisterResponse, error) {
	log.Printf("Received Register request for email: %s", req.GetEmail())
	resp, err := s.authService.Register(ctx, req)
	if err != nil {
		log.Printf("Error during registration: %v", err)
		return nil, err
	}
	return resp, nil
}

// Login implements the gRPC Login method.
func (s *AuthGRPCServer) Login(ctx context.Context, req *auth_client.LoginRequest) (*auth_client.LoginResponse, error) {
	log.Printf("Received Login request for email: %s", req.GetEmail())
	resp, err := s.authService.Login(ctx, req)
	if err != nil {
		log.Printf("Error during login: %v", err)
		return nil, err
	}
	return resp, nil
}

// LoginWithGoogle implements the gRPC LoginWithGoogle method.
func (s *AuthGRPCServer) LoginWithGoogle(ctx context.Context, req *auth_client.LoginWithGoogleRequest) (*auth_client.LoginResponse, error) {
	log.Printf("Received LoginWithGoogle request")
	resp, err := s.authService.LoginWithGoogle(ctx, req)
	if err != nil {
		log.Printf("Error during LoginWithGoogle: %v", err)
		return nil, err
	}
	return resp, nil
}

// RefreshToken implements the gRPC RefreshToken method.
func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *auth_client.RefreshTokenRequest) (*auth_client.RefreshTokenResponse, error) {
	log.Printf("Received RefreshToken request.")
	resp, err := s.authService.RefreshToken(ctx, req)
	if err != nil {
		log.Printf("Error refreshing token: %v", err)
		return nil, err
	}
	return resp, nil
}

// Logout implements the gRPC Logout method.
func (s *AuthGRPCServer) Logout(ctx context.Context, req *auth_client.LogoutRequest) (*auth_client.LogoutResponse, error) {
	log.Printf("Received Logout request.")
	resp, err := s.authService.Logout(ctx, req)
	if err != nil {
		log.Printf("Error during logout: %v", err)
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
