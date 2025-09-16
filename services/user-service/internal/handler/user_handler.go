// internal/handler/user_handler.go
package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	grpcClient pb.UserServiceClient
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(grpcClient pb.UserServiceClient) *UserHandler {
	return &UserHandler{
		grpcClient: grpcClient,
	}
}

// HTTP Request/Response models
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Phone    string `json:"phone" binding:"omitempty,min=10,max=20"`
	Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Phone    *string `json:"phone,omitempty" binding:"omitempty,min=10,max=20"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// =================================
// CRUD Handlers
// =================================

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	log.Printf("CreateUser HTTP handler called")

	// 1. Parse and validate HTTP request
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// 2. Convert HTTP request to gRPC request
	grpcReq := &pb.CreateUserRequest{
		Email:    req.Email,
		Name:     req.Name,
		Phone:    req.Phone,
		Password: req.Password,
	}

	// 3. Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 4. Call gRPC service
	log.Printf("Calling gRPC CreateUser for email: %s", req.Email)
	grpcResp, err := h.grpcClient.CreateUser(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC CreateUser failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// 5. Handle gRPC response
	if !grpcResp.Success {
		statusCode := http.StatusBadRequest
		if grpcResp.Message == "User with this email already exists" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: grpcResp.Message,
		})
		return
	}

	// 6. Convert gRPC response to HTTP response
	userResponse := h.protoToUserResponse(grpcResp.User)

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: grpcResp.Message,
		Data:    userResponse,
	})
}

// GetUser handles GET /api/v1/users/:id and GET /api/v1/users/email/:email
func (h *UserHandler) GetUser(c *gin.Context) {
	log.Printf("GetUser HTTP handler called")

	var grpcReq *pb.GetUserRequest

	// Check if request is by ID or email
	if userID := c.Param("id"); userID != "" {
		// Get by ID
		id, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid user ID",
				Error:   err.Error(),
			})
			return
		}

		grpcReq = &pb.GetUserRequest{
			Identifier: &pb.GetUserRequest_Id{Id: id},
		}
	} else if email := c.Param("email"); email != "" {
		// Get by email
		grpcReq = &pb.GetUserRequest{
			Identifier: &pb.GetUserRequest_Email{Email: email},
		}
	} else {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Either user ID or email is required",
		})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.GetUser(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC GetUser failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get user",
			Error:   err.Error(),
		})
		return
	}

	// Handle gRPC response
	if !grpcResp.Success {
		statusCode := http.StatusNotFound
		if grpcResp.Message != "User not found" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: grpcResp.Message,
		})
		return
	}

	// Convert gRPC response to HTTP response
	userResponse := h.protoToUserResponse(grpcResp.User)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: grpcResp.Message,
		Data:    userResponse,
	})
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	log.Printf("UpdateUser HTTP handler called")

	// Get user ID from path
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	// Parse and validate HTTP request
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Convert HTTP request to gRPC request
	grpcReq := &pb.UpdateUserRequest{
		Id: userID,
	}

	// Set optional fields if provided
	if req.Name != nil {
		grpcReq.Name = req.Name
	}
	if req.Phone != nil {
		grpcReq.Phone = req.Phone
	}
	if req.IsActive != nil {
		grpcReq.IsActive = req.IsActive
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.UpdateUser(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC UpdateUser failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	// Handle gRPC response
	if !grpcResp.Success {
		statusCode := http.StatusNotFound
		if grpcResp.Message != "User not found" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: grpcResp.Message,
		})
		return
	}

	// Convert gRPC response to HTTP response
	userResponse := h.protoToUserResponse(grpcResp.User)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: grpcResp.Message,
		Data:    userResponse,
	})
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	log.Printf("DeleteUser HTTP handler called")

	// Get user ID from path
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	// Create gRPC request
	grpcReq := &pb.DeleteUserRequest{
		Id: userID,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call gRPC service
	grpcResp, err := h.grpcClient.DeleteUser(ctx, grpcReq)
	if err != nil {
		log.Printf("gRPC DeleteUser failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	// Handle gRPC response
	if !grpcResp.Success {
		statusCode := http.StatusNotFound
		if grpcResp.Message != "User not found" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Message: grpcResp.Message,
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: grpcResp.Message,
	})
}

// =================================
// Helper Methods
// =================================

// protoToUserResponse converts protobuf User to HTTP UserResponse
func (h *UserHandler) protoToUserResponse(pbUser *pb.User) *UserResponse {
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

// RegisterRoutes registers all user-related routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.POST("", h.CreateUser)          // POST /api/v1/users
		users.GET("/:id", h.GetUser)          // GET /api/v1/users/:id
		users.GET("/email/:email", h.GetUser) // GET /api/v1/users/email/:email
		users.PUT("/:id", h.UpdateUser)       // PUT /api/v1/users/:id
		users.DELETE("/:id", h.DeleteUser)    // DELETE /api/v1/users/:id
	}
}
