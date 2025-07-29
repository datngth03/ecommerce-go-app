package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal" // Import cho graceful shutdown
	"syscall"   // Import cho graceful shutdown
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"     // THÊM: Import zap để dùng zap.Error

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to Product Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/order/application"
	order_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/order/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/order/infrastructure/repository" // Import gói repository mới
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                   // THÊM: Import logger mới
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"            // Generated Order gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"        // Product gRPC client
)

// main is the entry point for the Order Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log, không dùng log.Printf/Println nữa
		logger.Logger.Info("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.", zap.Error(err))
	}

	// Get gRPC port from environment variable "GRPC_PORT_ORDER"
	grpcPort := os.Getenv("GRPC_PORT_ORDER")
	if grpcPort == "" {
		grpcPort = "50053" // Default port for Order Service
		logger.Logger.Info("GRPC_PORT_ORDER không được đặt, sử dụng mặc định", zap.String("port", grpcPort))
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Dùng logger.Logger.Fatal để thoát nếu biến môi trường quan trọng không có
		logger.Logger.Fatal("DATABASE_URL environment variable is not set")
	}

	// Get Product Service gRPC address from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		logger.Logger.Info("PRODUCT_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", productSvcAddr))
	}

	// Initialize PostgreSQL database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Không thể kết nối đến PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second) // Dùng context.WithTimeout cho ping ngắn
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Không thể ping cơ sở dữ liệu PostgreSQL", zap.Error(err))
	}
	logger.Logger.Info("Đã kết nối thành công đến cơ sở dữ liệu PostgreSQL cho Order Service.")

	// Initialize Product Service gRPC client
	productConn, err := grpc.Dial(productSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Fatal("Không thể kết nối đến Product Service gRPC", zap.String("address", productSvcAddr), zap.Error(err))
	}
	defer productConn.Close()
	productClient := product_client.NewProductServiceClient(productConn)

	// Initialize PostgreSQL Order Repository
	orderRepo := repository.NewPostgreSQLOrderRepository(db)

	// Initialize Application Service (pass productClient to it)
	orderService := application.NewOrderService(orderRepo, productClient)

	// Initialize gRPC Server
	grpcServer := order_grpc_delivery.NewOrderGRPCServer(orderService)

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Không thể lắng nghe cổng gRPC", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register OrderGRPCServer with the gRPC server
	order_client.RegisterOrderServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Order Service (gRPC) đang lắng nghe", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Không thể phục vụ gRPC server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Dòng này block cho đến khi nhận được tín hiệu

	logger.Logger.Info("Đang tắt Order Service một cách nhẹ nhàng...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Order Service đã tắt.")
}
