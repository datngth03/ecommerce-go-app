// cmd/review-service/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"database/sql"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/datngth03/ecommerce-go-app/internal/review/application"
	review_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/review/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/review/infrastructure/repository"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Import Product gRPC client
	review_client "github.com/datngth03/ecommerce-go-app/pkg/client/review"   // Generated Review gRPC client
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port for Review Service
	grpcPort := os.Getenv("GRPC_PORT_REVIEW")
	if grpcPort == "" {
		grpcPort = "50060" // Default port for Review Service
	}

	// Get Database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatalf("DATABASE_URL environment variable not set.")
	}

	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}
	defer db.Close()

	// Ping database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database for Review Service.")

	// Get Product Service address
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
	}

	// Connect to Product Service
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service: %v", err)
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize Repository
	reviewRepo := repository.NewPostgreSQLReviewRepository(db)

	// Initialize Application Service
	reviewService := application.NewReviewService(reviewRepo, productClient)

	// Initialize gRPC Server
	grpcServer := review_grpc_delivery.NewReviewGRPCServer(reviewService)

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register ReviewGRPCServer with gRPC server
	review_client.RegisterReviewServiceServer(s, grpcServer)

	// Register reflection service (useful for gRPC client tools)
	reflection.Register(s)

	log.Printf("Review Service (gRPC) listening on port :%s...", grpcPort)

	// Start gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
