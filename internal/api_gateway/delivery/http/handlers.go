// internal/api_gateway/delivery/http/handlers.go
package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	// "time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status" // Import status for gRPC error handling

	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth"                 // Generated Auth gRPC client
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"                 // Generated Cart gRPC client
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory"       // Generated Inventory gRPC client
	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification" // Generated Notification gRPC client
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"               // Generated Order gRPC client
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment"           // Generated Payment gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"           // Generated Product gRPC client
	review_client "github.com/datngth03/ecommerce-go-app/pkg/client/review"             // Generated Review gRPC client
	shipping_client "github.com/datngth03/ecommerce-go-app/pkg/client/shipping"         // Generated Shipping gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"                 // Generated User gRPC client
)

// GatewayHandlers holds the gRPC clients for backend services.
// GatewayHandlers chứa các gRPC client cho các dịch vụ backend.
type GatewayHandlers struct {
	UserClient         user_client.UserServiceClient
	ProductClient      product_client.ProductServiceClient
	OrderClient        order_client.OrderServiceClient
	PaymentClient      payment_client.PaymentServiceClient
	CartClient         cart_client.CartServiceClient
	ShippingClient     shipping_client.ShippingServiceClient
	AuthClient         auth_client.AuthServiceClient
	NotificationClient notification_client.NotificationServiceClient
	InventoryClient    inventory_client.InventoryServiceClient
	ReviewClient       review_client.ReviewServiceClient // THÊM: ReviewClient

	// gRPC connections to be closed
	userConn         *grpc.ClientConn
	productConn      *grpc.ClientConn
	orderConn        *grpc.ClientConn
	paymentConn      *grpc.ClientConn
	cartConn         *grpc.ClientConn
	shippingConn     *grpc.ClientConn
	authConn         *grpc.ClientConn
	notificationConn *grpc.ClientConn
	inventoryConn    *grpc.ClientConn
	reviewConn       *grpc.ClientConn // THÊM: reviewConn
}

// NewGatewayHandlers creates a new instance of GatewayHandlers.
// NewGatewayHandlers tạo một thể hiện mới của GatewayHandlers.
func NewGatewayHandlers(
	userSvcAddr, productSvcAddr, orderSvcAddr, paymentSvcAddr, cartSvcAddr,
	shippingSvcAddr, authSvcAddr, notificationSvcAddr, inventorySvcAddr, reviewSvcAddr string, // THÊM: reviewSvcAddr
) (*GatewayHandlers, error) {
	var err error
	var handlers GatewayHandlers

	// Connect to User Service
	handlers.userConn, err = grpc.Dial(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to User Service: %w", err)
	}
	handlers.UserClient = user_client.NewUserServiceClient(handlers.userConn)

	// Connect to Product Service
	handlers.productConn, err = grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Product Service: %w", err)
	}
	handlers.ProductClient = product_client.NewProductServiceClient(handlers.productConn)

	// Connect to Order Service
	handlers.orderConn, err = grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Order Service: %w", err)
	}
	handlers.OrderClient = order_client.NewOrderServiceClient(handlers.orderConn)

	// Connect to Payment Service
	handlers.paymentConn, err = grpc.Dial(paymentSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Payment Service: %w", err)
	}
	handlers.PaymentClient = payment_client.NewPaymentServiceClient(handlers.paymentConn)

	// Connect to Cart Service
	handlers.cartConn, err = grpc.Dial(cartSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cart Service: %w", err)
	}
	handlers.CartClient = cart_client.NewCartServiceClient(handlers.cartConn)

	// Connect to Shipping Service
	handlers.shippingConn, err = grpc.Dial(shippingSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Shipping Service: %w", err)
	}
	handlers.ShippingClient = shipping_client.NewShippingServiceClient(handlers.shippingConn)

	// Connect to Auth Service
	handlers.authConn, err = grpc.Dial(authSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Auth Service: %w", err)
	}
	handlers.AuthClient = auth_client.NewAuthServiceClient(handlers.authConn)

	// Connect to Notification Service
	handlers.notificationConn, err = grpc.Dial(notificationSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Notification Service: %w", err)
	}
	handlers.NotificationClient = notification_client.NewNotificationServiceClient(handlers.notificationConn)

	// Connect to Inventory Service
	handlers.inventoryConn, err = grpc.Dial(inventorySvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Inventory Service: %w", err)
	}
	handlers.InventoryClient = inventory_client.NewInventoryServiceClient(handlers.inventoryConn)

	// THÊM: Connect to Review Service
	handlers.reviewConn, err = grpc.Dial(reviewSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Review Service: %w", err)
	}
	handlers.ReviewClient = review_client.NewReviewServiceClient(handlers.reviewConn)

	return &handlers, nil
}

