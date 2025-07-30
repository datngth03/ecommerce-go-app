// cmd/order-service/main.go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http" // Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"   // For graceful shutdown
	"time"      // To use time.Duration

	"github.com/joho/godotenv"                                                             // Import godotenv for .env file loading
	_ "github.com/lib/pq"                                                                  // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to Product Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/order/application"
	order_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/order/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/order/infrastructure/repository" // Import gói repository mới
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                   // Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"                  // Import shared tracing
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"            // Generated Order gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"        // Product gRPC client
)

// main là hàm entry point của Order Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "order-service"

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

	// Get gRPC port from environment variable "GRPC_PORT_ORDER"
	grpcPort := os.Getenv("GRPC_PORT_ORDER")
	if grpcPort == "" {
		grpcPort = "50053" // Default port for Order Service
		logger.Logger.Info("GRPC_PORT_ORDER not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9093
	metricsPort := os.Getenv("METRICS_PORT_ORDER")
	if metricsPort == "" {
		metricsPort = "9093"
		logger.Logger.Info("METRICS_PORT_ORDER not set, using default.", zap.String("port", metricsPort))
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable is not set")
	}

	// Get Product Service gRPC address from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		logger.Logger.Info("PRODUCT_GRPC_ADDR not set, using default.", zap.String("address", productSvcAddr))
	}

	// Initialize PostgreSQL database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
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
		logger.Logger.Fatal("Failed to ping PostgreSQL database", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for Order Service.")

	// Initialize Product Service gRPC client (SỬ DỤNG CÁCH MỚI CHO INTERCEPTOR)
	productConn, err := grpc.Dial(
		productSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))), // SỬA: Dùng WithStatsHandler và NewClientHandler
	)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to Product Service gRPC", zap.String("address", productSvcAddr), zap.Error(err))
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize PostgreSQL Order Repository
	orderRepo := repository.NewPostgreSQLOrderRepository(db)

	// Initialize Application Service (pass productClient to it)
	orderService := application.NewOrderService(orderRepo, productClient)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (SỬ DỤNG CÁCH MỚI CHO INTERCEPTOR)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))), // SỬA: Dùng StatsHandler và NewServerHandler
	)

	// Register OrderGRPCServer with the gRPC server
	order_client.RegisterOrderServiceServer(s, order_grpc_delivery.NewOrderGRPCServer(orderService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Order Service (gRPC) listening.", zap.String("port", grpcPort))

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
		logger.Logger.Info("Order Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Order Service gracefully...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Order Service stopped.")
}
