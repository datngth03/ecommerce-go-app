// cmd/user-service/main.go
package main

import (
	"context"
	"database/sql" // Thư viện chuẩn để tương tác với DB
	"fmt"
	"net"
	"net/http" // Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"   // For graceful shutdown
	"time"      // Để thiết lập timeout cho DB

	"github.com/joho/godotenv"                                                             // Để đọc biến môi trường từ file .env
	_ "github.com/lib/pq"                                                                  // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // Thêm zap để ghi log có cấu trúc

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // Cho phép gRPC reflection

	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"  // Thêm logger dùng chung
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing" // Thêm shared tracing
	"github.com/datngth03/ecommerce-go-app/internal/user/application"
	user_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/user/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/user/infrastructure/repository" // Import gói repository mới
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"             // Mã gRPC đã tạo
)

// main là hàm entry point của User Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Tải biến môi trường từ file .env (nếu có)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "user-service"

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

	// Lấy cổng gRPC từ biến môi trường "GRPC_PORT_USER"
	grpcPort := os.Getenv("GRPC_PORT_USER")
	if grpcPort == "" {
		grpcPort = "50051"
		logger.Logger.Info("GRPC_PORT_USER not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9091
	metricsPort := os.Getenv("METRICS_PORT_USER")
	if metricsPort == "" {
		metricsPort = "9101"
		logger.Logger.Info("METRICS_PORT_USER not set, using default.", zap.String("port", metricsPort))
	}

	// Lấy chuỗi kết nối DB từ biến môi trường "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://user:password@localhost:5432/ecommerce_core_db?sslmode=disable"
		logger.Logger.Info("DATABASE_URL not set, using default.", zap.String("url", databaseURL))
	}

	// Khởi tạo kết nối cơ sở dữ liệu PostgreSQL
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to PostgreSQL database", zap.Error(err))
	}
	defer db.Close() // Đảm bảo đóng kết nối DB khi ứng dụng kết thúc

	// Thiết lập các thông số kết nối DB (tùy chọn nhưng được khuyến nghị)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping DB để kiểm tra kết nối
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Failed to ping PostgreSQL database", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to PostgreSQL database for User Service.")

	// Khởi tạo PostgreSQLUserRepository
	userRepo := repository.NewPostgreSQLUserRepository(db)

	// Khởi tạo Application Service
	userService := application.NewUserService(userRepo)

	// Tạo gRPC server với StatsHandler cho OpenTelemetry
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	user_client.RegisterUserServiceServer(s, user_grpc_delivery.NewUserGRPCServer(userService))
	reflection.Register(s)

	logger.Logger.Info("User Service (gRPC) listening.", zap.String("port", grpcPort))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

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
		logger.Logger.Info("User Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Logger.Info("Shutting down User Service gracefully...")
	s.GracefulStop()
	logger.Logger.Info("User Service stopped.")
}
