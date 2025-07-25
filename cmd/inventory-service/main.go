// cmd/inventory-service/main.go
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
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/inventory/application"
	inventory_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/inventory/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/inventory/infrastructure/repository" // Import repository
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory"        // Generated Inventory gRPC client
)

// main is the entry point for the Inventory Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT_INVENTORY"
	grpcPort := os.Getenv("GRPC_PORT_INVENTORY")
	if grpcPort == "" {
		grpcPort = "50059" // Default port for Inventory Service
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default connection string for PostgreSQL in Docker Compose
		databaseURL = "postgres://user:password@localhost:5432/ecommerce_core_db?sslmode=disable"
		log.Printf("No DATABASE_URL environment variable found, using default: %s", databaseURL)
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
	log.Println("Successfully connected to PostgreSQL database for Inventory Service.")

	// Initialize PostgreSQL Inventory Repository
	inventoryRepo := repository.NewPostgreSQLInventoryRepository(db)

	// Initialize Application Service
	inventoryService := application.NewInventoryService(inventoryRepo)

	// Initialize gRPC Server
	grpcServer := inventory_grpc_delivery.NewInventoryGRPCServer(inventoryService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register InventoryGRPCServer with the gRPC server
	inventory_client.RegisterInventoryServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Inventory Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
