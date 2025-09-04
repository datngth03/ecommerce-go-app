// cmd/product-service/main.go
package main

import (
	"context"
	"database/sql" // Thư viện chuẩn để tương tác với DB
	"fmt"
	"net"
	"net/http" // THÊM: Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"   // For graceful shutdown
	"time"      // To use time.Duration

	"github.com/joho/godotenv"                                                             // Import godotenv for .env file loading
	_ "github.com/lib/pq"                                                                  // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/product/application"
	prod_delivery "github.com/datngth03/ecommerce-go-app/internal/product/delivery/grpc" // Đổi alias để tránh nhầm lẫn
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/messaging"
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"            // THÊM: Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"           // THÊM: Add shared tracing
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated Product gRPC client
)

// main is the entry point for the Product Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "product-service"

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

	// Get gRPC port from environment variable, default to 50052
	grpcPort := os.Getenv("GRPC_PORT_PRODUCT")
	if grpcPort == "" {
		grpcPort = "50052"
		logger.Logger.Info("GRPC_PORT_PRODUCT not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9092
	metricsPort := os.Getenv("METRICS_PORT_PRODUCT")
	if metricsPort == "" {
		metricsPort = "9102"
		logger.Logger.Info("METRICS_PORT_PRODUCT not set, using default.", zap.String("port", metricsPort))
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

	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
	}
	defer db.Close()

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Failed to ping PostgreSQL database", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for Product Service.")

	// Initialize repositories
	productRepo := repository.NewPostgresProductRepository(db)
	categoryRepo := repository.NewPostgresCategoryRepository(db)
	brandRepo := repository.NewPostgresBrandRepository(db)
	tagRepo := repository.NewPostgresTagRepository(db)

	// Initialize Kafka Event Publisher
	eventPublisher := messaging.NewKafkaProductEventPublisher(kafkaBroker, productEventsTopic)
	defer eventPublisher.Close() // Close writer when application shuts down
	logger.Logger.Info("Kafka Product Event Publisher initialized", zap.String("topic", productEventsTopic), zap.String("broker", kafkaBroker))

	// Initialize Application Service (Inject Event Publisher)
	productService := application.NewProductService(productRepo, categoryRepo, brandRepo, tagRepo, eventPublisher)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register ProductGRPCServer with the gRPC server
	product_client.RegisterProductServiceServer(s, prod_delivery.NewProductGRPCServer(productService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Product Service (gRPC) listening.", zap.String("port", grpcPort))

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
		logger.Logger.Info("Product Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	// Create channel to receive OS signals
	quit := make(chan os.Signal, 1)
	// Notify channel on interrupt and terminate signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Block until a signal is received
	<-quit

	logger.Logger.Info("Shutting down Product Service gracefully...")
	s.GracefulStop() // Stop gRPC server gracefully
	logger.Logger.Info("Product Service stopped.")
}
