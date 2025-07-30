// cmd/notification-service/main.go
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
	"time"

	"github.com/joho/godotenv"                                                             // For .env file
	_ "github.com/lib/pq"                                                                  // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/notification/application"
	notification_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/notification/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/notification/infrastructure/repository" // Import repository
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                          // THÊM: Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"                         // THÊM: Add shared tracing
	notification_client "github.com/datngth03/ecommerce-go-app/pkg/client/notification"     // Generated Notification gRPC client
)

// main is the entry point for the Notification Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "notification-service"

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

	// Get gRPC port from environment variable "GRPC_PORT_NOTIFICATION"
	grpcPort := os.Getenv("GRPC_PORT_NOTIFICATION")
	if grpcPort == "" {
		grpcPort = "50058" // Default port for Notification Service
		logger.Logger.Info("GRPC_PORT_NOTIFICATION not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9098
	metricsPort := os.Getenv("METRICS_PORT_NOTIFICATION")
	if metricsPort == "" {
		metricsPort = "9108"
		logger.Logger.Info("METRICS_PORT_NOTIFICATION not set, using default.", zap.String("port", metricsPort))
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable is not set.")
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
	logger.Logger.Info("Successfully connected to PostgreSQL database for Notification Service.")

	// Initialize PostgreSQL Notification Repository
	notificationRepo := repository.NewPostgreSQLNotificationRepository(db)

	// Initialize Application Service
	notificationService := application.NewNotificationService(notificationRepo)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register NotificationGRPCServer with the gRPC server
	notification_client.RegisterNotificationServiceServer(s, notification_grpc_delivery.NewNotificationGRPCServer(notificationService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Notification Service (gRPC) listening.", zap.String("port", grpcPort))

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
		logger.Logger.Info("Notification Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Notification Service gracefully...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Notification Service stopped.")
}
