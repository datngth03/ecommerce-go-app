package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/handler"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/service"
)

func main() {
	log.Println("🚀 Starting Product Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Println("✅ Configuration loaded successfully")

	// Connect to PostgreSQL database
	db, err := repository.ConnectPostgres(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("⚠️  Error closing database: %v", err)
		}
	}()
	log.Println("✅ Connected to PostgreSQL database")

	log.Println("✅ Connected to PostgreSQL database")

	// Initialize repositories
	repos, err := repository.NewPostgresRepository(&repository.RepositoryOptions{
		Database: db,
	})
	if err != nil {
		log.Fatalf("❌ Failed to initialize repositories: %v", err)
	}
	log.Println("✅ Repositories initialized")

	// Initialize gRPC clients (optional - only if you need to call other services)
	var grpcClients *client.GRPCClients
	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr != "" {
		grpcClients, err = client.NewGRPCClients(userServiceAddr)
		if err != nil {
			log.Printf("⚠️  Failed to initialize gRPC clients: %v (continuing without user service)", err)
			grpcClients = nil
		} else {
			defer grpcClients.CloseAll()
			log.Println("✅ gRPC clients initialized")
		}
	} else {
		log.Println("ℹ️  USER_SERVICE_ADDR not set, skipping gRPC client initialization")
	}

	// Initialize services
	productService := service.NewProductService(repos)
	categoryService := service.NewCategoryService(repos)
	log.Println("✅ Services initialized")

	// Initialize handlers
	var userClient client.UserServiceClient
	if grpcClients != nil {
		userClient = grpcClients.UserClient
	}
	productHandler := handler.NewProductHandler(productService, userClient)
	categoryHandler := handler.NewCategoryHandler(categoryService, userClient)
	log.Println("✅ Handlers initialized")

	// Setup HTTP server
	router := setupHTTPServer(cfg, productHandler, categoryHandler)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Setup gRPC server
	grpcServer := setupGRPCServer(productService, categoryService)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("❌ Failed to listen on gRPC port: %v", err)
	}

	// Start HTTP server
	go func() {
		log.Printf("🚀 HTTP server starting on %s:%s", cfg.Server.Host, cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		log.Printf("🚀 gRPC server starting on %s:%s", cfg.Server.Host, cfg.Server.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("❌ gRPC server error: %v", err)
		}
	}()

	log.Println("✅ Product Service is ready to handle requests")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down servers...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  HTTP server shutdown error: %v", err)
	} else {
		log.Println("✅ HTTP server stopped gracefully")
	}

	// Shutdown gRPC server
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("✅ gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("⚠️  gRPC server shutdown timeout, forcing stop")
		grpcServer.Stop()
	}

	log.Println("👋 Product Service shutdown complete")
}

// setupHTTPServer configures the HTTP server with routes and middleware
func setupHTTPServer(cfg *config.Config, productHandler *handler.ProductHandler, categoryHandler *handler.CategoryHandler) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger()) // Built-in Gin logger

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "product-service",
			"version": "1.0.0",
		})
	})

	// Readiness check endpoint
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Product routes
		products := v1.Group("/products")
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.POST("", productHandler.CreateProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
			products.GET("/slug/:slug", productHandler.GetProductBySlug)
		}

		// Category routes
		categories := v1.Group("/categories")
		{
			categories.GET("", categoryHandler.ListCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.POST("", categoryHandler.CreateCategory)
			categories.PUT("/:id", categoryHandler.UpdateCategory)
			categories.DELETE("/:id", categoryHandler.DeleteCategory)
		}
	}

	return router
}

// setupGRPCServer configures the gRPC server with services
func setupGRPCServer(productService *service.ProductService, categoryService *service.CategoryService) *grpc.Server {
	// Create gRPC server with options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
		grpc.ConnectionTimeout(120 * time.Second),
	}

	grpcServer := grpc.NewServer(opts...)

	// Register gRPC services
	productRPCServer := rpc.NewProductGRPCServer(productService, categoryService)
	categoryRPCServer := rpc.NewCategoryGRPCServer(categoryService)

	pb.RegisterProductServiceServer(grpcServer, productRPCServer)
	pb.RegisterCategoryServiceServer(grpcServer, categoryRPCServer)

	// Enable reflection for grpcurl testing
	reflection.Register(grpcServer)

	return grpcServer
}
