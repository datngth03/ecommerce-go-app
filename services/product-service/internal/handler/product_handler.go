package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/service"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	service    *service.ProductService
	userClient client.UserServiceClient
}

func NewProductHandler(service *service.ProductService, userClient client.UserServiceClient) *ProductHandler {
	return &ProductHandler{
		service:    service,
		userClient: userClient,
	}
}

// RegisterRoutes đăng ký tất cả các route cho product
func (h *ProductHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/products")
	{
		// public routes
		group.GET("", h.ListProducts)
		group.GET("/:id", h.GetProduct)
		group.GET("/slug/:slug", h.GetProductBySlug)

		// protected routes
		group.POST("", h.CreateProduct)
		group.PUT("/:id", h.UpdateProduct)
		group.DELETE("/:id", h.DeleteProduct)
		group.POST("/:id/activate", h.ActivateProduct)
		group.POST("/:id/deactivate", h.DeactivateProduct)
	}

	// Route riêng để lấy sản phẩm theo category
	categoryGroup := router.Group("/api/v1/categories")
	{
		categoryGroup.GET("/:id/products", h.ListProductsByCategory)
	}
}

// Helper: get user from token
func (h *ProductHandler) getUserFromToken(c *gin.Context) (*client.UserInfo, error) {
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

// Helper: Kiểm tra quyền admin
func (h *ProductHandler) requireAdmin(c *gin.Context) (*client.UserInfo, bool) {
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

// CreateProduct godoc
// @Summary      Create a new product
// @Description  Create a new product with the input payload
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        product  body      models.CreateProductRequest  true  "Create Product"
// @Success      201      {object}  models.ProductResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {

	userInfo, ok := h.requireAdmin(c)
	if !ok {
		return
	}

	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Product created successfully",
		"product":    product,
		"created_by": userInfo.Username,
	})
}

// GetProduct godoc
// @Summary      Get a product by ID
// @Description  Get a product's details by its ID
// @Tags         Products
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  models.ProductResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	product, err := h.service.GetProduct(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetProductBySlug godoc
// @Summary Get a product by slug
// @Description Get a product's details by its slug
// @Tags Products
// @Produce json
// @Param slug path string true "Product Slug"
// @Success 200 {object} models.ProductResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products/slug/{slug} [get]
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.service.GetProductBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Update a product's details by its ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id       path      string                       true  "Product ID"
// @Param        product  body      models.UpdateProductRequest  true  "Update Product"
// @Success      200      {object}  models.ProductResponse
// @Failure      400      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {

	userInfo, ok := h.requireAdmin(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Product updated successfully",
		"product":    product,
		"updated_by": userInfo.Username,
	})
	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Delete a product by its ID
// @Tags         Products
// @Param        id   path      string  true  "Product ID"
// @Success      204  "No Content"
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteProduct(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product: " + err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListProducts godoc
// @Summary      List products
// @Description  Get a list of products with pagination and filtering
// @Tags         Products
// @Produce      json
// @Param        page        query     int     false  "Page number"
// @Param        pageSize    query     int     false  "Number of items per page"
// @Param        categoryId  query     string  false  "Filter by Category ID"
// @Success      200         {object}  models.ListProductsResponse
// @Failure      400         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	var req models.ListProductsRequest
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	req.CategoryID = c.Query("categoryId")

	response, err := h.service.ListProducts(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "category not found") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListProductsByCategory godoc
// @Summary      List products by category
// @Description  Get a list of products for a specific category with pagination
// @Tags         Products
// @Produce      json
// @Param        id          path      string  true   "Category ID"
// @Param        page        query     int     false  "Page number"
// @Param        pageSize    query     int     false  "Number of items per page"
// @Success      200         {object}  models.ListProductsResponse
// @Failure      400         {object}  map[string]string
// @Failure      404         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Router       /categories/{id}/products [get]
func (h *ProductHandler) ListProductsByCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var req models.ListProductsRequest
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	response, err := h.service.ListProductsByCategory(c.Request.Context(), categoryID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "category not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products by category: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ActivateProduct godoc
// @Summary Activate a product
// @Description Activate a product by its ID
// @Tags Products
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products/{id}/activate [post]
func (h *ProductHandler) ActivateProduct(c *gin.Context) {
	id := c.Param("id")
	err := h.service.ActivateProduct(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "already active") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product activated successfully"})
}

// DeactivateProduct godoc
// @Summary Deactivate a product
// @Description Deactivate a product by its ID
// @Tags Products
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /products/{id}/deactivate [post]
func (h *ProductHandler) DeactivateProduct(c *gin.Context) {
	id := c.Param("id")
	err := h.service.DeactivateProduct(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "already inactive") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deactivated successfully"})
}
