package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/handler"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/middleware"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/proxy"
)

func main() {
	log.Println("🚀 Starting API Gateway...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Println("✅ Configuration loaded")

	// Print config in development mode
	if cfg.IsDevelopment() {
		cfg.PrintConfig()
	}

	// Initialize gRPC clients
	grpcClients, err := clients.NewClients(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize gRPC clients: %v", err)
	}
	defer grpcClients.Close()

	// Initialize proxies
	userProxy := proxy.NewUserProxy(grpcClients.User)
	productProxy := proxy.NewProductProxy(grpcClients.Product)
	log.Println("✅ Proxies initialized")

	// Initialize handlers
	userHandler := handler.NewUserHandler(userProxy)
	productHandler := handler.NewProductHandler(productProxy)
	log.Println("✅ Handlers initialized")

	// Setup HTTP server
	router := setupRouter(cfg, userHandler, productHandler, userProxy)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 API Gateway listening on %s", cfg.GetServerAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	log.Println("✅ API Gateway is ready to handle requests")

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down API Gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ API Gateway stopped gracefully")
}

// setupRouter configures all routes and middleware
func setupRouter(cfg *config.Config, userHandler *handler.UserHandler, productHandler *handler.ProductHandler, userProxy *proxy.UserProxy) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.RefreshToken)
		}

		// User routes
		users := v1.Group("/users")
		{
			// Public routes
			users.GET("/:id", userHandler.GetUser)

			// Protected routes (require authentication)
			// users.Use(middleware.AuthMiddleware(userProxy))
			users.GET("/me", userHandler.GetProfile)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Product routes
		products := v1.Group("/products")
		{
			// Public routes - anyone can browse products
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)

			// Protected routes - require authentication
			// products.Use(middleware.AuthMiddleware(userProxy))
			products.POST("", productHandler.CreateProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		// Category routes
		categories := v1.Group("/categories")
		{
			// Public routes
			categories.GET("", productHandler.ListCategories)
			categories.GET("/:id", productHandler.GetCategory)

			// Protected routes
			// categories.Use(middleware.AuthMiddleware(userProxy))
			categories.POST("", productHandler.CreateCategory)
			categories.PUT("/:id", productHandler.UpdateCategory)
			categories.DELETE("/:id", productHandler.DeleteCategory)
		}

		// TODO: Add order, payment, inventory routes when ready
	}

	return router
}
