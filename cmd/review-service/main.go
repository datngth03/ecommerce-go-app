package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal" // Import for graceful shutdown
	"syscall"   // Import for graceful shutdown
	"time"      // To use time.Duration

	"github.com/joho/godotenv" // To read environment variables from .env file
	_ "github.com/lib/pq"      // PostgreSQL driver
	"go.uber.org/zap"          // Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For insecure gRPC connections (dev only)
	"google.golang.org/grpc/reflection"           // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/review/application"
	review_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/review/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/review/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"            // Add shared logger
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Product Service Client
	review_client "github.com/datngth03/ecommerce-go-app/pkg/client/review"   // Generated Review gRPC client
)

// main is the entry point for the Review Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Get gRPC port for Review Service
	grpcPort := os.Getenv("GRPC_PORT_REVIEW")
	if grpcPort == "" {
		grpcPort = "50060" // Default port for Review Service
		logger.Logger.Info("GRPC_PORT_REVIEW not set, using default", zap.String("port", grpcPort))
	}

	// Get Database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable not set.")
	}

	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
	}
	defer db.Close() // Ensure DB connection is closed on application exit

	// Set connection pool settings (optional but recommended for production)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Failed to ping PostgreSQL database", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for Review Service.")

	// Get Product Service address
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		logger.Logger.Info("PRODUCT_GRPC_ADDR not set, using default", zap.String("address", productSvcAddr))
	}

	// Connect to Product Service
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Product Service", zap.String("address", productSvcAddr), zap.Error(err))
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
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register ReviewGRPCServer with gRPC server
	review_client.RegisterReviewServiceServer(s, grpcServer)

	// Register reflection service (useful for gRPC client tools)
	reflection.Register(s)

	logger.Logger.Info("Review Service (gRPC) listening on port", zap.String("port", grpcPort))

	// Start gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Failed to serve gRPC server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Review Service gracefully...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Review Service stopped.")
}
