// cmd/cart-service/main.go
package main

import (
	"context"
	"fmt"
	"net"
	"net/http" // THÊM: Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"   // For graceful shutdown
	"time"

	"github.com/go-redis/redis/v8"                                                         // Redis client
	"github.com/joho/godotenv"                                                             // For .env file
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to Product Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/cart/application"
	cart_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/cart/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/cart/infrastructure/repository" // Import Redis repository
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                  // THÊM: Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"                 // THÊM: Add shared tracing
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"             // Generated Cart gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"       // Product gRPC client
)

// main is the entry point for the Cart Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "cart-service"

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

	// Get gRPC port from environment variable "GRPC_PORT_CART"
	grpcPort := os.Getenv("GRPC_PORT_CART")
	if grpcPort == "" {
		grpcPort = "50054" // Default port for Cart Service
		logger.Logger.Info("GRPC_PORT_CART not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9094
	metricsPort := os.Getenv("METRICS_PORT_CART")
	if metricsPort == "" {
		metricsPort = "9104"
		logger.Logger.Info("METRICS_PORT_CART not set, using default.", zap.String("port", metricsPort))
	}

	// Get Redis address from environment variable "REDIS_ADDR"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address from Docker Compose
		logger.Logger.Info("REDIS_ADDR not set, using default.", zap.String("address", redisAddr))
	}

	// Get Product Service gRPC address from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		logger.Logger.Info("PRODUCT_GRPC_ADDR not set, using default.", zap.String("address", productSvcAddr))
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // Use default DB
	})

	// Ping Redis to check connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		logger.Logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to Redis for Cart Service.")

	// Initialize Product Service gRPC client (THÊM INTERCEPTOR CHO TRACING)
	productConn, err := grpc.Dial(
		productSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Product Service gRPC", zap.String("address", productSvcAddr), zap.Error(err))
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize Redis Cart Repository (e.g., with 24-hour TTL for carts)
	cartRepo := repository.NewRedisCartRepository(redisClient, 24*time.Hour)

	// Initialize Application Service (pass productClient to it)
	cartService := application.NewCartService(cartRepo, productClient)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register CartGRPCServer with the gRPC server
	cart_client.RegisterCartServiceServer(s, cart_grpc_delivery.NewCartGRPCServer(cartService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Cart Service (gRPC) listening.", zap.String("port", grpcPort))

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
		logger.Logger.Info("Cart Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Cart Service gracefully...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Cart Service stopped.")
}
