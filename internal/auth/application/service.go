// internal/auth/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5" // JWT library
	"github.com/google/uuid"       // For generating UUIDs for refresh tokens
	"go.uber.org/zap"

	"net/http"

	oauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/datngth03/ecommerce-go-app/internal/auth/domain"
	auth_domain "github.com/datngth03/ecommerce-go-app/internal/auth/domain"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"
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

	// Register registers a new user.
	Register(ctx context.Context, req *auth_client.RegisterRequest) (*auth_client.RegisterResponse, error)
	// Login authenticates a user and generates new access and refresh tokens.
	Login(ctx context.Context, req *auth_client.LoginRequest) (*auth_client.LoginResponse, error)
	// LoginWithGoogle authenticates a user using Google ID token.
	LoginWithGoogle(ctx context.Context, req *auth_client.LoginWithGoogleRequest) (*auth_client.LoginResponse, error)
	// GenerateTokens generates new access and refresh tokens for a user.
	GenerateTokens(ctx context.Context, userID string) (*auth_client.LoginResponse, error)
	// RefreshToken generates a new access token using a refresh token.
	RefreshToken(ctx context.Context, req *auth_client.RefreshTokenRequest) (*auth_client.RefreshTokenResponse, error)
	// ValidateToken validates an access token.
	ValidateToken(ctx context.Context, req *auth_client.ValidateTokenRequest) (*auth_client.ValidateTokenResponse, error)
	// Logout invalidates a refresh token, effectively logging the user out.
	Logout(ctx context.Context, req *auth_client.LogoutRequest) (*auth_client.LogoutResponse, error)
}

func NewAuthService(
	repo auth_domain.RefreshTokenRepository,
	userClient user_client.UserServiceClient,
	secret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) AuthService {
	return &authService{
		refreshTokenRepo: repo,
		userClient:       userClient,
		jwtSecret:        secret,
		accessTokenTTL:   accessTTL,
		refreshTokenTTL:  refreshTTL,
	}
}

// authService implements the AuthService interface.
// authService triển khai interface AuthService.
type authService struct {
	refreshTokenRepo auth_domain.RefreshTokenRepository
	userClient       user_client.UserServiceClient // Client to call User Service
	jwtSecret        string
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	// TODO: Add other dependencies like logger
}

// Register registers a new user.
// Register triển khai logic đăng ký người dùng mới.
func (s *authService) Register(ctx context.Context, req *auth_client.RegisterRequest) (*auth_client.RegisterResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" || req.GetFirstName() == "" || req.GetLastName() == "" {
		return nil, errors.New("missing required fields")
	}

	createReq := &user_client.RegisterUserRequest{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(), // gửi password raw
		FirstName: req.GetFirstName(),
		LastName:  req.GetLastName(),
	}

	regResp, err := s.userClient.RegisterUser(ctx, createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &auth_client.RegisterResponse{
		UserId:  regResp.GetUserId(),
		Message: "User registered successfully",
	}, nil
}

