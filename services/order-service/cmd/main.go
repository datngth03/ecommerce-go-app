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

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/events"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/service"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq" // PostgreSQL driver
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Order Service v%s starting in %s mode...", cfg.Service.Version, cfg.Service.Environment)

	// 2. Initialize Database Connection
	db, err := repository.ConnectPostgres(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println(" PostgreSQL connection established")

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println(" Database connection closed")
		}
	}()

	// 3. Initialize Redis Connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println(" Redis connection established")

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		} else {
			log.Println(" Redis connection closed")
		}
	}()

	// 4. Initialize Repositories
	orderRepo := repository.NewOrderPostgresRepository(db)
	cartRepo := repository.NewCartPostgresRepository(db, redisClient)
	log.Println("✓ Repositories initialized")

	// 5. Initialize RabbitMQ Publisher
	publisher, err := events.NewPublisher(cfg.GetRabbitMQURL())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	log.Println("✓ RabbitMQ connection established")

	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Error closing RabbitMQ publisher: %v", err)
		} else {
			log.Println("✓ RabbitMQ publisher closed")
		}
	}()

	// 6. Initialize gRPC Clients
	userClient, err := client.NewUserClient(cfg.Services.UserService)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}
	log.Println("✓ User service client initialized")

	defer func() {
		if err := userClient.Close(); err != nil {
			log.Printf("Error closing user client: %v", err)
		} else {
			log.Println("✓ User service client closed")
		}
	}()

	productClient, err := client.NewProductClient(cfg.Services.ProductService)
	if err != nil {
		log.Fatalf("Failed to create product client: %v", err)
	}
	log.Println("✓ Product service client initialized")

	defer func() {
		if err := productClient.Close(); err != nil {
			log.Printf("Error closing product client: %v", err)
		} else {
			log.Println("✓ Product service client closed")
		}
	}()

	// 7. Initialize Services
	orderService := service.NewOrderService(orderRepo, cartRepo, productClient, userClient, publisher)
	cartService := service.NewCartService(cartRepo, productClient)
	log.Println("✓ Services initialized")

	// 6. Initialize gRPC Server
	grpcServer := grpc.NewServer()

	// Register Order Service
	orderGRPCServer := rpc.NewOrderServer(orderService, cartService)
	pb.RegisterOrderServiceServer(grpcServer, orderGRPCServer)

	// Register Health Check Service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("order_service.OrderService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	// 7. Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.Server.GRPCPort, err)
		}

		log.Printf(" Order gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 8. Start HTTP Health Check Server
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Server.HTTPPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"healthy","service":"order-service"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	}

	go func() {
		log.Printf(" HTTP health check server listening on port %s", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println(" Order Service is running. Press Ctrl+C to exit...")
	<-quit

	log.Println("Shutting down Order Service...")

	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println(" HTTP server stopped")

	// Stop gRPC server gracefully
	grpcServer.GracefulStop()
	log.Println(" gRPC server stopped")

	log.Println(" Order Service shutdown completed")
}
