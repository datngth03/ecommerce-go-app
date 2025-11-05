// internal/handler/auth_handler.go
package handler

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/validator"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	grpcClient pb.UserServiceClient
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(grpcClient pb.UserServiceClient) *AuthHandler {
	return &AuthHandler{
		grpcClient: grpcClient,
	}
}

// HTTP Request/Response models for Auth
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    time.Time     `json:"expires_at"`
	User         *UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	ResetToken  string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// =================================
// Authentication Handlers
// =================================

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	log.Printf("Login HTTP handler called")

	// 1. Parse and validate HTTP request
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid login request: %v", err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate login credentials
	if err := validator.ValidateLoginRequest(req.Email, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// 2. Convert HTTP request to gRPC request
	grpcReq := &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// 3. Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 4. Call gRPC service
	log.Printf("Calling gRPC Login for email: %s", req.Email)
	grpcResp, err := h.grpcClient.Login(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC Login failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Login failed",
			Error:   err.Error(),
		})
		return
	}

	// 5. Handle gRPC response
	if !grpcResp.Success {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: grpcResp.Message,
		})
		return
	}

	// 6. Convert gRPC response to HTTP response
	userResponse := h.protoToUserResponse(grpcResp.User)
	loginResponse := &LoginResponse{
		AccessToken:  grpcResp.AccessToken,
		RefreshToken: grpcResp.RefreshToken,
		User:         userResponse,
	}

	if grpcResp.ExpiresAt != nil {
		loginResponse.ExpiresAt = grpcResp.ExpiresAt.AsTime()
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: grpcResp.Message,
		Data:    loginResponse,
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	log.Printf("RefreshToken HTTP handler called")

	// Parse request
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Convert to gRPC request
	grpcReq := &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.RefreshToken(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC RefreshToken failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Token refresh failed",
			Error:   err.Error(),
		})
		return
	}

	// Handle response
	if !grpcResp.Success {
		statusCode := http.StatusUnauthorized
		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: grpcResp.Message,
		})
		return
	}

	// Convert response
	userResponse := h.protoToUserResponse(grpcResp.User)
	loginResponse := &LoginResponse{
		AccessToken:  grpcResp.AccessToken,
		RefreshToken: grpcResp.RefreshToken,
		User:         userResponse,
	}

	if grpcResp.ExpiresAt != nil {
		loginResponse.ExpiresAt = grpcResp.ExpiresAt.AsTime()
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: grpcResp.Message,
		Data:    loginResponse,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	log.Printf("Logout HTTP handler called")

	// Get access token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Authorization header is required",
		})
		return
	}

	// Extract token (Bearer <token>)
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")
	if accessToken == authHeader {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid authorization header format",
		})
		return
	}

	// Get refresh token from request body (optional)
	var refreshToken *string
	type LogoutRequest struct {
		RefreshToken string `json:"refresh_token,omitempty"`
	}

	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
		refreshToken = &req.RefreshToken
	}

	// Convert to gRPC request
	grpcReq := &pb.LogoutRequest{
		AccessToken: accessToken,
	}
	if refreshToken != nil {
		grpcReq.RefreshToken = refreshToken
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.Logout(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC Logout failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Logout failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: grpcResp.Success,
		Message: grpcResp.Message,
	})
}

// =================================
// Password Management Handlers
// =================================

// ChangePassword handles POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	log.Printf("ChangePassword HTTP handler called")

	// Parse request
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Convert to gRPC request
	grpcReq := &pb.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	// Create context with timeout and add authorization
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Add authorization header to context metadata for gRPC
	ctx = h.addAuthToContext(ctx, c)

	// Call gRPC service
	grpcResp, err := h.grpcClient.ChangePassword(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC ChangePassword failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to change password",
			Error:   err.Error(),
		})
		return
	}

	// Handle response
	statusCode := http.StatusOK
	if !grpcResp.Success {
		if strings.Contains(grpcResp.Message, "incorrect") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(grpcResp.Message, "Authentication") {
			statusCode = http.StatusUnauthorized
		} else {
			statusCode = http.StatusBadRequest
		}
	}

	c.JSON(statusCode, APIResponse{
		Success: grpcResp.Success,
		Message: grpcResp.Message,
	})
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	log.Printf("ForgotPassword HTTP handler called")

	// Parse request
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Convert to gRPC request
	grpcReq := &pb.ForgotPasswordRequest{
		Email: req.Email,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.ForgotPassword(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC ForgotPassword failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to process forgot password request",
			Error:   err.Error(),
		})
		return
	}

	// Always return success for security (don't reveal if email exists)
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: grpcResp.Message,
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	log.Printf("ResetPassword HTTP handler called")

	// Parse request
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Convert to gRPC request
	grpcReq := &pb.ResetPasswordRequest{
		Email:       req.Email,
		ResetToken:  req.ResetToken,
		NewPassword: req.NewPassword,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.ResetPassword(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC ResetPassword failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to reset password",
			Error:   err.Error(),
		})
		return
	}

	// Handle response
	statusCode := http.StatusOK
	if !grpcResp.Success {
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, APIResponse{
		Success: grpcResp.Success,
		Message: grpcResp.Message,
	})
}

// =================================
// Helper Methods
// =================================

// protoToUserResponse converts protobuf User to HTTP UserResponse
func (h *AuthHandler) protoToUserResponse(pbUser *pb.User) *UserResponse {
	if pbUser == nil {
		return nil
	}

	userResponse := &UserResponse{
		ID:       pbUser.Id,
		Email:    pbUser.Email,
		Name:     pbUser.Name,
		Phone:    pbUser.Phone,
		IsActive: pbUser.IsActive,
	}

	// Convert timestamps
	if pbUser.CreatedAt != nil {
		userResponse.CreatedAt = pbUser.CreatedAt.AsTime()
	}
	if pbUser.UpdatedAt != nil {
		userResponse.UpdatedAt = pbUser.UpdatedAt.AsTime()
	}

	return userResponse
}

// addAuthToContext adds authorization header to gRPC context metadata
func (h *AuthHandler) addAuthToContext(ctx context.Context, c *gin.Context) context.Context {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Add to gRPC metadata - this would need proper metadata handling
		// For now, this is a placeholder
		// In practice, you'd use google.golang.org/grpc/metadata
		// md := metadata.Pairs("authorization", authHeader)
		// ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// RegisterRoutes registers all auth-related routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		// Authentication endpoints
		auth.POST("/login", h.Login)          // POST /api/v1/auth/login
		auth.POST("/refresh", h.RefreshToken) // POST /api/v1/auth/refresh
		auth.POST("/logout", h.Logout)        // POST /api/v1/auth/logout

		// Password management endpoints
		auth.POST("/change-password", h.ChangePassword) // POST /api/v1/auth/change-password
		auth.POST("/forgot-password", h.ForgotPassword) // POST /api/v1/auth/forgot-password
		auth.POST("/reset-password", h.ResetPassword)   // POST /api/v1/auth/reset-password
	}
}