// CloseConnections closes all gRPC client connections.
// CloseConnections đóng tất cả các kết nối gRPC client.
func (h *GatewayHandlers) CloseConnections() {
	if h.userConn != nil {
		h.userConn.Close()
	}
	if h.productConn != nil {
		h.productConn.Close()
	}
	if h.orderConn != nil {
		h.orderConn.Close()
	}
	if h.paymentConn != nil {
		h.paymentConn.Close()
	}
	if h.cartConn != nil {
		h.cartConn.Close()
	}
	if h.shippingConn != nil {
		h.shippingConn.Close()
	}
	if h.authConn != nil {
		h.authConn.Close()
	}
	if h.notificationConn != nil {
		h.notificationConn.Close()
	}
	if h.inventoryConn != nil {
		h.inventoryConn.Close()
	}
	if h.reviewConn != nil { // THÊM: Đóng kết nối Review Service
		h.reviewConn.Close()
	}
	log.Println("Đã đóng tất cả các kết nối gRPC.")
}

// HealthCheck godoc
// @Summary Kiểm tra trạng thái API Gateway
// @Description Trả về trạng thái healthy nếu Gateway đang hoạt động
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *GatewayHandlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "API Gateway đang hoạt động!", "status": "healthy"})
}

// --- User Service Handlers ---

