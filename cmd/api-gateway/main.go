// cmd/api-gateway/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"                                                     // Import godotenv để load biến môi trường
	"github.com/prometheus/client_golang/prometheus/promhttp"                      // For metrics endpoint
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin" // THÊM: OpenTelemetry Gin instrumentation

	// "go.opentelemetry.io/otel"                               // THÊM: Import otel để lấy global TracerProvider
	// "go.opentelemetry.io/otel/sdk/trace"                     // THÊM: Import trace SDK
	"go.uber.org/zap" // For structured logging

	// "google.golang.org/grpc"
	// "google.golang.org/grpc/credentials/insecure" // For gRPC client connections

	api_gateway_http "github.com/datngth03/ecommerce-go-app/internal/api_gateway/delivery/http"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"  // Import shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing" // THÊM: Import shared tracing
)

// main là hàm entry point của API Gateway.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load .env file
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "api-gateway"

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

	// Get HTTP port from environment variable "HTTP_PORT", default to 8080
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
		logger.Logger.Info("HTTP_PORT not set, using default.", zap.String("port", httpPort))
	}

	// Get Metrics port from environment variable, default to 9080
	metricsPort := os.Getenv("METRICS_PORT_API_GATEWAY")
	if metricsPort == "" {
		metricsPort = "9080"
		logger.Logger.Info("METRICS_PORT_API_GATEWAY not set, using default.", zap.String("port", metricsPort))
	}

	// Get gRPC service addresses from environment variables
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051"
		logger.Logger.Info("USER_GRPC_ADDR not set, using default.", zap.String("address", userSvcAddr))
	}

	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052"
		logger.Logger.Info("PRODUCT_GRPC_ADDR not set, using default.", zap.String("address", productSvcAddr))
	}

	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053"
		logger.Logger.Info("ORDER_GRPC_ADDR not set, using default.", zap.String("address", orderSvcAddr))
	}

	cartSvcAddr := os.Getenv("CART_GRPC_ADDR")
	if cartSvcAddr == "" {
		cartSvcAddr = "localhost:50054"
		logger.Logger.Info("CART_GRPC_ADDR not set, using default.", zap.String("address", cartSvcAddr))
	}

	paymentSvcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentSvcAddr == "" {
		paymentSvcAddr = "localhost:50055"
		logger.Logger.Info("PAYMENT_GRPC_ADDR not set, using default.", zap.String("address", paymentSvcAddr))
	}

	shippingSvcAddr := os.Getenv("SHIPPING_GRPC_ADDR")
	if shippingSvcAddr == "" {
		shippingSvcAddr = "localhost:50056"
		logger.Logger.Info("SHIPPING_GRPC_ADDR not set, using default.", zap.String("address", shippingSvcAddr))
	}

	authSvcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authSvcAddr == "" {
		authSvcAddr = "localhost:50057"
		logger.Logger.Info("AUTH_GRPC_ADDR not set, using default.", zap.String("address", authSvcAddr))
	}

	notificationSvcAddr := os.Getenv("NOTIFICATION_GRPC_ADDR")
	if notificationSvcAddr == "" {
		notificationSvcAddr = "localhost:50058"
		logger.Logger.Info("NOTIFICATION_GRPC_ADDR not set, using default.", zap.String("address", notificationSvcAddr))
	}

	inventorySvcAddr := os.Getenv("INVENTORY_GRPC_ADDR")
	if inventorySvcAddr == "" {
		inventorySvcAddr = "localhost:50059"
		logger.Logger.Info("INVENTORY_GRPC_ADDR not set, using default.", zap.String("address", inventorySvcAddr))
	}

	reviewSvcAddr := os.Getenv("REVIEW_GRPC_ADDR")
	if reviewSvcAddr == "" {
		reviewSvcAddr = "localhost:50060"
		logger.Logger.Info("REVIEW_GRPC_ADDR not set, using default.", zap.String("address", reviewSvcAddr))
	}

	searchSvcAddr := os.Getenv("SEARCH_GRPC_ADDR")
	if searchSvcAddr == "" {
		searchSvcAddr = "localhost:50061"
		logger.Logger.Info("SEARCH_GRPC_ADDR not set, using default.", zap.String("address", searchSvcAddr))
	}

	recommendationSvcAddr := os.Getenv("RECOMMENDATION_GRPC_ADDR")
	if recommendationSvcAddr == "" {
		recommendationSvcAddr = "localhost:50062" // Default address for Recommendation Service
		logger.Logger.Info("RECOMMENDATION_GRPC_ADDR not set, using default.", zap.String("address", recommendationSvcAddr))
	}

	// Initialize API Gateway handlers with gRPC clients
	// Note: gRPC clients will automatically pick up the global TracerProvider
	// set by otel.SetTracerProvider(tp)
	handlers, err := api_gateway_http.NewGatewayHandlers(
		userSvcAddr,
		productSvcAddr,
		orderSvcAddr,
		paymentSvcAddr,
		cartSvcAddr,
		shippingSvcAddr,
		authSvcAddr,
		notificationSvcAddr,
		inventorySvcAddr,
		reviewSvcAddr,
		searchSvcAddr,
		recommendationSvcAddr,
	)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize Gateway Handlers", zap.Error(err))
	}
	defer handlers.CloseConnections() // Ensure all gRPC connections are closed on application exit

	// Initialize Gin router
	router := gin.Default()

	// THÊM: Gin OpenTelemetry middleware
	// This middleware will create a span for each incoming HTTP request
	router.Use(otelgin.Middleware(serviceName))

	// Register routes
	api_gateway_http.RegisterRoutes(router, handlers)

	// Configure HTTP server
	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	// Start HTTP server in a goroutine
	go func() {
		logger.Logger.Info("API Gateway listening.", zap.String("port", httpPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Failed to start API Gateway", zap.Error(err))
		}
	}()

	// THÊM: Start HTTP server for Prometheus metrics in a separate goroutine
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsSrv := &http.Server{
			Addr: fmt.Sprintf(":%s", metricsPort),
		}
		logger.Logger.Info("API Gateway Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Listen for interrupt and terminate signals
	<-quit                                               // Block until a signal is received

	logger.Logger.Info("Shutting down API Gateway gracefully...")
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Logger.Fatal("API Gateway shut down ungracefully", zap.Error(err))
	}
	logger.Logger.Info("API Gateway stopped.")
}
