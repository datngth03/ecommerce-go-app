// internal/api_gateway/delivery/http/router.go
package http

import (
	"github.com/datngth03/ecommerce-go-app/internal/api_gateway/middleware" // Import middleware
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the API Gateway.

func RegisterRoutes(router *gin.Engine, handlers *GatewayHandlers) {

	// Khởi tạo Auth Middleware
	authMiddleware := middleware.NewAuthMiddleware(handlers.AuthClient)

	// Public Routes (Không yêu cầu xác thực)
	public := router.Group("/api/v1")
	{
		public.GET("/health", handlers.HealthCheck)

		// User Public Routes (Đăng ký, không yêu cầu xác thực)
		public.POST("/users/register", handlers.RegisterUser)

		// Auth Public Routes (Đăng nhập, làm mới token)
		public.POST("/auth/login", handlers.Login)
		public.POST("/auth/refresh", handlers.RefreshToken)
		public.POST("/auth/validate", handlers.ValidateToken) // Có thể dùng cho client test token

		// Product Public Routes (Có thể xem sản phẩm mà không cần đăng nhập)
		public.GET("/products", handlers.ListProducts)
		public.GET("/products/:id", handlers.GetProductById)
		public.GET("/categories", handlers.ListCategories)
		public.GET("/categories/:id", handlers.GetCategoryById)

		// Review Public Routes (Có thể xem đánh giá mà không cần đăng nhập)
		public.GET("/reviews", handlers.ListAllReviews)                            // List all reviews with filters
		public.GET("/reviews/:id", handlers.GetReviewById)                         // Get review by ID
		public.GET("/reviews/products/:product_id", handlers.ListReviewsByProduct) // List reviews by product
	}

	// Authenticated Routes (Yêu cầu xác thực JWT)
	authenticated := router.Group("/api/v1")
	authenticated.Use(authMiddleware.AuthRequired()) //middleware AuthRequired
	{
		// User Authenticated Routes
		authenticated.GET("/users/:id", handlers.GetUserProfile) // Lấy hồ sơ người dùng

		// Product Authenticated Routes (Chỉ admin/người quản lý mới có quyền)
		authenticated.POST("/products", handlers.CreateProduct)
		authenticated.PUT("/products/:id", handlers.UpdateProduct)
		authenticated.DELETE("/products/:id", handlers.DeleteProduct)
		authenticated.POST("/categories", handlers.CreateCategory)
		// Product & Category GET methods are public, so not here

		// Order Routes (Yêu cầu xác thực)
		authenticated.POST("/orders", handlers.CreateOrder)
		authenticated.GET("/orders/:id", handlers.GetOrderById)
		authenticated.PUT("/orders/:id/status", handlers.UpdateOrderStatus)
		authenticated.POST("/orders/:id/cancel", handlers.CancelOrder)
		authenticated.GET("/orders", handlers.ListOrders)

		// Cart Routes (Yêu cầu xác thực)
		authenticated.POST("/carts/add", handlers.AddItemToCart)
		authenticated.GET("/carts/:userId", handlers.GetCart)
		authenticated.PUT("/carts/update-quantity", handlers.UpdateCartItemQuantity)
		authenticated.DELETE("/carts/remove", handlers.RemoveItemFromCart)
		authenticated.DELETE("/carts/:userId/clear", handlers.ClearCart)

		// Payment Routes (Yêu cầu xác thực)
		authenticated.POST("/payments", handlers.CreatePayment)
		authenticated.GET("/payments/:id", handlers.GetPaymentById)
		authenticated.POST("/payments/:id/confirm", handlers.ConfirmPayment)
		authenticated.POST("/payments/:id/refund", handlers.RefundPayment)
		authenticated.GET("/payments", handlers.ListPayments)

		// Shipping Routes (Yêu cầu xác thực)
		authenticated.POST("/shipping/calculate-cost", handlers.CalculateShippingCost)
		authenticated.POST("/shipping", handlers.CreateShipment)
		authenticated.GET("/shipping/:id", handlers.GetShipmentById)
		authenticated.PUT("/shipping/:id/status", handlers.UpdateShipmentStatus)
		authenticated.GET("/shipping/:id/track", handlers.TrackShipment)
		authenticated.GET("/shipping", handlers.ListShipments)

		// Notification Routes (Yêu cầu xác thực)
		authenticated.POST("/notifications/email", handlers.SendEmail)
		authenticated.POST("/notifications/sms", handlers.SendSMS)
		authenticated.POST("/notifications/push", handlers.SendPushNotification)

		// Inventory Routes (Yêu cầu xác thực, thường là admin hoặc internal) (THÊM PHẦN NÀY)
		authenticated.GET("/inventory/:id", handlers.GetStockQuantity) // Get stock by product ID
		authenticated.POST("/inventory/:id/set", handlers.SetStock)
		authenticated.POST("/inventory/:id/increase", handlers.IncreaseStock)
		authenticated.POST("/inventory/:id/decrease", handlers.DecreaseStock)
		authenticated.POST("/inventory/:id/reserve", handlers.ReserveStock)
		authenticated.POST("/inventory/:id/release", handlers.ReleaseStock)

		// Review Authenticated Routes (Để gửi, cập nhật, xóa đánh giá)
		authenticated.POST("/reviews", handlers.SubmitReview)        // Submit a new review
		authenticated.PUT("/reviews/{id}", handlers.UpdateReview)    // Update existing review
		authenticated.DELETE("/reviews/{id}", handlers.DeleteReview) // Delete a review
	}
}
