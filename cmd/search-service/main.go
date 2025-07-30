// cmd/search-service/main.go
package main

import (
	"context"
	"fmt"
	"net"
	"net/http" // THÊM: Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"

	// "time" // For context timeouts

	"github.com/joho/godotenv"                                                             // Import godotenv for .env file loading
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Import insecure credentials
	"google.golang.org/grpc/reflection"           // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/search/application"
	search_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/search/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/search/infrastructure/messaging" // Import messaging package
	"github.com/datngth03/ecommerce-go-app/internal/search/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"            // Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"           // THÊM: Add shared tracing
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Product gRPC client for Kafka Consumer
	search_client "github.com/datngth03/ecommerce-go-app/pkg/client/search"   // Generated Search gRPC client
)

// main is the entry point for the Search Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "search-service"

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

	// Get gRPC port from environment variable, default to 50061
	grpcPort := os.Getenv("GRPC_PORT_SEARCH")
	if grpcPort == "" {
		grpcPort = "50061"
		logger.Logger.Info("GRPC_PORT_SEARCH not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9101
	metricsPort := os.Getenv("METRICS_PORT_SEARCH")
	if metricsPort == "" {
		metricsPort = "9111"
		logger.Logger.Info("METRICS_PORT_SEARCH not set, using default.", zap.String("port", metricsPort))
	}

	// Get Elasticsearch address from environment variable
	elasticSearchAddr := os.Getenv("ELASTICSEARCH_ADDR")
	if elasticSearchAddr == "" {
		elasticSearchAddr = "http://localhost:9200" // Default Elasticsearch address
		logger.Logger.Info("ELASTICSEARCH_ADDR not set, using default.", zap.String("address", elasticSearchAddr))
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
	kafkaConsumerGroupID := os.Getenv("KAFKA_SEARCH_CONSUMER_GROUP_ID") // Group ID for Search Service consumer
	if kafkaConsumerGroupID == "" {
		kafkaConsumerGroupID = "search-service-group"
		logger.Logger.Info("KAFKA_SEARCH_CONSUMER_GROUP_ID not set, using default.", zap.String("group_id", kafkaConsumerGroupID))
	}

	// Initialize Elasticsearch Repository
	esRepo, err := repository.NewElasticsearchProductRepository(elasticSearchAddr)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize Elasticsearch repository", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to Elasticsearch for Search Service.")

	// Initialize Application Service
	searchService := application.NewSearchService(esRepo)

	// Context for Kafka Consumer and gRPC server for graceful shutdown
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel() // Ensure context is cancelled on main exit

	// Khởi tạo Kafka Product Event Consumer
	// Kết nối đến Product Service để lấy thông tin chi tiết nếu cần (từ consumer).
	// Hiện tại consumer này sẽ gọi Product Service khi cần thông tin chi tiết sản phẩm để lập chỉ mục.
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR") // Địa chỉ Product Service
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052"
		logger.Logger.Info("PRODUCT_GRPC_ADDR not set for consumer, using default.", zap.String("address", productSvcAddr))
	}
	productConn, err := grpc.Dial(
		productSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))), // THÊM: Client Interceptor cho Tracing
	)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Product Service for consumer", zap.String("address", productSvcAddr), zap.Error(err))
	}
	defer productConn.Close()
	prodClient := product_client.NewProductServiceClient(productConn)

	productEventConsumer := messaging.NewKafkaProductEventConsumer(
		kafkaBroker,
		productEventsTopic,
		kafkaConsumerGroupID,
		searchService,
		prodClient, // Truyền ProductClient vào consumer
	)
	defer productEventConsumer.Close()

	// Start Kafka consumer in a goroutine
	go productEventConsumer.StartConsuming(rootCtx)
	logger.Logger.Info("Search Service: Started consuming events from Kafka topic", zap.String("topic", productEventsTopic))

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register SearchGRPCServer with the gRPC server
	search_client.RegisterSearchServiceServer(s, search_grpc_delivery.NewSearchGRPCServer(searchService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Search Service (gRPC) listening.", zap.String("port", grpcPort))

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
		logger.Logger.Info("Search Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Search Service gracefully...")
	rootCancel()     // Cancel context to stop consumer
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Search Service stopped.")
}
