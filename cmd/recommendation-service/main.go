// cmd/recommendation-service/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time" // Để sử dụng time.Duration

	"github.com/joho/godotenv" // Để đọc biến môi trường từ file .env
	_ "github.com/lib/pq"      // PostgreSQL driver

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Cho kết nối gRPC không mã hóa (chỉ trong dev)
	"google.golang.org/grpc/reflection"           // Cho phép gRPC reflection (hữu ích cho công cụ client)

	"github.com/datngth03/ecommerce-go-app/internal/recommendation/application"
	recommendation_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/recommendation/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/recommendation/infrastructure/repository"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Client cho Product Service

	// Import mã gRPC đã tạo cho Recommendation Service
	recommendation_client "github.com/datngth03/ecommerce-go-app/pkg/client/recommendation"
)

// init loads environment variables from .env file.
func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not loaded, falling back to environment variables: %v", err)
	}
}

// main is the entry point for the Recommendation Service.
func main() {
	// Get gRPC port from environment variable, default to 50061
	grpcPort := os.Getenv("GRPC_PORT_RECOMMENDATION")
	if grpcPort == "" {
		grpcPort = "50061" // Cổng mặc định cho Recommendation Service
	}

	// Get Product Service gRPC address
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Địa chỉ mặc định của Product Service
	}

	// Connect to PostgreSQL database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL database: %v", err)
	}
	defer db.Close() // Đảm bảo đóng kết nối DB khi ứng dụng tắt

	// Ping database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping PostgreSQL database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL database for Recommendation Service.")

	// Set connection pool settings (optional but recommended for production)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize Repository
	interactionRepo := repository.NewPostgreSQLUserInteractionRepository(db)

	// Initialize Product Service gRPC client
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Product Service for Recommendation: %v", err)
	}
	defer productConn.Close()
	prodClient := product_client.NewProductServiceClient(productConn)

	// Initialize Application Service
	recommendationService := application.NewRecommendationService(interactionRepo, prodClient)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register RecommendationGRPCServer with the gRPC server
	recommendation_client.RegisterRecommendationServiceServer(s, recommendation_grpc_delivery.NewRecommendationGRPCServer(recommendationService))

	// Register reflection service. This allows gRPC client tools
	// to discover available services and methods without .proto files.
	reflection.Register(s)

	log.Printf("Recommendation Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
