// cmd/inventory-service/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http" // THÊM: Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"   // For graceful shutdown
	"time"

	"github.com/joho/godotenv"                                                             // For .env file
	_ "github.com/lib/pq"                                                                  // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"            // For gRPC client connections
	"google.golang.org/grpc/reflection" // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/inventory/application"
	inv_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/inventory/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/inventory/infrastructure/messaging" // Import messaging package
	"github.com/datngth03/ecommerce-go-app/internal/inventory/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                // THÊM: Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"               // THÊM: Add shared tracing
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Generated Inventory gRPC client
	// product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"     // Product gRPC client for Kafka Consumer
)

// main is the entry point for the Inventory Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "inventory-service"

	// Init TracerProvider for OpenTelemetry
	jaegerCollectorURL := os.Getenv("JAEGER_COLLECTOR_URL")
	if jaegerCollectorURL == "" {
		jaegerCollectorURL = "http://localhost:14268/api/traces" // Default Jaeger collector URL
		logger.Logger.Info("JAEGER_COLLECTOR_URL not set, using default.", zap.String("address", jaegerCollectorURL))
	}

	tp, err := tracing.InitTracerProvider(context.Background(), serviceName, jaegerCollectorURL)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize TracerProvider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	// Get gRPC port from environment variable, default to 50059
	grpcPort := os.Getenv("GRPC_PORT_INVENTORY")
	if grpcPort == "" {
		grpcPort = "50059"
		logger.Logger.Info("GRPC_PORT_INVENTORY not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9099
	metricsPort := os.Getenv("METRICS_PORT_INVENTORY")
	if metricsPort == "" {
		metricsPort = "9109"
		logger.Logger.Info("METRICS_PORT_INVENTORY not set, using default.", zap.String("port", metricsPort))
	}

	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable is not set.")
	}

	// Get Kafka broker address from environment variable
	kafkaBroker := os.Getenv("KAFKA_BROKER_ADDR")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092" // Default Kafka broker address
		logger.Logger.Info("KAFKA_BROKER_ADDR not set, using default.", zap.String("address", kafkaBroker))
	}
	productEventsTopic := os.Getenv("KAFKA_PRODUCT_EVENTS_TOPIC")
	if productEventsTopic == "" {
		productEventsTopic = "product_events" // Default Kafka topic for product events
		logger.Logger.Info("KAFKA_PRODUCT_EVENTS_TOPIC not set, using default.", zap.String("topic", productEventsTopic))
	}
	kafkaConsumerGroupID := os.Getenv("KAFKA_INVENTORY_CONSUMER_GROUP_ID") // Group ID for Inventory
	if kafkaConsumerGroupID == "" {
		kafkaConsumerGroupID = "inventory-service-group"
		logger.Logger.Info("KAFKA_INVENTORY_CONSUMER_GROUP_ID not set, using default.", zap.String("group_id", kafkaConsumerGroupID))
	}

	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
	}
	defer db.Close()

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Failed to ping PostgreSQL database", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for Inventory Service.")

	// Initialize repository
	inventoryRepo := repository.NewPostgreSQLInventoryRepository(db)

	// Initialize Application Service
	inventoryService := application.NewInventoryService(inventoryRepo)

	// Context for Kafka Consumer and gRPC server for graceful shutdown
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel() // Ensure context is cancelled on main exit

	// Khởi tạo Kafka Product Event Consumer
	// Inventory Service consumer does not directly call Product Service, but the consumer
	// struct expects a product_client.ProductServiceClient for its NewKafkaProductEventConsumer constructor.
	// For now, we pass a nil client as it's not used in its current implementation.
	// If it were to fetch product details, a real client would be needed.
	// var prodClient product_client.ProductServiceClient // No need to initialize if not used
	productEventConsumer := messaging.NewKafkaProductEventConsumer(
		kafkaBroker,
		productEventsTopic,
		kafkaConsumerGroupID,
		inventoryService,
		// prodClient, // No longer needed here as the consumer's constructor doesn't take it
	)
	defer productEventConsumer.Close()

	// Start Kafka consumer in a goroutine
	go productEventConsumer.StartConsuming(rootCtx)
	logger.Logger.Info("Inventory Service: Started consuming events from Kafka topic", zap.String("topic", productEventsTopic))

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register InventoryGRPCServer with the gRPC server
	inventory_client.RegisterInventoryServiceServer(s, inv_grpc_delivery.NewInventoryGRPCServer(inventoryService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Inventory Service (gRPC) listening.", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Failed to serve gRPC server", zap.Error(err))
		}
	}()

	// Start HTTP server for Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsSrv := &http.Server{
			Addr: fmt.Sprintf(":%s", metricsPort),
		}
		logger.Logger.Info("Inventory Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Inventory Service gracefully...")
	rootCancel()     // Cancel context to stop consumer and other goroutines
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Inventory Service stopped.")
}