// Login authenticates a user and generates new access and refresh tokens.
// Login xác thực người dùng và tạo access và refresh token mới.
func (s *authService) Login(ctx context.Context, req *auth_client.LoginRequest) (*auth_client.LoginResponse, error) {
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

// LoginWithGoogle xử lý logic xác thực người dùng bằng Google Token.
// LoginWithGoogle authenticates a user using Google ID token.
func (s *authService) LoginWithGoogle(ctx context.Context, req *auth_client.LoginWithGoogleRequest) (*auth_client.LoginResponse, error) {
	if req.IdToken == "" {
		return nil, status.Error(codes.InvalidArgument, "id token is required")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		return nil, status.Error(codes.Internal, "missing GOOGLE_CLIENT_ID env variable")
	}

	// Init OAuth2 service
	oauth2Service, err := oauth2.NewService(ctx, option.WithHTTPClient(&http.Client{}))
	if err != nil {
		logger.Logger.Error("Failed to create OAuth2 service", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create auth service")
	}

	// Validate Google ID token
	tokenInfo, err := oauth2Service.Tokeninfo().IdToken(req.IdToken).Context(ctx).Do()
	if err != nil {
		logger.Logger.Error("Failed to validate Google token", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid Google token")
	}

	if tokenInfo.Audience != googleClientID {
		logger.Logger.Warn("Google token audience mismatch",
			zap.String("expected_audience", googleClientID),
			zap.String("actual_audience", tokenInfo.Audience),
		)
		return nil, status.Error(codes.Unauthenticated, "token audience mismatch")
	}

	userinfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch userinfo: %v", err)
	}
	fmt.Println("Google User:", userinfo.Email, userinfo.Name, userinfo.Picture)

	userEmail := userinfo.Email
	userLastName := userinfo.FamilyName
	userFirstName := userinfo.GivenName

	// Try to find user
	getUserReq := &user_client.GetUserByEmailRequest{Email: userEmail}
	userResp, err := s.userClient.GetUserByEmail(ctx, getUserReq)

	var userID string

	if err != nil || userResp.GetUser() == nil {
		// If not found → create
		createReq := &user_client.RegisterUserRequest{
			Email:     userEmail,
			Password:  uuid.NewString(),
			FirstName: userFirstName,
			LastName:  userLastName,
		}

		regResp, err := s.userClient.RegisterUser(ctx, createReq)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
		}

		userID = regResp.GetUserId()
	} else {
		userID = userResp.GetUser().GetUserId()
	}

	// Issue JWT + Refresh token
	tokens, err := s.GenerateTokens(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return tokens, nil
}

// GenerateTokens generates new access and refresh tokens for a user.
// GenerateTokens tạo access và refresh token mới cho người dùng.
func (s *authService) GenerateTokens(ctx context.Context, userID string) (*auth_client.LoginResponse, error) {
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
	accessTokenExpiresAt := time.Now().Add(s.accessTokenTTL)
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

	return &auth_client.LoginResponse{
		AccessToken:           accessTokenString,
		RefreshToken:          refreshTokenValue,
		AccessTokenExpiresAt:  timestamppb.New(accessTokenExpiresAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshTokenExpiresAt),
		UserId:                userID,
	}, nil
}

// RefreshToken generates a new access token using a refresh token.
// RefreshToken tạo access token mới bằng refresh token.
func (s *authService) RefreshToken(ctx context.Context, req *auth_client.RefreshTokenRequest) (*auth_client.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required")
	}

	// Tìm refresh token trong repository
	refreshTokenEntity, err := s.refreshTokenRepo.FindByToken(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			return nil, status.Error(codes.Unauthenticated, "Invalid refresh token")
		}
		return nil, status.Errorf(codes.Internal, "Failed to find refresh token: %v", err)
	}

	// Kiểm tra xem token có hết hạn chưa
	if time.Now().After(refreshTokenEntity.ExpiresAt) {
		// Vô hiệu hóa refresh token đã hết hạn
		s.refreshTokenRepo.Delete(ctx, refreshTokenEntity.Token)
		return nil, status.Error(codes.Unauthenticated, "Refresh token expired")
	}

	// Tạo các token mới
	authResp, err := s.GenerateTokens(ctx, refreshTokenEntity.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate new tokens: %v", err)
	}

	// Vô hiệu hóa refresh token cũ để tăng cường bảo mật
	if err := s.refreshTokenRepo.Delete(ctx, refreshTokenEntity.Token); err != nil {
		logger.Logger.Info("Failed to delete old refresh token")
		// Không trả về lỗi, vì việc tạo token mới vẫn thành công
	}

	// Trả về phản hồi mới
	return &auth_client.RefreshTokenResponse{
		AccessToken:           authResp.AccessToken,
		RefreshToken:          authResp.RefreshToken,
		AccessTokenExpiresAt:  timestamppb.New(authResp.AccessTokenExpiresAt.AsTime().UTC()),
		RefreshTokenExpiresAt: timestamppb.New(authResp.RefreshTokenExpiresAt.AsTime().UTC()),
	}, nil
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
		return &auth_client.ValidateTokenResponse{IsValid: false}, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &auth_client.ValidateTokenResponse{IsValid: true, UserId: claims.UserID}, nil
	}

	return &auth_client.ValidateTokenResponse{IsValid: false}, errors.New("invalid token claims")
}

// Logout invalidates a refresh token, effectively logging the user out.
// Logout vô hiệu hóa refresh token, đăng xuất người dùng.
func (s *authService) Logout(ctx context.Context, req *auth_client.LogoutRequest) (*auth_client.LogoutResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required")
	}

	// Xóa refresh token khỏi kho lưu trữ (repository) để vô hiệu hóa nó.
	err := s.refreshTokenRepo.Delete(ctx, req.GetRefreshToken())
	if err != nil {
		// Ngay cả khi token không được tìm thấy, chúng ta vẫn coi đây là một thành công,
		// vì người dùng đã không còn có thể sử dụng token đó nữa.
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			logger.Logger.Info("Logout success: refresh token not found but the goal is achieved.")
			return &auth_client.LogoutResponse{
				Message: "Logout successful",
			}, nil
		}
		// Trả về lỗi nếu có lỗi khác xảy ra trong quá trình xóa.
		return nil, status.Errorf(codes.Internal, "Failed to delete refresh token: %v", err)
	}

	logger.Logger.Info("Logout successful for refresh token", zap.String("token", req.GetRefreshToken()))
	return &auth_client.LogoutResponse{
		Message: "Logout successful",
	}, nil
}
