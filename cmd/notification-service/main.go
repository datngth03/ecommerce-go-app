// cmd/notification-service/main.go
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

	"github.com/datngth03/ecommerce-go-app/internal/notification/application"
	notification_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/notification/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/notification/infrastructure/repository" // Import repository
	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification"     // Generated Notification gRPC client
)

// main is the entry point for the Notification Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT_NOTIFICATION"
	grpcPort := os.Getenv("GRPC_PORT_NOTIFICATION")
	if grpcPort == "" {
		grpcPort = "50058" // Default port for Notification Service
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
	log.Println("Successfully connected to PostgreSQL database for Notification Service.")

	// Initialize PostgreSQL Notification Repository
	notificationRepo := repository.NewPostgreSQLNotificationRepository(db)

	// Initialize Application Service
	notificationService := application.NewNotificationService(notificationRepo)

	// Initialize gRPC Server
	grpcServer := notification_grpc_delivery.NewNotificationGRPCServer(notificationService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register NotificationGRPCServer with the gRPC server
	notification_client.RegisterNotificationServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Notification Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

