// internal/api_gateway/delivery/http/router.go
package http

import (
	"github.com/datngth03/ecommerce-go-app/internal/api_gateway/middleware" // Import middleware
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the API Gateway.
// It takes a pointer to gin.Engine (the main router) and an instance of GatewayHandlers.
func RegisterRoutes(router *gin.Engine, handlers *GatewayHandlers) {
	// Khởi tạo Auth Middleware
	authMiddleware := middleware.NewAuthMiddleware(handlers.AuthClient)

	// Health check endpoint (public)
	router.GET("/health", handlers.HealthCheckHandler)

	// Public routes (không yêu cầu xác thực)
	publicGroup := router.Group("/api/v1")
	{
		// User registration (public)
		publicGroup.POST("/users/register", handlers.RegisterUser)
		// Auth login and token refresh (public)
		publicGroup.POST("/auth/login", handlers.Login)
		publicGroup.POST("/auth/refresh", handlers.RefreshAuthToken)
		// Public product/category listing (read-only)
		publicGroup.GET("/products", handlers.ListProducts)
		publicGroup.GET("/products/:id", handlers.GetProductById)
		publicGroup.GET("/categories", handlers.ListCategories)
		publicGroup.GET("/categories/:id", handlers.GetCategoryById)
	}

	// Authenticated routes (yêu cầu JWT token hợp lệ)
	authenticatedGroup := router.Group("/api/v1")
	authenticatedGroup.Use(authMiddleware.AuthRequired()) // Áp dụng middleware xác thực

	{
		// User Profile (authenticated)
		authenticatedGroup.GET("/users/:id", handlers.GetUserProfile) // Example: GET /api/v1/users/123
		// authenticatedGroup.PUT("/users/:id", handlers.UpdateUserProfile) // Add when implemented

		// Product Management (authenticated - e.g., for admin/seller)
		authenticatedGroup.POST("/products", handlers.CreateProduct)
		// authenticatedGroup.PUT("/products/:id", handlers.UpdateProduct)
		// authenticatedGroup.DELETE("/products/:id", handlers.DeleteProduct)

		// Category Management (authenticated - e.g., for admin/seller)
		authenticatedGroup.POST("/categories", handlers.CreateCategory)
		// authenticatedGroup.DELETE("/categories/:id", handlers.DeleteCategory)

		// Order Management (authenticated)
		authenticatedGroup.POST("/orders", handlers.CreateOrder)
		authenticatedGroup.GET("/orders/:id", handlers.GetOrderById)
		authenticatedGroup.PUT("/orders/:id/status", handlers.UpdateOrderStatus)
		authenticatedGroup.PUT("/orders/:id/cancel", handlers.CancelOrder)
		authenticatedGroup.GET("/orders", handlers.ListOrders)

		// Payment Management (authenticated)
		authenticatedGroup.POST("/payments", handlers.CreatePayment)
		authenticatedGroup.GET("/payments/:id", handlers.GetPaymentById)
		authenticatedGroup.POST("/payments/confirm", handlers.ConfirmPayment) // Webhook might bypass this, but for direct call
		authenticatedGroup.POST("/payments/refund", handlers.RefundPayment)
		authenticatedGroup.GET("/payments", handlers.ListPayments)

		// Cart Management (authenticated)
		authenticatedGroup.POST("/carts/add", handlers.AddItemToCart)
		authenticatedGroup.PUT("/carts/update-quantity", handlers.UpdateCartItemQuantity)
		authenticatedGroup.DELETE("/carts/remove", handlers.RemoveItemFromCart)
		authenticatedGroup.GET("/carts/:user_id", handlers.GetCart)            // Get cart for a specific user
		authenticatedGroup.DELETE("/carts/:user_id/clear", handlers.ClearCart) // Clear cart for a specific user

		// Shipping Management (authenticated)
		authenticatedGroup.POST("/shipping/calculate-cost", handlers.CalculateShippingCost)
		authenticatedGroup.POST("/shipping", handlers.CreateShipment)
		authenticatedGroup.GET("/shipping/:id", handlers.GetShipmentById)
		authenticatedGroup.PUT("/shipping/:id/status", handlers.UpdateShipmentStatus)
		authenticatedGroup.GET("/shipping/:id/track", handlers.TrackShipment)
		authenticatedGroup.GET("/shipping", handlers.ListShipments)

		// Auth token validation (for internal/testing purposes, might not be exposed directly in production)
		authenticatedGroup.POST("/auth/validate", handlers.ValidateAuthToken)
	}

	// Add other service routes here as you develop them
}
