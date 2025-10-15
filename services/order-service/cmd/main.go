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

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/events"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/handler"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/service"
)

func main() {
	log.Println("🚀 Starting Order Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}
	cfg.PrintConfig()

	// Connect to PostgreSQL
	db, err := repository.ConnectPostgres(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	// Connect to Redis
	redisClient, err := repository.ConnectRedis(
		fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✅ Connected to Redis")

	// Initialize RabbitMQ publisher
	var eventPublisher *events.Publisher
	if cfg.RabbitMQ.Enabled {
		rabbitMQURL := cfg.GetRabbitMQURL()
		eventPublisher, err = events.NewPublisher(rabbitMQURL)
		if err != nil {
			log.Printf("⚠️  Failed to initialize RabbitMQ publisher: %v (continuing without events)", err)
			eventPublisher = nil
		} else {
			defer eventPublisher.Close()
			log.Println("✅ Connected to RabbitMQ")
		}
	} else {
		log.Println("ℹ️  RabbitMQ not configured, skipping event publisher")
	}

	// Initialize gRPC clients for external services
	productClient, err := client.NewProductClient(cfg.Services.ProductService)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Product Service: %v", err)
	}
	defer productClient.Close()
	log.Println("✅ Connected to Product Service")

	userClient, err := client.NewUserClient(cfg.Services.UserService)
	if err != nil {
		log.Fatalf("❌ Failed to connect to User Service: %v", err)
	}
	defer userClient.Close()
	log.Println("✅ Connected to User Service")

	// Initialize repositories
	orderRepo := repository.NewOrderPostgresRepository(db)
	cartRepo := repository.NewCartPostgresRepository(db, redisClient)
	log.Println("✅ Repositories initialized")

	// Initialize services
	orderService := service.NewOrderService(orderRepo, cartRepo, productClient, userClient, eventPublisher)
	cartService := service.NewCartService(cartRepo, productClient)
	log.Println("✅ Services initialized")

	// Initialize HTTP handlers
	orderHandler := handler.NewOrderHandler(orderService)
	cartHandler := handler.NewCartHandler(cartService)
	log.Println("✅ HTTP handlers initialized")

	// Initialize gRPC server
	orderServer := rpc.NewOrderServer(orderService, cartService)
	log.Println("✅ gRPC server initialized")

	// Setup HTTP server (Gin)
	router := setupHTTPServer(cfg, orderHandler, cartHandler)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Setup gRPC server
	grpcServer := setupGRPCServer(orderServer)
	grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.GRPCPort))
	if err != nil {
		log.Fatalf("❌ Failed to listen on gRPC port: %v", err)
	}

	// Start HTTP server
	go func() {
		log.Printf("🌐 HTTP server starting on http://%s:%s", cfg.Server.Host, cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP server error: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		log.Printf("🔌 gRPC server starting on %s:%s", cfg.Server.Host, cfg.Server.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("❌ gRPC server error: %v", err)
		}
	}()

	log.Println("✅ Order Service is ready to handle requests")
	log.Printf("📝 HTTP endpoints: http://%s:%s", cfg.Server.Host, cfg.Server.HTTPPort)
	log.Printf("📝 gRPC endpoint: %s:%s", cfg.Server.Host, cfg.Server.GRPCPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down servers...")

	// Graceful shutdown
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

	log.Println("👋 Order Service shutdown complete")
}

func setupHTTPServer(cfg *config.Config, orderHandler *handler.OrderHandler, cartHandler *handler.CartHandler) *gin.Engine {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "order-service",
			"version": "1.0.0",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Order routes
		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
		}

		// Cart routes
		cart := v1.Group("/cart")
		{
			cart.GET("", cartHandler.GetCart)
			cart.POST("/items", cartHandler.AddToCart)
			cart.PATCH("/items/:product_id", cartHandler.UpdateCartItem)
			cart.DELETE("/items/:product_id", cartHandler.RemoveFromCart)
			cart.DELETE("", cartHandler.ClearCart)
		}
	}

	return router
}

func setupGRPCServer(orderServer *rpc.OrderServer) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Register services
	pb.RegisterOrderServiceServer(grpcServer, orderServer)

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	return grpcServer
}
