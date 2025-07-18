// cmd/product-service/main.go
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
	_ "github.com/lib/pq"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/product/application"
	product_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/product/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/repository" // Import gói repository mới
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"          // Mã gRPC đã tạo
)

// main is the entry point for the Product Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT"
	grpcPort := os.Getenv("GRPC_PORT_PRODUCT")
	if grpcPort == "" {
		grpcPort = "50052" // Default port for Product Service
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default connection string for PostgreSQL in Docker Compose
		// Assuming Product Service shares the same DB instance as User Service for simplicity in dev
		databaseURL = "postgres://user:password@localhost:5432/user_service_db?sslmode=disable"
		log.Printf("No DATABASE_URL environment variable found, using default: %s", databaseURL)
	}

	// Initialize PostgreSQL database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set connection pool parameters (optional but recommended)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping DB to check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database for Product Service.")

	// Initialize PostgreSQL Repositories
	productRepo := repository.NewPostgreSQLProductRepository(db)
	categoryRepo := repository.NewPostgreSQLCategoryRepository(db)

	// Initialize Application Service
	productService := application.NewProductService(productRepo, categoryRepo)

	// Initialize gRPC Server
	grpcServer := product_grpc_delivery.NewProductGRPCServer(productService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register ProductGRPCServer with the gRPC server
	product_client.RegisterProductServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Product Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
