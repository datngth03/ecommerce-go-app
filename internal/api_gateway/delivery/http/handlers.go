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

	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth"                 // Generated Auth gRPC client
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"                 // Generated Cart gRPC client
	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification" // Generated Notification gRPC client
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"               // Generated Order gRPC client
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment"           // Generated Payment gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"           // Generated Product gRPC client
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
	NotificationClient notification_client.NotificationServiceClient // Add Notification Service client
	// Add other service clients here
}

// NewGatewayHandlers creates a new instance of GatewayHandlers.
// NewGatewayHandlers tạo một thể hiện mới của GatewayHandlers.
func NewGatewayHandlers(userSvcAddr, productSvcAddr, orderSvcAddr, paymentSvcAddr, cartSvcAddr, shippingSvcAddr, authSvcAddr, notificationSvcAddr string) (*GatewayHandlers, error) {
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

	// Connect to Order Service
	orderConn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		productConn.Close()
		return nil, fmt.Errorf("failed to connect to Order Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	// Connect to Payment Service
	paymentConn, err := grpc.Dial(paymentSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		productConn.Close()
		orderConn.Close()
		return nil, fmt.Errorf("failed to connect to Payment Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	// Connect to Cart Service
	cartConn, err := grpc.Dial(cartSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		productConn.Close()
		orderConn.Close()
		paymentConn.Close()
		return nil, fmt.Errorf("failed to connect to Cart Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	// Connect to Shipping Service
	shippingConn, err := grpc.Dial(shippingSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		productConn.Close()
		orderConn.Close()
		paymentConn.Close()
		cartConn.Close()
		return nil, fmt.Errorf("failed to connect to Shipping Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	// Connect to Auth Service
	authConn, err := grpc.Dial(authSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		productConn.Close()
		orderConn.Close()
		paymentConn.Close()
		cartConn.Close()
		shippingConn.Close()
		return nil, fmt.Errorf("failed to connect to Auth Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	// Connect to Notification Service (THÊM PHẦN NÀY)
	notificationConn, err := grpc.Dial(notificationSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		userConn.Close()
		productConn.Close()
		orderConn.Close()
		paymentConn.Close()
		cartConn.Close()
		shippingConn.Close()
		authConn.Close()
		return nil, fmt.Errorf("failed to connect to Notification Service: %w", err)
	}
	// No defer conn.Close() here, connections will be closed in main.go

	return &GatewayHandlers{
		UserClient:         user_client.NewUserServiceClient(userConn),
		ProductClient:      product_client.NewProductServiceClient(productConn),
		OrderClient:        order_client.NewOrderServiceClient(orderConn),
		PaymentClient:      payment_client.NewPaymentServiceClient(paymentConn),
		CartClient:         cart_client.NewCartServiceClient(cartConn),
		ShippingClient:     shipping_client.NewShippingServiceClient(shippingConn),
		AuthClient:         auth_client.NewAuthServiceClient(authConn),
		NotificationClient: notification_client.NewNotificationServiceClient(notificationConn), // THÊM DÒNG NÀY
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

// // Login handles user login requests via Auth Service. (ĐÃ SỬA TÊN HÀM VÀ LOGIC)
// func (h *GatewayHandlers) Login(c *gin.Context) {
// 	var req auth_client.LoginRequest // Sử dụng LoginRequest từ auth_client
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout for auth
// 	defer cancel()

// 	// Call Auth Service's Login RPC
// 	authResp, err := h.AuthClient.Login(ctx, &req)
// 	if err != nil {
// 		log.Printf("Error calling Auth Service Login: %v", err)
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or login failed"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, authResp) // Return tokens and user ID
// }

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
// ListCategories xử lý việc liệt kê tất cả các danh mục.
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

// --- Order Service Handlers ---

// CreateOrder handles creating a new order.
func (h *GatewayHandlers) CreateOrder(c *gin.Context) {
	var req order_client.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout for order creation
	defer cancel()

	resp, err := h.OrderClient.CreateOrder(ctx, &req)
	if err != nil {
		log.Printf("Error calling CreateOrder: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetOrderById handles retrieving order details by ID.
func (h *GatewayHandlers) GetOrderById(c *gin.Context) {
	orderID := c.Param("id")

	req := &order_client.GetOrderByIdRequest{
		Id: orderID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.OrderClient.GetOrderById(ctx, req)
	if err != nil {
		log.Printf("Error calling GetOrderById: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateOrderStatus handles updating an order's status.
func (h *GatewayHandlers) UpdateOrderStatus(c *gin.Context) {
	var req order_client.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.OrderClient.UpdateOrderStatus(ctx, &req)
	if err != nil {
		log.Printf("Error calling UpdateOrderStatus: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CancelOrder handles cancelling an order.
func (h *GatewayHandlers) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	req := &order_client.CancelOrderRequest{
		OrderId: orderID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.OrderClient.CancelOrder(ctx, req)
	if err != nil {
		log.Printf("Error calling CancelOrder: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListOrders handles listing orders.
func (h *GatewayHandlers) ListOrders(c *gin.Context) {
	var req order_client.ListOrdersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.OrderClient.ListOrders(ctx, &req)
	if err != nil {
		log.Printf("Error calling ListOrders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Payment Service Handlers ---

// CreatePayment handles requests to initiate a new payment.
func (h *GatewayHandlers) CreatePayment(c *gin.Context) {
	var req payment_client.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout for payment
	defer cancel()

	resp, err := h.PaymentClient.CreatePayment(ctx, &req)
	if err != nil {
		log.Printf("Error calling CreatePayment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetPaymentById handles requests to get payment details by ID.
func (h *GatewayHandlers) GetPaymentById(c *gin.Context) {
	paymentID := c.Param("id")

	req := &payment_client.GetPaymentByIdRequest{
		Id: paymentID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.PaymentClient.GetPaymentById(ctx, req)
	if err != nil {
		log.Printf("Error calling GetPaymentById: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ConfirmPayment handles requests to confirm a payment (e.g., from a webhook).
func (h *GatewayHandlers) ConfirmPayment(c *gin.Context) {
	var req payment_client.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout
	defer cancel()

	resp, err := h.PaymentClient.ConfirmPayment(ctx, &req)
	if err != nil {
		log.Printf("Error calling ConfirmPayment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RefundPayment handles requests to refund a payment.
func (h *GatewayHandlers) RefundPayment(c *gin.Context) {
	var req payment_client.RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout
	defer cancel()

	resp, err := h.PaymentClient.RefundPayment(ctx, &req)
	if err != nil {
		log.Printf("Error calling RefundPayment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListPayments handles requests to list payments.
func (h *GatewayHandlers) ListPayments(c *gin.Context) {
	var req payment_client.ListPaymentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.PaymentClient.ListPayments(ctx, &req)
	if err != nil {
		log.Printf("Error calling ListPayments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Cart Service Handlers ---

// AddItemToCart handles adding an item to the user's cart.
func (h *GatewayHandlers) AddItemToCart(c *gin.Context) {
	var req cart_client.AddItemToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.CartClient.AddItemToCart(ctx, &req)
	if err != nil {
		log.Printf("Error calling AddItemToCart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateCartItemQuantity handles updating the quantity of an item in the cart.
func (h *GatewayHandlers) UpdateCartItemQuantity(c *gin.Context) {
	var req cart_client.UpdateCartItemQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.CartClient.UpdateCartItemQuantity(ctx, &req)
	if err != nil {
		log.Printf("Error calling UpdateCartItemQuantity: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RemoveItemFromCart handles removing an item from the cart.
func (h *GatewayHandlers) RemoveItemFromCart(c *gin.Context) {
	var req cart_client.RemoveItemFromCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.CartClient.RemoveItemFromCart(ctx, &req)
	if err != nil {
		log.Printf("Error calling RemoveItemFromCart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCart handles retrieving the current state of the user's cart.
func (h *GatewayHandlers) GetCart(c *gin.Context) {
	userID := c.Param("user_id") // Assuming user_id is passed as a path parameter

	req := &cart_client.GetCartRequest{
		UserId: userID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.CartClient.GetCart(ctx, req)
	if err != nil {
		log.Printf("Error calling GetCart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ClearCart handles clearing all items from the user's cart.
func (h *GatewayHandlers) ClearCart(c *gin.Context) {
	userID := c.Param("user_id") // Assuming user_id is passed as a path parameter

	req := &cart_client.ClearCartRequest{
		UserId: userID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.CartClient.ClearCart(ctx, req)
	if err != nil {
		log.Printf("Error calling ClearCart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Shipping Service Handlers ---

// CalculateShippingCost handles requests to calculate shipping cost.
func (h *GatewayHandlers) CalculateShippingCost(c *gin.Context) {
	var req shipping_client.CalculateShippingCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ShippingClient.CalculateShippingCost(ctx, &req)
	if err != nil {
		log.Printf("Error calling CalculateShippingCost: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateShipment handles requests to create a new shipment.
func (h *GatewayHandlers) CreateShipment(c *gin.Context) {
	var req shipping_client.CreateShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout for shipment creation
	defer cancel()

	resp, err := h.ShippingClient.CreateShipment(ctx, &req)
	if err != nil {
		log.Printf("Error calling CreateShipment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetShipmentById handles requests to get shipment details by ID.
func (h *GatewayHandlers) GetShipmentById(c *gin.Context) {
	shipmentID := c.Param("id")

	req := &shipping_client.GetShipmentByIdRequest{
		Id: shipmentID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ShippingClient.GetShipmentById(ctx, req)
	if err != nil {
		log.Printf("Error calling GetShipmentById: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateShipmentStatus handles requests to update shipment status.
func (h *GatewayHandlers) UpdateShipmentStatus(c *gin.Context) {
	var req shipping_client.UpdateShipmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ShippingClient.UpdateShipmentStatus(ctx, &req)
	if err != nil {
		log.Printf("Error calling UpdateShipmentStatus: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// TrackShipment handles requests to track a shipment.
func (h *GatewayHandlers) TrackShipment(c *gin.Context) {
	shipmentID := c.Param("id")

	req := &shipping_client.TrackShipmentRequest{
		ShipmentId: shipmentID,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ShippingClient.TrackShipment(ctx, req)
	if err != nil {
		log.Printf("Error calling TrackShipment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListShipments handles requests to list shipments.
func (h *GatewayHandlers) ListShipments(c *gin.Context) {
	var req shipping_client.ListShipmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.ShippingClient.ListShipments(ctx, &req)
	if err != nil {
		log.Printf("Error calling ListShipments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Auth Service Handlers ---

// Login handles user login requests via Auth Service.
func (h *GatewayHandlers) Login(c *gin.Context) {
	var req auth_client.LoginRequest // Sử dụng LoginRequest từ auth_client
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout for auth
	defer cancel()

	// Call Auth Service's Login RPC
	authResp, err := h.AuthClient.Login(ctx, &req)
	if err != nil {
		log.Printf("Error calling Auth Service Login: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or login failed"})
		return
	}

	c.JSON(http.StatusOK, authResp) // Return tokens and user ID
}

// RefreshAuthToken handles requests to refresh an access token using a refresh token.
func (h *GatewayHandlers) RefreshAuthToken(c *gin.Context) {
	var req auth_client.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.AuthClient.RefreshToken(ctx, &req)
	if err != nil {
		log.Printf("Error refreshing token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ValidateAuthToken handles requests to validate an access token.
// This might be used by a middleware or for direct testing.
func (h *GatewayHandlers) ValidateAuthToken(c *gin.Context) {
	var req auth_client.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.AuthClient.ValidateToken(ctx, &req)
	if err != nil {
		log.Printf("Error validating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- Notification Service Handlers (THÊM PHẦN NÀY) ---

// SendEmail handles sending an email notification.
func (h *GatewayHandlers) SendEmail(c *gin.Context) {
	var req notification_client.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.NotificationClient.SendEmail(ctx, &req)
	if err != nil {
		log.Printf("Error calling SendEmail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendSMS handles sending an SMS notification.
func (h *GatewayHandlers) SendSMS(c *gin.Context) {
	var req notification_client.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.NotificationClient.SendSMS(ctx, &req)
	if err != nil {
		log.Printf("Error calling SendSMS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendPushNotification handles sending a push notification.
func (h *GatewayHandlers) SendPushNotification(c *gin.Context) {
	var req notification_client.SendPushNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := h.NotificationClient.SendPushNotification(ctx, &req)
	if err != nil {
		log.Printf("Error calling SendPushNotification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
