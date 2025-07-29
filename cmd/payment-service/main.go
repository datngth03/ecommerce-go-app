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
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to Order Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/payment/application"
	payment_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/payment/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/payment/infrastructure/repository" // Import repository
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                     // THÊM: Import logger mới
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"              // Order gRPC client
	payment_client "github.com/datngth03/ecommerce-go-app/pkg/client/payment"          // Generated Payment gRPC client
)

// main is the entry point for the Payment Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log, không dùng log.Printf/Println nữa
		logger.Logger.Info("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.", zap.Error(err))
	}

	// Get gRPC port from environment variable "GRPC_PORT_PAYMENT"
	grpcPort := os.Getenv("GRPC_PORT_PAYMENT")
	if grpcPort == "" {
		grpcPort = "50055" // Default port for Payment Service
		logger.Logger.Info("GRPC_PORT_PAYMENT không được đặt, sử dụng mặc định", zap.String("port", grpcPort))
	}

	// Get DB connection string from environment variable "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Dùng logger.Logger.Fatal để thoát nếu biến môi trường quan trọng không có
		logger.Logger.Fatal("DATABASE_URL environment variable is not set")
	}

	// Get Order Service gRPC address from environment variable "ORDER_GRPC_ADDR"
	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053" // Default port for Order Service
		logger.Logger.Info("ORDER_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", orderSvcAddr))
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
	logger.Logger.Info("Đã kết nối thành công đến cơ sở dữ liệu PostgreSQL cho Payment Service.")

	// Initialize Order Service gRPC client
	orderConn, err := grpc.Dial(orderSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Fatal("Không thể kết nối đến Order Service gRPC", zap.String("address", orderSvcAddr), zap.Error(err))
	}
	defer orderConn.Close()
	orderClient := order_client.NewOrderServiceClient(orderConn)

	// Initialize PostgreSQL Payment Repository
	paymentRepo := repository.NewPostgreSQLPaymentRepository(db)

	// Initialize Application Service (pass orderClient to it)
	paymentService := application.NewPaymentService(paymentRepo, orderClient)

	// Initialize gRPC Server
	grpcServer := payment_grpc_delivery.NewPaymentGRPCServer(paymentService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Không thể lắng nghe cổng gRPC", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register PaymentGRPCServer with the gRPC server
	payment_client.RegisterPaymentServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Payment Service (gRPC) đang lắng nghe", zap.String("port", grpcPort))

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

	logger.Logger.Info("Đang tắt Payment Service một cách nhẹ nhàng...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Payment Service đã tắt.")
}
