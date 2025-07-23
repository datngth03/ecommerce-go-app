// cmd/payment-service/main.go
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
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to Order Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/payment/application"
	payment_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/payment/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/payment/infrastructure/repository" // Import repository
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"              // Order gRPC client
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment"          // Generated Payment gRPC client
)

// main is the entry point for the Payment Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT_PAYMENT"
	grpcPort := os.Getenv("GRPC_PORT_PAYMENT")
	if grpcPort == "" {
		grpcPort = "50055" // Default port for Payment Service
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Default connection string for PostgreSQL in Docker Compose
		databaseURL = "postgres://user:password@localhost:5432/ecommerce_core_db?sslmode=disable"
		log.Printf("No DATABASE_URL environment variable found, using default: %s", databaseURL)
	}

	// Get Order Service gRPC address from environment variable "ORDER_GRPC_ADDR"
	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053" // Default port for Order Service
		log.Printf("No ORDER_GRPC_ADDR environment variable found, using default: %s", orderSvcAddr)
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
	log.Println("Successfully connected to PostgreSQL database for Payment Service.")

	// Initialize Order Service gRPC client
	orderConn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Order Service gRPC: %v", err)
	}
	defer orderConn.Close()
	orderClient := order_client.NewOrderServiceClient(orderConn)

	// Initialize PostgreSQL Payment Repository
	paymentRepo := repository.NewPostgreSQLPaymentRepository(db)

	// Initialize Application Service (pass orderClient to it)
	paymentService := application.NewPaymentService(paymentRepo, orderClient)

	// Initialize gRPC Server
	grpcServer := payment_grpc_delivery.NewPaymentGRPCServer(paymentService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register PaymentGRPCServer with the gRPC server
	payment_client.RegisterPaymentServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Payment Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
