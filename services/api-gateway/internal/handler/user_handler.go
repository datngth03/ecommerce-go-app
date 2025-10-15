package handler

import (
	"net/http"
	"strconv"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	proxy *proxy.UserProxy
}

// NewUserHandler creates a new user handler
func NewUserHandler(proxy *proxy.UserProxy) *UserHandler {
	return &UserHandler{proxy: proxy}
}

// Register handles POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req pb.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.proxy.CreateUser(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

// Login handles POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req pb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.proxy.Login(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.proxy.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetUser handles GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.proxy.GetUser(c.Request.Context(), id)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// GetProfile handles GET /api/v1/users/me (current user profile)
func (h *UserHandler) GetProfile(c *gin.Context) {
	// User ID should be set by auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id format"})
		return
	}

	user, err := h.proxy.GetUser(c.Request.Context(), id)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req pb.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.Id = id

	user, err := h.proxy.UpdateUser(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.proxy.DeleteUser(c.Request.Context(), id); err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
