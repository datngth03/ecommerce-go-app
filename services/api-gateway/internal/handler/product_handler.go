package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProductHandler handles HTTP requests for products
type ProductHandler struct {
	proxy *proxy.ProductProxy
}

// NewProductHandler creates a new product handler
func NewProductHandler(proxy *proxy.ProductProxy) *ProductHandler {
	return &ProductHandler{proxy: proxy}
}

// GetProduct handles GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
		return
	}

	product, err := h.proxy.GetProduct(c.Request.Context(), id)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

// ListProducts handles GET /api/v1/products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryID := c.Query("category_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	products, total, err := h.proxy.ListProducts(c.Request.Context(), int32(page), int32(pageSize), categoryID)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"products":    products,
			"total_count": total,
			"page":        page,
			"page_size":   pageSize,
		},
	})
}

// CreateProduct handles POST /api/v1/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req pb.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	product, err := h.proxy.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": product})
}

// UpdateProduct handles PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
		return
	}

	var req pb.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.Id = id

	product, err := h.proxy.UpdateProduct(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": product})
}

// DeleteProduct handles DELETE /api/v1/products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product id is required"})
		return
	}

	if err := h.proxy.DeleteProduct(c.Request.Context(), id); err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetCategory handles GET /api/v1/categories/:id
func (h *ProductHandler) GetCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id is required"})
		return
	}

	category, err := h.proxy.GetCategory(c.Request.Context(), id)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": category})
}

// ListCategories handles GET /api/v1/categories
func (h *ProductHandler) ListCategories(c *gin.Context) {
	categories, err := h.proxy.ListCategories(c.Request.Context())
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// CreateCategory handles POST /api/v1/categories
func (h *ProductHandler) CreateCategory(c *gin.Context) {
	var req pb.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	category, err := h.proxy.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": category})
}

// UpdateCategory handles PUT /api/v1/categories/:id
func (h *ProductHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id is required"})
		return
	}

	var req pb.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.Id = id

	category, err := h.proxy.UpdateCategory(c.Request.Context(), &req)
	if err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": category})
}

// DeleteCategory handles DELETE /api/v1/categories/:id
func (h *ProductHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category id is required"})
		return
	}

	if err := h.proxy.DeleteCategory(c.Request.Context(), id); err != nil {
		handleGRPCError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// handleGRPCError converts gRPC errors to HTTP responses
func handleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var httpStatus int
	switch st.Code() {
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.FailedPrecondition:
		httpStatus = http.StatusBadRequest
	default:
		httpStatus = http.StatusInternalServerError
	}

	c.JSON(httpStatus, gin.H{"error": st.Message()})
}

// MarshalJSON ensures proper JSON marshaling
func (h *ProductHandler) marshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
