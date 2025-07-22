// cmd/cart-service/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-redis/redis/v8" // Redis client
	"github.com/joho/godotenv"     // For .env file

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/cart/application"
	cart_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/cart/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/cart/infrastructure/repository" // Import Redis repository
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"             // Generated Cart gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"       // Product gRPC client
)

// main is the entry point for the Cart Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT_CART"
	grpcPort := os.Getenv("GRPC_PORT_CART")
	if grpcPort == "" {
		grpcPort = "50054" // Default port for Cart Service
	}

	// Get Redis address from environment variable "REDIS_ADDR"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address from Docker Compose
		log.Printf("No REDIS_ADDR environment variable found, using default: %s", redisAddr)
	}

	// Get Product Service gRPC address from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		log.Printf("No PRODUCT_GRPC_ADDR environment variable found, using default: %s", productSvcAddr)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // Use default DB
	})

	// Ping Redis to check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis for Cart Service.")

	// Initialize Product Service gRPC client
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service gRPC: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize Redis Cart Repository (e.g., with 24-hour TTL for carts)
	cartRepo := repository.NewRedisCartRepository(redisClient, 24*time.Hour)

	// Initialize Application Service (pass productClient to it)
	cartService := application.NewCartService(cartRepo, productClient)

	// Initialize gRPC Server
	grpcServer := cart_grpc_delivery.NewCartGRPCServer(cartService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register CartGRPCServer with the gRPC server
	cart_client.RegisterCartServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Cart Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
