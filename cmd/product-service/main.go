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

	"github.com/joho/godotenv" // Import godotenv for .env file loading
	_ "github.com/lib/pq"      // PostgreSQL driver
	"go.uber.org/zap"          // THÊM: Import zap để dùng zap.Error

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection" // For gRPC reflection (useful for client tools)

	"github.com/datngth03/ecommerce-go-app/internal/product/application"
	prod_delivery "github.com/datngth03/ecommerce-go-app/internal/product/delivery/grpc" // Đổi alias để tránh nhầm lẫn
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/messaging"
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"            // THÊM: Import logger mới
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Import mã gRPC đã tạo (chứa RegisterProductServiceServer)
)

// main is the entry point for the Product Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log, không dùng log.Printf/Println nữa
		logger.Logger.Info("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.", zap.Error(err))
	}

	// Get gRPC port from environment variable, default to 50052
	grpcPort := os.Getenv("GRPC_PORT_PRODUCT")
	if grpcPort == "" {
		grpcPort = "50052"
		logger.Logger.Info("GRPC_PORT_PRODUCT không được đặt, sử dụng mặc định", zap.String("port", grpcPort))
	}

	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Dùng logger.Logger.Fatal để thoát nếu biến môi trường quan trọng không có
		logger.Logger.Fatal("DATABASE_URL environment variable is not set")
	}

	// Get Kafka broker address from environment variable
	kafkaBroker := os.Getenv("KAFKA_BROKER_ADDR")
	if kafkaBroker == "" {
		kafkaBroker = "localhost:9092" // Default Kafka broker address
		logger.Logger.Info("KAFKA_BROKER_ADDR không được đặt, sử dụng mặc định", zap.String("address", kafkaBroker))
	}
	productEventsTopic := os.Getenv("KAFKA_PRODUCT_EVENTS_TOPIC")
	if productEventsTopic == "" {
		productEventsTopic = "product_events" // Default Kafka topic for product events
		logger.Logger.Info("KAFKA_PRODUCT_EVENTS_TOPIC không được đặt, sử dụng mặc định", zap.String("topic", productEventsTopic))
	}

	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Không thể kết nối đến PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Ping the database to verify connection
	// Use a short-lived context for the ping
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Không thể ping cơ sở dữ liệu PostgreSQL", zap.Error(err))
	}
	logger.Logger.Info("Đã kết nối thành công đến cơ sở dữ liệu PostgreSQL cho Product Service.")

	// Initialize repositories
	productRepo := repository.NewPostgreSQLProductRepository(db)
	categoryRepo := repository.NewPostgreSQLCategoryRepository(db)

	// Initialize Kafka Event Publisher
	eventPublisher := messaging.NewKafkaProductEventPublisher(kafkaBroker, productEventsTopic)
	defer eventPublisher.Close() // Close writer when application shuts down
	logger.Logger.Info("Kafka Product Event Publisher initialized", zap.String("topic", productEventsTopic), zap.String("broker", kafkaBroker))

	// Initialize Application Service (Inject Event Publisher)
	productService := application.NewProductService(productRepo, categoryRepo, eventPublisher)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Không thể lắng nghe cổng gRPC", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a gRPC server instance
	s := grpc.NewServer()

	// Đăng ký ProductGRPCServer với gRPC server (SỬA LẠI ĐÂY)
	// Gọi RegisterProductServiceServer từ gói 'product_client' (mã gRPC đã tạo)
	// Truyền vào instance của triển khai gRPC server của bạn (từ gói 'prod_delivery')
	product_client.RegisterProductServiceServer(s, prod_delivery.NewProductGRPCServer(productService))

	// Register reflection service. This allows gRPC client tools
	// to discover available services and methods without the .proto file.
	reflection.Register(s)

	logger.Logger.Info("Product Service (gRPC) đang lắng nghe", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Không thể phục vụ gRPC server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	// Create channel to receive OS signals
	quit := make(chan os.Signal, 1)
	// Notify channel on interrupt and terminate signals
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Block until a signal is received
	<-quit

	logger.Logger.Info("Đang tắt Product Service một cách nhẹ nhàng...")
	s.GracefulStop() // Stop gRPC server gracefully
	logger.Logger.Info("Product Service đã tắt.")
}
