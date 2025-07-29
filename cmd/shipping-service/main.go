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

	"github.com/datngth03/ecommerce-go-app/internal/shared/logger" // Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shipping/application"
	shipping_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/shipping/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/shipping/infrastructure/repository"
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"       // Order Service Client
	shipping_client "github.com/datngth03/ecommerce-go-app/pkg/client/shipping" // Generated Shipping gRPC client
)

// main is the entry point for the Shipping Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found or failed to load.", zap.Error(err))
	}

	// Get gRPC port from environment variable "GRPC_PORT_SHIPPING"
	grpcPort := os.Getenv("GRPC_PORT_SHIPPING")
	if grpcPort == "" {
		grpcPort = "50056" // Default port for Shipping Service
		logger.Logger.Info("GRPC_PORT_SHIPPING not set, using default.", zap.String("port", grpcPort))
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable is not set.")
	}

	// Get Order Service gRPC address from environment variable "ORDER_GRPC_ADDR"
	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053" // Default port for Order Service
		logger.Logger.Info("ORDER_GRPC_ADDR not set, using default.", zap.String("address", orderSvcAddr))
	}

	// Initialize PostgreSQL database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database.", zap.Error(err))
	}
	defer db.Close()

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping DB to check connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Failed to ping PostgreSQL database.", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for Shipping Service.")

	// Initialize Order Service gRPC client
	orderConn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Order Service gRPC.", zap.String("address", orderSvcAddr), zap.Error(err))
	}
	defer orderConn.Close()
	orderClient := order_client.NewOrderServiceClient(orderConn)

	// Initialize PostgreSQL Shipment Repository
	shipmentRepo := repository.NewPostgreSQLShipmentRepository(db)

	// Initialize Application Service (pass orderClient to it)
	shippingService := application.NewShippingService(shipmentRepo, orderClient)

	// Initialize gRPC Server
	grpcServer := shipping_grpc_delivery.NewShippingGRPCServer(shippingService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on port.", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register ShippingGRPCServer with the gRPC server
	shipping_client.RegisterShippingServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Shipping Service (gRPC) listening on port.", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Failed to serve gRPC server.", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Listen for interrupt and terminate signals
	<-quit                                               // Block until a signal is received

	logger.Logger.Info("Shutting down Shipping Service gracefully...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Shipping Service stopped.")
}
