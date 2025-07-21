// cmd/order-service/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to Product Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/order/application"
	order_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/order/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/order/infrastructure/repository" // Import gói repository mới
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"            // Generated Order gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"        // Product gRPC client
)

// main is the entry point for the Order Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT_ORDER"
	grpcPort := os.Getenv("GRPC_PORT_ORDER")
	if grpcPort == "" {
		grpcPort = "50053" // Default port for Order Service
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default connection string for PostgreSQL in Docker Compose
		databaseURL = "postgres://user:password@localhost:5432/user_service_db?sslmode=disable"
		log.Printf("No DATABASE_URL environment variable found, using default: %s", databaseURL)
	}

	// Get Product Service gRPC address from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		log.Printf("No PRODUCT_GRPC_ADDR environment variable found, using default: %s", productSvcAddr)
	}

	// Initialize PostgreSQL database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping DB to check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database for Order Service.")

	// Initialize Product Service gRPC client
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service gRPC: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize PostgreSQL Order Repository
	orderRepo := repository.NewPostgreSQLOrderRepository(db)

	// Initialize Application Service (pass productClient to it)
	orderService := application.NewOrderService(orderRepo, productClient)

	// Initialize gRPC Server
	grpcServer := order_grpc_delivery.NewOrderGRPCServer(orderService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register OrderGRPCServer with the gRPC server
	order_client.RegisterOrderServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Order Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
