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

	// Add other service routes here as you develop them
}
