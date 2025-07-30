// cmd/recommendation-service/main.go
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
	"time"      // To use time.Duration

	"github.com/joho/godotenv"                                                             // To read environment variables from .env file
	_ "github.com/lib/pq"                                                                  // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For insecure gRPC connections (dev only)
	"google.golang.org/grpc/reflection"           // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/recommendation/application"
	recommendation_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/recommendation/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/recommendation/infrastructure/messaging" // Add Kafka consumer
	"github.com/datngth03/ecommerce-go-app/internal/recommendation/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"            // Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"           // THÊM: Import shared tracing
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Product Service Client

	// Generated gRPC client for Recommendation Service
	recommendation_client "github.com/datngth03/ecommerce-go-app/pkg/client/recommendation"
)

// main is the entry point for the Recommendation Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "recommendation-service"

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

	// Get gRPC port from environment variable, default to 50062
	grpcPort := os.Getenv("GRPC_PORT_RECOMMENDATION")
	if grpcPort == "" {
		grpcPort = "50062" // Default port for Recommendation Service
		logger.Logger.Info("GRPC_PORT_RECOMMENDATION not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9102
	metricsPort := os.Getenv("METRICS_PORT_RECOMMENDATION")
	if metricsPort == "" {
		metricsPort = "9112"
		logger.Logger.Info("METRICS_PORT_RECOMMENDATION not set, using default.", zap.String("port", metricsPort))
	}

	// Get Product Service gRPC address
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default address for Product Service
		logger.Logger.Info("PRODUCT_GRPC_ADDR not set, using default.", zap.String("address", productSvcAddr))
	}

	// Get Kafka broker address and topic
	kafkaBroker := os.Getenv("KAFKA_BROKER_ADDR")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092"
		logger.Logger.Info("KAFKA_BROKER_ADDR not set, using default.", zap.String("address", kafkaBroker))
	}
	productEventsTopic := os.Getenv("KAFKA_PRODUCT_EVENTS_TOPIC")
	if productEventsTopic == "" {
		productEventsTopic = "product_events"
		logger.Logger.Info("KAFKA_PRODUCT_EVENTS_TOPIC not set, using default.", zap.String("topic", productEventsTopic))
	}
	kafkaConsumerGroupID := os.Getenv("KAFKA_RECOMMENDATION_CONSUMER_GROUP_ID") // Group ID for Recommendation consumer
	if kafkaConsumerGroupID == "" {
		kafkaConsumerGroupID = "recommendation-service-group"
		logger.Logger.Info("KAFKA_RECOMMENDATION_CONSUMER_GROUP_ID not set, using default.", zap.String("group_id", kafkaConsumerGroupID))
	}

	// Connect to PostgreSQL database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable is not set.")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
	}
	defer db.Close() // Ensure DB connection is closed on application exit

	// Ping database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Failed to ping PostgreSQL database", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for Recommendation Service.")

	// Set connection pool settings (optional but recommended for production)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Initialize Repository
	interactionRepo := repository.NewPostgreSQLUserInteractionRepository(db)

	// Initialize Product Service gRPC client (THÊM INTERCEPTOR CHO TRACING)
	productConn, err := grpc.Dial(
		productSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Product Service for Recommendation", zap.String("address", productSvcAddr), zap.Error(err))
	}
	defer productConn.Close()
	prodClient := product_client.NewProductServiceClient(productConn)

	// Initialize Application Service
	recommendationService := application.NewRecommendationService(interactionRepo, prodClient)

	// Context for Kafka Consumer and gRPC server for graceful shutdown
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel() // Ensure context is cancelled on main exit

	// Initialize Kafka Product Event Consumer
	productEventConsumer := messaging.NewKafkaProductEventConsumer(
		kafkaBroker,
		productEventsTopic,
		kafkaConsumerGroupID,
		recommendationService,
	)
	defer productEventConsumer.Close()

	// Start Kafka consumer in a goroutine
	go productEventConsumer.StartConsuming(rootCtx)
	logger.Logger.Info("Recommendation Service: Started consuming events from Kafka topic", zap.String("topic", productEventsTopic))

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register RecommendationGRPCServer with the gRPC server
	recommendation_client.RegisterRecommendationServiceServer(s, recommendation_grpc_delivery.NewRecommendationGRPCServer(recommendationService))

	// Register reflection service. This allows gRPC client tools
	// to discover available services and methods without .proto files.
	reflection.Register(s)

	logger.Logger.Info("Recommendation Service (gRPC) listening.", zap.String("port", grpcPort))

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
		logger.Logger.Info("Recommendation Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Recommendation Service gracefully...")
	rootCancel()     // Cancel context to stop consumer and other goroutines
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Recommendation Service stopped.")
}
