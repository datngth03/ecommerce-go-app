// internal/auth/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5" // JWT library
	"github.com/google/uuid"       // For generating UUIDs for refresh tokens

	"github.com/datngth03/ecommerce-go-app/internal/auth/domain"
	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth" // Generated Auth gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user" // User gRPC client for user validation
)

// Claims defines the JWT claims structure.
// Claims định nghĩa cấu trúc của JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthService defines the application service interface for authentication and authorization operations.
// AuthService định nghĩa interface dịch vụ ứng dụng cho các thao tác xác thực và ủy quyền.
type AuthService interface {
	// Login authenticates a user and generates new access and refresh tokens.
	Login(ctx context.Context, req *auth_client.LoginRequest) (*auth_client.AuthResponse, error) // THÊM DÒNG NÀY
	// AuthenticateUser authenticates a user using email and password (calls User Service).
	// This is an internal method, not an RPC.
	AuthenticateUser(ctx context.Context, req *auth_client.AuthenticateUserRequest) (*auth_client.AuthenticateUserResponse, error) // THÊM DÒNG NÀY (đã đổi từ email, password string sang req)

	// GenerateTokens generates new access and refresh tokens for a user.
	GenerateTokens(ctx context.Context, userID string) (*auth_client.AuthResponse, error)
	// RefreshToken generates a new access token using a refresh token.
	RefreshToken(ctx context.Context, req *auth_client.RefreshTokenRequest) (*auth_client.AuthResponse, error)
	// ValidateToken validates an access token.
	ValidateToken(ctx context.Context, req *auth_client.ValidateTokenRequest) (*auth_client.ValidateTokenResponse, error)
}

// authService implements the AuthService interface.
// authService triển khai interface AuthService.
type authService struct {
	refreshTokenRepo domain.RefreshTokenRepository
	userClient       user_client.UserServiceClient // Client to call User Service
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	// TODO: Add other dependencies like logger
}

// NewAuthService creates a new instance of AuthService.
// NewAuthService tạo một thể hiện mới của AuthService.
func NewAuthService(
	refreshTokenRepo domain.RefreshTokenRepository,
	userClient user_client.UserServiceClient,
	jwtSecret string,
	accessTokenTTL, refreshTokenTTL time.Duration,
) AuthService {
	return &authService{
		refreshTokenRepo: refreshTokenRepo,
		userClient:       userClient,
		jwtSecret:        []byte(jwtSecret),
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
	}
}

// Login authenticates a user and generates new access and refresh tokens.
// Login xác thực người dùng và tạo access và refresh token mới.
func (s *authService) Login(ctx context.Context, req *auth_client.LoginRequest) (*auth_client.AuthResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, errors.New("email and password are required")
	}

	// Authenticate user credentials via User Service
	userLoginReq := &user_client.LoginUserRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	userLoginResp, err := s.userClient.LoginUser(ctx, userLoginReq)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	if userLoginResp.GetUserId() == "" {
		return nil, errors.New("authentication failed: no user ID returned")
	}

	// Generate tokens for the authenticated user
	authResp, err := s.GenerateTokens(ctx, userLoginResp.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return authResp, nil
}

// AuthenticateUser authenticates a user by calling the User Service.
// This is an internal method, not an RPC.
// AuthenticateUser xác thực người dùng bằng cách gọi User Service.
// Đây là một phương thức nội bộ, không phải là một RPC.
func (s *authService) AuthenticateUser(ctx context.Context, req *auth_client.AuthenticateUserRequest) (*auth_client.AuthenticateUserResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, errors.New("email and password are required")
	}

	userLoginReq := &user_client.LoginUserRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}
	userLoginResp, err := s.userClient.LoginUser(ctx, userLoginReq)
	if err != nil {
		return &auth_client.AuthenticateUserResponse{
			IsAuthenticated: false,
			ErrorMessage:    fmt.Sprintf("authentication failed: %v", err),
		}, nil
	}
	if userLoginResp.GetUserId() == "" {
		return &auth_client.AuthenticateUserResponse{
			IsAuthenticated: false,
			ErrorMessage:    "authentication failed: no user ID returned",
		}, nil
	}

	return &auth_client.AuthenticateUserResponse{
		IsAuthenticated: true,
		UserId:          userLoginResp.GetUserId(),
	}, nil
}

// GenerateTokens generates new access and refresh tokens for a user.
// GenerateTokens tạo access và refresh token mới cho người dùng.
func (s *authService) GenerateTokens(ctx context.Context, userID string) (*auth_client.AuthResponse, error) {
	// Access Token
	accessTokenClaims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshTokenValue := uuid.New().String() // Generate a random UUID for refresh token
	refreshTokenExpiresAt := time.Now().Add(s.refreshTokenTTL)
	refreshTokenEntity := domain.NewRefreshToken(refreshTokenValue, userID, refreshTokenExpiresAt)

	if err := s.refreshTokenRepo.Save(ctx, refreshTokenEntity); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &auth_client.AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenValue,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		UserId:       userID,
	}, nil
}

// RefreshToken generates a new access token using a refresh token.
// RefreshToken tạo access token mới bằng refresh token.
func (s *authService) RefreshToken(ctx context.Context, req *auth_client.RefreshTokenRequest) (*auth_client.AuthResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, errors.New("refresh token is required")
	}

	refreshTokenEntity, err := s.refreshTokenRepo.FindByToken(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, errors.New("refresh token not found")) { // Assuming specific error from repo
			return nil, errors.New("invalid refresh token")
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	if time.Now().After(refreshTokenEntity.ExpiresAt) {
		// Invalidate expired refresh token
		s.refreshTokenRepo.Delete(ctx, refreshTokenEntity.Token)
		return nil, errors.New("refresh token expired")
	}

	// Generate new tokens
	newAuthResponse, err := s.GenerateTokens(ctx, refreshTokenEntity.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Invalidate the old refresh token (optional, but good for security)
	s.refreshTokenRepo.Delete(ctx, refreshTokenEntity.Token)

	return newAuthResponse, nil
}

// ValidateToken validates an access token.
// ValidateToken xác thực access token.
func (s *authService) ValidateToken(ctx context.Context, req *auth_client.ValidateTokenRequest) (*auth_client.ValidateTokenResponse, error) {
	if req.GetAccessToken() == "" {
		return nil, errors.New("access token is required")
	}

	token, err := jwt.ParseWithClaims(req.GetAccessToken(), &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		// Handle specific JWT errors
		if errors.Is(err, jwt.ErrTokenExpired) {
			return &auth_client.ValidateTokenResponse{IsValid: false, ErrorMessage: "token expired"}, nil
		}
		if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return &auth_client.ValidateTokenResponse{IsValid: false, ErrorMessage: "invalid token"}, nil
		}
		return &auth_client.ValidateTokenResponse{IsValid: false, ErrorMessage: fmt.Sprintf("token validation failed: %v", err)}, nil
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &auth_client.ValidateTokenResponse{IsValid: true, UserId: claims.UserID}, nil
	}

	return &auth_client.ValidateTokenResponse{IsValid: false, ErrorMessage: "invalid token claims"}, nil
}