// RegisterUser godoc
// @Summary Đăng ký người dùng mới
// @Description Tạo một tài khoản người dùng mới
// @Tags User
// @Accept json
// @Produce json
// @Param user body user_client.RegisterUserRequest true "Thông tin đăng ký người dùng"
// @Success 200 {object} user_client.RegisterUserResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /users/register [post]
func (h *GatewayHandlers) RegisterUser(c *gin.Context) {
	var req user_client.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.UserClient.RegisterUser(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetUserProfile godoc
// @Summary Lấy hồ sơ người dùng
// @Description Lấy thông tin chi tiết hồ sơ của một người dùng theo ID
// @Tags User
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} user_client.UserProfileResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /users/{id} [get]
func (h *GatewayHandlers) GetUserProfile(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	resp, err := h.UserClient.GetUserProfile(context.Background(), &user_client.GetUserProfileRequest{UserId: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Product Service Handlers ---

// CreateProduct godoc
// @Summary Tạo sản phẩm mới
// @Description Tạo một sản phẩm mới trong hệ thống
// @Tags Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param product body product_client.CreateProductRequest true "Thông tin sản phẩm cần tạo"
// @Success 200 {object} product_client.ProductResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /products [post]
func (h *GatewayHandlers) CreateProduct(c *gin.Context) {
	var req product_client.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ProductClient.CreateProduct(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetProductById godoc
// @Summary Lấy chi tiết sản phẩm
// @Description Lấy thông tin chi tiết của một sản phẩm theo ID
// @Tags Product
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} product_client.ProductResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /products/{id} [get]
func (h *GatewayHandlers) GetProductById(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	resp, err := h.ProductClient.GetProductById(context.Background(), &product_client.GetProductByIdRequest{Id: productID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateProduct godoc
// @Summary Cập nhật sản phẩm
// @Description Cập nhật thông tin chi tiết của một sản phẩm
// @Tags Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body product_client.UpdateProductRequest true "Thông tin sản phẩm cần cập nhật"
// @Success 200 {object} product_client.ProductResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /products/{id} [put]
func (h *GatewayHandlers) UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req product_client.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = productID // Set ID from path parameter

	resp, err := h.ProductClient.UpdateProduct(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteProduct godoc
// @Summary Xóa sản phẩm
// @Description Xóa một sản phẩm theo ID
// @Tags Product
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} product_client.DeleteProductResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /products/{id} [delete]
func (h *GatewayHandlers) DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	resp, err := h.ProductClient.DeleteProduct(context.Background(), &product_client.DeleteProductRequest{Id: productID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListProducts godoc
// @Summary Liệt kê tất cả sản phẩm
// @Description Lấy danh sách tất cả sản phẩm, có thể lọc theo danh mục và phân trang
// @Tags Product
// @Produce json
// @Param category_id query string false "Lọc theo Category ID"
// @Param limit query int false "Số lượng bản ghi tối đa" default(10)
// @Param offset query int false "Số lượng bản ghi bỏ qua" default(0)
// @Success 200 {object} product_client.ListProductsResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /products [get]
func (h *GatewayHandlers) ListProducts(c *gin.Context) {
	categoryID := c.Query("category_id")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	resp, err := h.ProductClient.ListProducts(context.Background(), &product_client.ListProductsRequest{
		CategoryId: categoryID,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateCategory godoc
// @Summary Tạo danh mục mới
// @Description Tạo một danh mục sản phẩm mới
// @Tags Product
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param category body product_client.CreateCategoryRequest true "Thông tin danh mục cần tạo"
// @Success 200 {object} product_client.CategoryResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /categories [post]
func (h *GatewayHandlers) CreateCategory(c *gin.Context) {
	var req product_client.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ProductClient.CreateCategory(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCategoryById godoc
// @Summary Lấy chi tiết danh mục
// @Description Lấy thông tin chi tiết của một danh mục theo ID
// @Tags Product
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} product_client.CategoryResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /categories/{id} [get]
func (h *GatewayHandlers) GetCategoryById(c *gin.Context) {
	categoryID := c.Param("id")
	if categoryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
		return
	}

	resp, err := h.ProductClient.GetCategoryById(context.Background(), &product_client.GetCategoryByIdRequest{Id: categoryID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListCategories godoc
// @Summary Liệt kê tất cả danh mục
// @Description Lấy danh sách tất cả danh mục sản phẩm
// @Tags Product
// @Produce json
// @Success 200 {object} product_client.ListCategoriesResponse
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /categories [get]
func (h *GatewayHandlers) ListCategories(c *gin.Context) {
	resp, err := h.ProductClient.ListCategories(context.Background(), &product_client.ListCategoriesRequest{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Order Service Handlers ---

// CreateOrder godoc
// @Summary Tạo đơn hàng mới
// @Description Tạo một đơn hàng mới từ giỏ hàng hoặc các sản phẩm được chỉ định
// @Tags Order
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param order body order_client.CreateOrderRequest true "Thông tin đơn hàng cần tạo"
// @Success 200 {object} order_client.OrderResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /orders [post]
func (h *GatewayHandlers) CreateOrder(c *gin.Context) {
	var req order_client.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.OrderClient.CreateOrder(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetOrderById godoc
// @Summary Lấy chi tiết đơn hàng
// @Description Lấy thông tin chi tiết của một đơn hàng theo ID
// @Tags Order
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} order_client.OrderResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /orders/{id} [get]
func (h *GatewayHandlers) GetOrderById(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	resp, err := h.OrderClient.GetOrderById(context.Background(), &order_client.GetOrderByIdRequest{Id: orderID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus godoc
// @Summary Cập nhật trạng thái đơn hàng
// @Description Cập nhật trạng thái của một đơn hàng theo ID
// @Tags Order
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body order_client.UpdateOrderStatusRequest true "Thông tin trạng thái cần cập nhật"
// @Success 200 {object} order_client.OrderResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /orders/{id}/status [put]
func (h *GatewayHandlers) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	var req order_client.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.OrderId = orderID // Set ID from path parameter

	resp, err := h.OrderClient.UpdateOrderStatus(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CancelOrder godoc
// @Summary Hủy đơn hàng
// @Description Hủy một đơn hàng theo ID
// @Tags Order
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} order_client.OrderResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /orders/{id}/cancel [post]
func (h *GatewayHandlers) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	resp, err := h.OrderClient.CancelOrder(context.Background(), &order_client.CancelOrderRequest{OrderId: orderID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListOrders godoc
// @Summary Liệt kê đơn hàng
// @Description Lấy danh sách tất cả đơn hàng, có thể lọc theo người dùng và trạng thái
// @Tags Order
// @Security BearerAuth
// @Produce json
// @Param user_id query string false "Lọc theo User ID"
// @Param status query string false "Lọc theo trạng thái đơn hàng"
// @Param limit query int false "Số lượng bản ghi tối đa" default(10)
// @Param offset query int false "Số lượng bản ghi bỏ qua" default(0)
// @Success 200 {object} order_client.ListOrdersResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /orders [get]
func (h *GatewayHandlers) ListOrders(c *gin.Context) {
	userID := c.Query("user_id")
	statusFilter := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	resp, err := h.OrderClient.ListOrders(context.Background(), &order_client.ListOrdersRequest{
		UserId: userID,
		Status: statusFilter,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Cart Service Handlers ---

// AddItemToCart godoc
// @Summary Thêm sản phẩm vào giỏ hàng
// @Description Thêm một mặt hàng vào giỏ hàng của người dùng. Nếu sản phẩm đã có, sẽ tăng số lượng.
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cart_item body cart_client.AddItemToCartRequest true "Thông tin sản phẩm cần thêm vào giỏ hàng"
// @Success 200 {object} cart_client.CartResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /carts/add [post]
func (h *GatewayHandlers) AddItemToCart(c *gin.Context) {
	var req cart_client.AddItemToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.CartClient.AddItemToCart(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCart godoc
// @Summary Lấy giỏ hàng của người dùng
// @Description Lấy thông tin giỏ hàng hiện tại của một người dùng
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} cart_client.CartResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /carts/{userId} [get]
func (h *GatewayHandlers) GetCart(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	resp, err := h.CartClient.GetCart(context.Background(), &cart_client.GetCartRequest{UserId: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateCartItemQuantity godoc
// @Summary Cập nhật số lượng mặt hàng trong giỏ hàng
// @Description Cập nhật số lượng của một mặt hàng cụ thể trong giỏ hàng của người dùng
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param update_item body cart_client.UpdateCartItemQuantityRequest true "Thông tin cập nhật số lượng mặt hàng"
// @Success 200 {object} cart_client.CartResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /carts/update-quantity [put]
func (h *GatewayHandlers) UpdateCartItemQuantity(c *gin.Context) {
	var req cart_client.UpdateCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.CartClient.UpdateCartItemQuantity(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RemoveItemFromCart godoc
// @Summary Xóa mặt hàng khỏi giỏ hàng
// @Description Xóa một mặt hàng cụ thể khỏi giỏ hàng của người dùng
// @Tags Cart
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param remove_item body cart_client.RemoveItemFromCartRequest true "Thông tin mặt hàng cần xóa"
// @Success 200 {object} cart_client.CartResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /carts/remove [delete]
func (h *GatewayHandlers) RemoveItemFromCart(c *gin.Context) {
	var req cart_client.RemoveItemFromCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.CartClient.RemoveItemFromCart(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ClearCart godoc
// @Summary Xóa trắng giỏ hàng
// @Description Xóa tất cả các mặt hàng khỏi giỏ hàng của người dùng
// @Tags Cart
// @Security BearerAuth
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} cart_client.CartResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /carts/{userId}/clear [delete]
func (h *GatewayHandlers) ClearCart(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	resp, err := h.CartClient.ClearCart(context.Background(), &cart_client.ClearCartRequest{UserId: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Payment Service Handlers ---

// CreatePayment godoc
// @Summary Tạo giao dịch thanh toán mới
// @Description Bắt đầu một giao dịch thanh toán cho một đơn hàng
// @Tags Payment
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payment body payment_client.CreatePaymentRequest true "Thông tin thanh toán cần tạo"
// @Success 200 {object} payment_client.PaymentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /payments [post]
func (h *GatewayHandlers) CreatePayment(c *gin.Context) {
	var req payment_client.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.PaymentClient.CreatePayment(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetPaymentById godoc
// @Summary Lấy chi tiết giao dịch thanh toán
// @Description Lấy thông tin chi tiết của một giao dịch thanh toán theo ID
// @Tags Payment
// @Security BearerAuth
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} payment_client.PaymentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /payments/{id} [get]
func (h *GatewayHandlers) GetPaymentById(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	resp, err := h.PaymentClient.GetPaymentById(context.Background(), &payment_client.GetPaymentByIdRequest{Id: paymentID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ConfirmPayment godoc
// @Summary Xác nhận giao dịch thanh toán
// @Description Cập nhật trạng thái của một giao dịch thanh toán (ví dụ: từ pending sang completed)
// @Tags Payment
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param confirmation body payment_client.ConfirmPaymentRequest true "Thông tin xác nhận thanh toán"
// @Success 200 {object} payment_client.PaymentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /payments/{id}/confirm [post]
func (h *GatewayHandlers) ConfirmPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	var req payment_client.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PaymentId = paymentID // Set ID from path parameter

	resp, err := h.PaymentClient.ConfirmPayment(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RefundPayment godoc
// @Summary Hoàn tiền giao dịch
// @Description Hoàn tiền một giao dịch thanh toán đã hoàn tất
// @Tags Payment
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Param refund body payment_client.RefundPaymentRequest true "Thông tin hoàn tiền"
// @Success 200 {object} payment_client.PaymentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /payments/{id}/refund [post]
func (h *GatewayHandlers) RefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	var req payment_client.RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PaymentId = paymentID // Set ID from path parameter

	resp, err := h.PaymentClient.RefundPayment(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListPayments godoc
// @Summary Liệt kê các giao dịch thanh toán
// @Description Lấy danh sách tất cả các giao dịch thanh toán, có thể lọc theo người dùng và trạng thái
// @Tags Payment
// @Security BearerAuth
// @Produce json
// @Param user_id query string false "Lọc theo User ID"
// @Param order_id query string false "Lọc theo Order ID"
// @Param status query string false "Lọc theo trạng thái thanh toán"
// @Param limit query int false "Số lượng bản ghi tối đa" default(10)
// @Param offset query int false "Số lượng bản ghi bỏ qua" default(0)
// @Success 200 {object} payment_client.ListPaymentsResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /payments [get]
func (h *GatewayHandlers) ListPayments(c *gin.Context) {
	userID := c.Query("user_id")
	orderID := c.Query("order_id")
	statusFilter := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	resp, err := h.PaymentClient.ListPayments(context.Background(), &payment_client.ListPaymentsRequest{
		UserId:  userID,
		OrderId: orderID,
		Status:  statusFilter,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Shipping Service Handlers ---

// CalculateShippingCost godoc
// @Summary Tính toán chi phí vận chuyển
// @Description Tính toán chi phí vận chuyển ước tính cho một đơn hàng dựa trên địa chỉ
// @Tags Shipping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cost_request body shipping_client.CalculateShippingCostRequest true "Thông tin để tính toán chi phí vận chuyển"
// @Success 200 {object} shipping_client.ShippingCostResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /shipping/calculate-cost [post]
func (h *GatewayHandlers) CalculateShippingCost(c *gin.Context) {
	var req shipping_client.CalculateShippingCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ShippingClient.CalculateShippingCost(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateShipment godoc
// @Summary Tạo lô hàng mới
// @Description Tạo một lô hàng mới cho một đơn hàng đã đặt
// @Tags Shipping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param shipment body shipping_client.CreateShipmentRequest true "Thông tin lô hàng cần tạo"
// @Success 200 {object} shipping_client.ShipmentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /shipping [post]
func (h *GatewayHandlers) CreateShipment(c *gin.Context) {
	var req shipping_client.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ShippingClient.CreateShipment(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetShipmentById godoc
// @Summary Lấy chi tiết lô hàng
// @Description Lấy thông tin chi tiết của một lô hàng theo ID
// @Tags Shipping
// @Security BearerAuth
// @Produce json
// @Param id path string true "Shipment ID"
// @Success 200 {object} shipping_client.ShipmentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /shipping/{id} [get]
func (h *GatewayHandlers) GetShipmentById(c *gin.Context) {
	shipmentID := c.Param("id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shipment ID is required"})
		return
	}

	resp, err := h.ShippingClient.GetShipmentById(context.Background(), &shipping_client.GetShipmentByIdRequest{Id: shipmentID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateShipmentStatus godoc
// @Summary Cập nhật trạng thái lô hàng
// @Description Cập nhật trạng thái vận chuyển của một lô hàng theo ID
// @Tags Shipping
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Shipment ID"
// @Param status body shipping_client.UpdateShipmentStatusRequest true "Thông tin trạng thái cần cập nhật"
// @Success 200 {object} shipping_client.ShipmentResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /shipping/{id}/status [put]
func (h *GatewayHandlers) UpdateShipmentStatus(c *gin.Context) {
	shipmentID := c.Param("id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shipment ID is required"})
		return
	}

	var req shipping_client.UpdateShipmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ShipmentId = shipmentID // Set ID from path parameter

	resp, err := h.ShippingClient.UpdateShipmentStatus(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// TrackShipment godoc
// @Summary Theo dõi lô hàng
// @Description Lấy thông tin theo dõi trạng thái của một lô hàng theo ID
// @Tags Shipping
// @Security BearerAuth
// @Produce json
// @Param id path string true "Shipment ID"
// @Success 200 {object} shipping_client.TrackingResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /shipping/{id}/track [get]
func (h *GatewayHandlers) TrackShipment(c *gin.Context) {
	shipmentID := c.Param("id")
	if shipmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Shipment ID is required"})
		return
	}

	resp, err := h.ShippingClient.TrackShipment(context.Background(), &shipping_client.TrackShipmentRequest{ShipmentId: shipmentID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListShipments godoc
// @Summary Liệt kê các lô hàng
// @Description Lấy danh sách tất cả các lô hàng, có thể lọc theo người dùng hoặc đơn hàng và phân trang
// @Tags Shipping
// @Security BearerAuth
// @Produce json
// @Param user_id query string false "Lọc theo User ID"
// @Param order_id query string false "Lọc theo Order ID"
// @Param status query string false "Lọc theo trạng thái lô hàng"
// @Param limit query int false "Số lượng bản ghi tối đa" default(10)
// @Param offset query int false "Số lượng bản ghi bỏ qua" default(0)
// @Success 200 {object} shipping_client.ListShipmentsResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /shipping [get]
func (h *GatewayHandlers) ListShipments(c *gin.Context) {
	userID := c.Query("user_id")
	orderID := c.Query("order_id")
	statusFilter := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	resp, err := h.ShippingClient.ListShipments(context.Background(), &shipping_client.ListShipmentsRequest{
		UserId:  userID,
		OrderId: orderID,
		Status:  statusFilter,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Auth Service Handlers ---

// Login godoc
// @Summary Đăng nhập người dùng
// @Description Xác thực người dùng bằng email và mật khẩu, trả về Access Token và Refresh Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body auth_client.LoginRequest true "Thông tin đăng nhập người dùng"
// @Success 200 {object} auth_client.AuthResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /auth/login [post]
func (h *GatewayHandlers) Login(c *gin.Context) {
	var req auth_client.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.AuthClient.Login(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 || st.Code() == 16 { // gRPC Not Found or Unauthenticated
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RefreshToken godoc
// @Summary Làm mới Access Token
// @Description Sử dụng Refresh Token để lấy Access Token mới
// @Tags Auth
// @Accept json
// @Produce json
// @Param token body auth_client.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} auth_client.AuthResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /auth/refresh [post]
func (h *GatewayHandlers) RefreshToken(c *gin.Context) {
	var req auth_client.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.AuthClient.RefreshToken(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 16 { // gRPC Unauthenticated
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ValidateToken godoc
// @Summary Xác thực Access Token
// @Description Xác thực một Access Token để kiểm tra tính hợp lệ
// @Tags Auth
// @Accept json
// @Produce json
// @Param token body auth_client.ValidateTokenRequest true "Access Token"
// @Success 200 {object} auth_client.ValidateTokenResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /auth/validate [post]
func (h *GatewayHandlers) ValidateToken(c *gin.Context) {
	var req auth_client.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.AuthClient.ValidateToken(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 16 { // gRPC Unauthenticated
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired access token"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Notification Service Handlers ---

// SendEmail godoc
// @Summary Gửi email thông báo
// @Description Gửi một email thông báo đến người dùng
// @Tags Notification
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param notification body notification_client.SendEmailRequest true "Thông tin email cần gửi"
// @Success 200 {object} notification_client.SendNotificationResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /notifications/email [post]
func (h *GatewayHandlers) SendEmail(c *gin.Context) {
	var req notification_client.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.NotificationClient.SendEmail(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendSMS godoc
// @Summary Gửi SMS thông báo
// @Description Gửi một tin nhắn SMS thông báo đến người dùng
// @Tags Notification
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param notification body notification_client.SendSMSRequest true "Thông tin SMS cần gửi"
// @Success 200 {object} notification_client.SendNotificationResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /notifications/sms [post]
func (h *GatewayHandlers) SendSMS(c *gin.Context) {
	var req notification_client.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.NotificationClient.SendSMS(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendPushNotification godoc
// @Summary Gửi thông báo đẩy (Push Notification)
// @Description Gửi một thông báo đẩy đến thiết bị của người dùng
// @Tags Notification
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param notification body notification_client.SendPushNotificationRequest true "Thông tin Push Notification cần gửi"
// @Success 200 {object} notification_client.SendNotificationResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /notifications/push [post]
func (h *GatewayHandlers) SendPushNotification(c *gin.Context) {
	var req notification_client.SendPushNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.NotificationClient.SendPushNotification(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Inventory Service Handlers ---

// GetStockQuantity godoc
// @Summary Lấy số lượng tồn kho sản phẩm
// @Description Lấy số lượng tồn kho hiện tại và số lượng đặt trước của một sản phẩm
// @Tags Inventory
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} inventory_client.StockQuantityResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /inventory/{id} [get]
func (h *GatewayHandlers) GetStockQuantity(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	resp, err := h.InventoryClient.GetStockQuantity(context.Background(), &inventory_client.GetStockQuantityRequest{ProductId: productID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SetStock godoc
// @Summary Đặt số lượng tồn kho ban đầu
// @Description Thiết lập số lượng tồn kho ban đầu cho một sản phẩm. Chỉ dành cho admin.
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param stock body inventory_client.SetStockRequest true "Số lượng tồn kho cần đặt"
// @Success 200 {object} inventory_client.StockQuantityResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /inventory/{id}/set [post]
func (h *GatewayHandlers) SetStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req inventory_client.SetStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductId = productID // Set ID from path parameter

	resp, err := h.InventoryClient.SetStock(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// IncreaseStock godoc
// @Summary Tăng số lượng tồn kho
// @Description Tăng số lượng tồn kho của một sản phẩm.
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param stock body inventory_client.IncreaseStockRequest true "Số lượng tồn kho cần tăng"
// @Success 200 {object} inventory_client.StockQuantityResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /inventory/{id}/increase [post]
func (h *GatewayHandlers) IncreaseStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req inventory_client.IncreaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductId = productID // Set ID from path parameter

	resp, err := h.InventoryClient.IncreaseStock(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DecreaseStock godoc
// @Summary Giảm số lượng tồn kho
// @Description Giảm số lượng tồn kho của một sản phẩm.
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param stock body inventory_client.DecreaseStockRequest true "Số lượng tồn kho cần giảm"
// @Success 200 {object} inventory_client.StockQuantityResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /inventory/{id}/decrease [post]
func (h *GatewayHandlers) DecreaseStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req inventory_client.DecreaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductId = productID // Set ID from path parameter

	resp, err := h.InventoryClient.DecreaseStock(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReserveStock godoc
// @Summary Đặt trước số lượng tồn kho
// @Description Đặt trước một số lượng sản phẩm trong tồn kho, làm giảm số lượng khả dụng nhưng không giảm tổng tồn kho.
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param reservation body inventory_client.ReserveStockRequest true "Số lượng tồn kho cần đặt trước"
// @Success 200 {object} inventory_client.StockQuantityResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /inventory/{id}/reserve [post]
func (h *GatewayHandlers) ReserveStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req inventory_client.ReserveStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductId = productID // Set ID from path parameter

	resp, err := h.InventoryClient.ReserveStock(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReleaseStock godoc
// @Summary Giải phóng số lượng tồn kho đã đặt trước
// @Description Giải phóng một số lượng sản phẩm đã được đặt trước, tăng số lượng khả dụng.
// @Tags Inventory
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param release body inventory_client.ReleaseStockRequest true "Số lượng tồn kho cần giải phóng"
// @Success 200 {object} inventory_client.StockQuantityResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /inventory/{id}/release [post]
func (h *GatewayHandlers) ReleaseStock(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	var req inventory_client.ReleaseStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductId = productID // Set ID from path parameter

	resp, err := h.InventoryClient.ReleaseStock(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Review Service Handlers (THÊM PHẦN NÀY) ---

// SubmitReview godoc
// @Summary Gửi đánh giá sản phẩm mới
// @Description Cho phép người dùng gửi một đánh giá cho một sản phẩm
// @Tags Review
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param review body review_client.SubmitReviewRequest true "Thông tin đánh giá sản phẩm"
// @Success 200 {object} review_client.ReviewResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 409 {object} map[string]string "Conflict (Review already exists)"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /reviews [post]
func (h *GatewayHandlers) SubmitReview(c *gin.Context) {
	var req review_client.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.ReviewClient.SubmitReview(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 6 { // Already Exists (gRPC code for conflict)
				c.JSON(http.StatusConflict, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetReviewById godoc
// @Summary Lấy đánh giá theo ID
// @Description Lấy thông tin chi tiết của một đánh giá theo ID
// @Tags Review
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} review_client.ReviewResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /reviews/{id} [get]
func (h *GatewayHandlers) GetReviewById(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review ID is required"})
		return
	}

	resp, err := h.ReviewClient.GetReviewById(context.Background(), &review_client.GetReviewByIdRequest{Id: reviewID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateReview godoc
// @Summary Cập nhật đánh giá
// @Description Cập nhật rating và/hoặc comment của một đánh giá hiện có
// @Tags Review
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Param review body review_client.UpdateReviewRequest true "Thông tin đánh giá cần cập nhật"
// @Success 200 {object} review_client.ReviewResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /reviews/{id} [put]
func (h *GatewayHandlers) UpdateReview(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review ID is required"})
		return
	}

	var req review_client.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Id = reviewID // Set ID from path parameter

	resp, err := h.ReviewClient.UpdateReview(context.Background(), &req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteReview godoc
// @Summary Xóa đánh giá
// @Description Xóa một đánh giá dựa trên ID
// @Tags Review
// @Security BearerAuth
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} review_client.DeleteReviewResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /reviews/{id} [delete]
func (h *GatewayHandlers) DeleteReview(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review ID is required"})
		return
	}

	resp, err := h.ReviewClient.DeleteReview(context.Background(), &review_client.DeleteReviewRequest{Id: reviewID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == 7 { // gRPC Not Found
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListReviewsByProduct godoc
// @Summary Liệt kê đánh giá theo sản phẩm
// @Description Lấy danh sách đánh giá cho một sản phẩm cụ thể
// @Tags Review
// @Produce json
// @Param product_id path string true "Product ID"
// @Param limit query int false "Số lượng bản ghi tối đa" default(10)
// @Param offset query int false "Số lượng bản ghi bỏ qua" default(0)
// @Success 200 {object} review_client.ListReviewsResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /reviews/products/{product_id} [get]
func (h *GatewayHandlers) ListReviewsByProduct(c *gin.Context) {
	productID := c.Param("product_id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	resp, err := h.ReviewClient.ListReviewsByProduct(context.Background(), &review_client.ListReviewsByProductRequest{
		ProductId: productID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListAllReviews godoc
// @Summary Liệt kê tất cả đánh giá
// @Description Lấy danh sách tất cả đánh giá, có thể lọc theo sản phẩm, người dùng, rating tối thiểu và phân trang
// @Tags Review
// @Produce json
// @Param product_id query string false "Lọc theo Product ID"
// @Param user_id query string false "Lọc theo User ID"
// @Param min_rating query int false "Lọc theo Rating tối thiểu (1-5)"
// @Param limit query int false "Số lượng bản ghi tối đa" default(10)
// @Param offset query int false "Số lượng bản ghi bỏ qua" default(0)
// @Success 200 {object} review_client.ListReviewsResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /reviews [get]
func (h *GatewayHandlers) ListAllReviews(c *gin.Context) {
	productID := c.Query("product_id")
	userID := c.Query("user_id")
	minRatingStr := c.Query("min_rating")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}
	offset, err := strconv.ParseInt(offsetStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	var minRating int32
	if minRatingStr != "" {
		val, err := strconv.ParseInt(minRatingStr, 10, 32)
		if err != nil || val < 1 || val > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_rating parameter (must be 1-5)"})
			return
		}
		minRating = int32(val)
	}

	resp, err := h.ReviewClient.ListAllReviews(context.Background(), &review_client.ListAllReviewsRequest{
		ProductId: productID,
		UserId:    userID,
		MinRating: minRating,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}
