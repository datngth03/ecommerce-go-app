// internal/api_gateway/delivery/http/router.go
package http

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all routes for the API Gateway.
// It takes a pointer to gin.Engine (the main router) and an instance of GatewayHandlers.
func RegisterRoutes(router *gin.Engine, handlers *GatewayHandlers) {
	// Health check endpoint
	router.GET("/health", handlers.HealthCheckHandler)

	// User Service Routes
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("/register", handlers.RegisterUser)
		userGroup.POST("/login", handlers.LoginUser)
		userGroup.GET("/:id", handlers.GetUserProfile) // Example: GET /api/v1/users/123
		// userGroup.PUT("/:id", handlers.UpdateUserProfile) // Add this when implementing UpdateUserProfile
	}

	// Product Service Routes
	productGroup := router.Group("/api/v1/products")
	{
		productGroup.POST("/", handlers.CreateProduct)
		productGroup.GET("/:id", handlers.GetProductById) // Example: GET /api/v1/products/abc
		productGroup.GET("/", handlers.ListProducts)      // Example: GET /api/v1/products?category_id=xyz&limit=10
		// productGroup.PUT("/:id", handlers.UpdateProduct) // Add this when implementing UpdateProduct
		// productGroup.DELETE("/:id", handlers.DeleteProduct) // Add this when implementing DeleteProduct
	}

	// Category Service Routes (part of Product Service)
	categoryGroup := router.Group("/api/v1/categories")
	{
		categoryGroup.POST("/", handlers.CreateCategory)
		categoryGroup.GET("/:id", handlers.GetCategoryById) // Example: GET /api/v1/categories/def
		categoryGroup.GET("/", handlers.ListCategories)
		// categoryGroup.DELETE("/:id", handlers.DeleteCategory) // Add this when implementing DeleteCategory
	}

	// Order Service Routes (placeholder for future integration)
	orderGroup := router.Group("/api/v1/orders")
	{
		orderGroup.POST("/", handlers.CreateOrder)
		orderGroup.GET("/:id", handlers.GetOrderById)
		orderGroup.PUT("/:id/status", handlers.UpdateOrderStatus) // Update status of an order
		orderGroup.PUT("/:id/cancel", handlers.CancelOrder)       // Cancel an order
		orderGroup.GET("/", handlers.ListOrders)                  // List orders
	}

	// Payment Service Routes
	paymentGroup := router.Group("/api/v1/payments")
	{
		paymentGroup.POST("/", handlers.CreatePayment)         // Create a new payment
		paymentGroup.GET("/:id", handlers.GetPaymentById)      // Get payment details by ID
		paymentGroup.POST("/confirm", handlers.ConfirmPayment) // Confirm payment (e.g., webhook callback)
		paymentGroup.POST("/refund", handlers.RefundPayment)   // Refund a payment
		paymentGroup.GET("/", handlers.ListPayments)           // List payments
	}

	// Cart Service Routes (ĐÃ THÊM PHẦN NÀY)
	cartGroup := router.Group("/api/v1/carts")
	{
		cartGroup.POST("/add", handlers.AddItemToCart)
		cartGroup.PUT("/update-quantity", handlers.UpdateCartItemQuantity)
		cartGroup.DELETE("/remove", handlers.RemoveItemFromCart)
		cartGroup.GET("/:user_id", handlers.GetCart)            // Get cart for a specific user
		cartGroup.DELETE("/:user_id/clear", handlers.ClearCart) // Clear cart for a specific user
	}

	// Shipping Service Routes
	shippingGroup := router.Group("/api/v1/shipping")
	{
		shippingGroup.POST("/calculate-cost", handlers.CalculateShippingCost) // Calculate shipping cost
		shippingGroup.POST("/", handlers.CreateShipment)                      // Create a new shipment
		shippingGroup.GET("/:id", handlers.GetShipmentById)                   // Get shipment details by ID
		shippingGroup.PUT("/:id/status", handlers.UpdateShipmentStatus)       // Update shipment status
		shippingGroup.GET("/:id/track", handlers.TrackShipment)               // Track shipment
		shippingGroup.GET("/", handlers.ListShipments)                        // List shipments
	}

	// Add other service routes here as you develop them
}
