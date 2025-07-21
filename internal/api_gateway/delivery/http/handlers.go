// internal/api_gateway/delivery/http/handlers.go
package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"       // Generated User gRPC client
)

// GatewayHandlers holds the gRPC clients for backend services.
// GatewayHandlers chứa các gRPC client cho các dịch vụ backend.
type GatewayHandlers struct {
	UserClient    user_client.UserServiceClient
	ProductClient product_client.ProductServiceClient
	// Add other service clients here
}

// NewGatewayHandlers creates a new instance of GatewayHandlers.
// NewGatewayHandlers tạo một thể hiện mới của GatewayHandlers.
func NewGatewayHandlers(userSvcAddr, productSvcAddr string) (*GatewayHandlers, error) {
	// Connect to User Service
	userConn, err := grpc.Dial(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to User Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	// Connect to Product Service
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close() // Close userConn if productConn fails
		return nil, fmt.Errorf("failed to connect to Product Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	return &GatewayHandlers{
		UserClient:    user_client.NewUserServiceClient(userConn),
		ProductClient: product_client.NewProductServiceClient(productConn),
	}, nil
}

// healthCheckHandler is a simple handler to check the gateway's health.
// healthCheckHandler là một handler đơn giản để kiểm tra trạng thái của gateway.
func (h *GatewayHandlers) HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "API Gateway is running!",
	})
}

// --- User Service Handlers ---

// RegisterUser handles user registration requests.
// RegisterUser xử lý các yêu cầu đăng ký người dùng.
func (h *GatewayHandlers) RegisterUser(c *gin.Context) {
	var req user_client.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.UserClient.RegisterUser(ctx, &req)
	if err != nil {
		log.Printf("Error calling RegisterUser: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// LoginUser handles user login requests.
// LoginUser xử lý các yêu cầu đăng nhập người dùng.
func (h *GatewayHandlers) LoginUser(c *gin.Context) {
	var req user_client.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.UserClient.LoginUser(ctx, &req)
	if err != nil {
		log.Printf("Error calling LoginUser: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetUserProfile handles requests to get user profile.
// GetUserProfile xử lý các yêu cầu lấy hồ sơ người dùng.
func (h *GatewayHandlers) GetUserProfile(c *gin.Context) {
	userID := c.Param("id") // Lấy ID từ URL parameter

	req := &user_client.GetUserProfileRequest{
		UserId: userID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.UserClient.GetUserProfile(ctx, req)
	if err != nil {
		log.Printf("Error calling GetUserProfile: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Product Service Handlers ---

// CreateProduct handles requests to create a new product.
// CreateProduct xử lý các yêu cầu tạo sản phẩm mới.
func (h *GatewayHandlers) CreateProduct(c *gin.Context) {
	var req product_client.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ProductClient.CreateProduct(ctx, &req)
	if err != nil {
		log.Printf("Error calling CreateProduct: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetProductById handles requests to get product details by ID.
// GetProductById xử lý các yêu cầu lấy chi tiết sản phẩm theo ID.
func (h *GatewayHandlers) GetProductById(c *gin.Context) {
	productID := c.Param("id")

	req := &product_client.GetProductByIdRequest{
		Id: productID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ProductClient.GetProductById(ctx, req)
	if err != nil {
		log.Printf("Error calling GetProductById: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListProducts handles requests to list products.
// ListProducts xử lý các yêu cầu liệt kê sản phẩm.
func (h *GatewayHandlers) ListProducts(c *gin.Context) {
	var req product_client.ListProductsRequest
	// Bind query parameters for limit, offset, category_id
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ProductClient.ListProducts(ctx, &req)
	if err != nil {
		log.Printf("Error calling ListProducts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateCategory handles requests to create a new category.
// CreateCategory xử lý các yêu cầu tạo danh mục mới.
func (h *GatewayHandlers) CreateCategory(c *gin.Context) {
	var req product_client.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ProductClient.CreateCategory(ctx, &req)
	if err != nil {
		log.Printf("Error calling CreateCategory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetCategoryById handles requests to get category details by ID.
// GetCategoryById xử lý các yêu cầu lấy chi tiết danh mục theo ID.
func (h *GatewayHandlers) GetCategoryById(c *gin.Context) {
	categoryID := c.Param("id")

	req := &product_client.GetCategoryByIdRequest{
		Id: categoryID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ProductClient.GetCategoryById(ctx, req)
	if err != nil {
		log.Printf("Error calling GetCategoryById: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListCategories handles requests to list all categories.
// ListCategories xử lý các yêu cầu liệt kê tất cả các danh mục.
func (h *GatewayHandlers) ListCategories(c *gin.Context) {
	req := &product_client.ListCategoriesRequest{} // No specific fields for now

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ProductClient.ListCategories(ctx, req)
	if err != nil {
		log.Printf("Error calling ListCategories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
