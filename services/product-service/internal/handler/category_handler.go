package handler

import (
	"net/http"
	"strings"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/service"
	"github.com/gin-gonic/gin"
)

// CategoryHandler xử lý các request HTTP liên quan đến category.
type CategoryHandler struct {
	service    *service.CategoryService
	userClient client.UserServiceClient
}

// NewCategoryHandler khởi tạo một category handler mới.
func NewCategoryHandler(service *service.CategoryService, userClient client.UserServiceClient) *CategoryHandler {
	return &CategoryHandler{
		service:    service,
		userClient: userClient,
	}
}

// RegisterRoutes đăng ký tất cả các route cho category.
func (h *CategoryHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/categories")
	{
		// public routes
		group.GET("", h.ListCategories)

		// protected routes
		group.POST("", h.CreateCategory)
		group.GET("/:id", h.GetCategory)
		group.GET("/slug/:slug", h.GetCategoryBySlug)
		group.PUT("/:id", h.UpdateCategory)
		group.DELETE("/:id", h.DeleteCategory)
	}
}

// Helper: get user from token
func (h *CategoryHandler) getUserFromToken(c *gin.Context) (*client.UserInfo, error) {
	if h.userClient == nil {
		return nil, nil
	}
	// Lấy token từ header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, nil
	}

	token := parts[1]

	// Validate token với User Service
	userInfo, err := h.userClient.ValidateToken(c.Request.Context(), token)
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (h *CategoryHandler) requireAdmin(c *gin.Context) (*client.UserInfo, bool) {
	userInfo, err := h.getUserFromToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return nil, false
	}

	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return nil, false
	}

	if userInfo.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return nil, false
	}

	return userInfo, true
}

// CreateCategory godoc
// @Summary      Create a new category
// @Description  Create a new category with the input payload
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        category  body      models.CreateCategoryRequest  true  "Create Category"
// @Success      201       {object}  models.CategoryResponse
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /categories [post]

func (h *CategoryHandler) CreateCategory(c *gin.Context) {

	userInfo, ok := h.requireAdmin(c)
	if !ok {
		return
	}

	var req models.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Category created successfully",
		"product":    category,
		"created_by": userInfo.Username,
	})
}

// GetCategory godoc
// @Summary      Get a category by ID
// @Description  Get a category's details by its ID
// @Tags         Categories
// @Produce      json
// @Param        id   path      string  true  "Category ID"
// @Success      200  {object}  models.CategoryResponse
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /categories/{id} [get]
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")

	category, err := h.service.GetCategory(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// GetCategoryBySlug godoc
// @Summary      Get a category by slug
// @Description  Get a category's details by its slug
// @Tags         Categories
// @Produce      json
// @Param        slug   path      string  true  "Category Slug"
// @Success      200    {object}  models.CategoryResponse
// @Failure      404    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /categories/slug/{slug} [get]
func (h *CategoryHandler) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	category, err := h.service.GetCategoryBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get category by slug: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// UpdateCategory godoc
// @Summary      Update a category
// @Description  Update a category's details by its ID
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Category ID"
// @Param        category body      models.UpdateCategoryRequest  true  "Update Category"
// @Success      200      {object}  models.CategoryResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	category, err := h.service.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary      Delete a category
// @Description  Delete a category by its ID
// @Tags         Categories
// @Param        id   path      string  true  "Category ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "cannot delete category") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListCategories godoc
// @Summary      List all categories
// @Description  Get a list of all categories
// @Tags         Categories
// @Produce      json
// @Success      200  {object}  models.ListCategoriesResponse
// @Failure      500  {object}  map[string]string
// @Router       /categories [get]
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	response, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list categories: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
