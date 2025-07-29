package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal" // Import cho graceful shutdown
	"syscall"   // Import cho graceful shutdown
	"time"

	"github.com/go-redis/redis/v8" // Redis client
	"github.com/joho/godotenv"     // For .env file
	"go.uber.org/zap"              // Import zap để dùng zap.Error

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/cart/application"
	cart_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/cart/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/cart/infrastructure/repository" // Import Redis repository
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                  // THÊM: Import logger mới
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"             // Generated Cart gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"       // Product gRPC client
)

// main is the entry point for the Cart Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log, không dùng log.Printf/Println nữa
		logger.Logger.Info("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.", zap.Error(err))
	}

	// Get gRPC port from environment variable "GRPC_PORT_CART"
	grpcPort := os.Getenv("GRPC_PORT_CART")
	if grpcPort == "" {
		grpcPort = "50054" // Default port for Cart Service
		logger.Logger.Info("GRPC_PORT_CART không được đặt, sử dụng mặc định", zap.String("port", grpcPort))
	}

	// Get Redis address from environment variable "REDIS_ADDR"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address from Docker Compose
		logger.Logger.Info("REDIS_ADDR không được đặt, sử dụng mặc định", zap.String("address", redisAddr))
	}

	// Get Product Service gRPC address from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		logger.Logger.Info("PRODUCT_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", productSvcAddr))
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // Use default DB
	})

	// Ping Redis to check connection
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing() // Release resources associated with this context
	if err := redisClient.Ping(ctxPing).Err(); err != nil {
		// Dùng logger.Logger.Fatal để thoát nếu kết nối Redis thất bại
		logger.Logger.Fatal("Không thể kết nối đến Redis", zap.Error(err))
	}
	logger.Logger.Info("Đã kết nối thành công đến Redis cho Cart Service.")

	// Initialize Product Service gRPC client
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// Dùng logger.Logger.Fatal để thoát nếu kết nối Product Service thất bại
		logger.Logger.Fatal("Không thể kết nối đến Product Service gRPC", zap.String("address", productSvcAddr), zap.Error(err))
	}
	defer productConn.Close() // Đảm bảo đóng kết nối gRPC
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize Redis Cart Repository (e.g., with 24-hour TTL for carts)
	cartRepo := repository.NewRedisCartRepository(redisClient, 24*time.Hour)

	// Initialize Application Service (pass productClient to it)
	cartService := application.NewCartService(cartRepo, productClient)

	// Initialize gRPC Server
	grpcServer := cart_grpc_delivery.NewCartGRPCServer(cartService)

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		// Dùng logger.Logger.Fatal để thoát nếu không thể lắng nghe cổng
		logger.Logger.Fatal("Không thể lắng nghe cổng", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register CartGRPCServer with the gRPC server
	cart_client.RegisterCartServiceServer(s, grpcServer)

	// Register reflection service (useful for gRPC client tools like grpcurl)
	reflection.Register(s)

	logger.Logger.Info("Cart Service (gRPC) đang lắng nghe", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			// Dùng logger.Logger.Fatal nếu server thất bại (non-recoverable error)
			logger.Logger.Fatal("Không thể phục vụ gRPC server", zap.Error(err))
		}
	}()

	// Graceful shutdown: Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	// Bind SIGINT (Ctrl+C) and SIGTERM (kill command) to the quit channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block cho đến khi nhận được tín hiệu

	logger.Logger.Info("Đang tắt Cart Service một cách nhẹ nhàng...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Cart Service đã tắt.")
}
