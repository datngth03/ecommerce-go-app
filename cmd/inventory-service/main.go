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
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/inventory/application"
	inv_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/inventory/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/inventory/infrastructure/messaging" // Import messaging package
	"github.com/datngth03/ecommerce-go-app/internal/inventory/infrastructure/repository"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                // THÊM: Import logger mới
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Generated Inventory gRPC client
)

// main is the entry point for the Inventory Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log
		logger.Logger.Info("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.", zap.Error(err))
	}

	// Get gRPC port from environment variable, default to 50059
	grpcPort := os.Getenv("GRPC_PORT_INVENTORY")
	if grpcPort == "" {
		grpcPort = "50059"
		logger.Logger.Info("GRPC_PORT_INVENTORY không được đặt, sử dụng mặc định", zap.String("port", grpcPort))
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
	kafkaConsumerGroupID := os.Getenv("KAFKA_INVENTORY_CONSUMER_GROUP_ID") // Group ID cho Inventory
	if kafkaConsumerGroupID == "" {
		kafkaConsumerGroupID = "inventory-service-group"
		logger.Logger.Info("KAFKA_INVENTORY_CONSUMER_GROUP_ID không được đặt, sử dụng mặc định", zap.String("group_id", kafkaConsumerGroupID))
	}

	// Connect to PostgreSQL database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Không thể kết nối đến PostgreSQL", zap.Error(err))
	}
	defer db.Close()

	// Ping the database to verify connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Không thể ping cơ sở dữ liệu PostgreSQL", zap.Error(err))
	}
	logger.Logger.Info("Đã kết nối thành công đến cơ sở dữ liệu PostgreSQL cho Inventory Service.")

	// Initialize repository
	inventoryRepo := repository.NewPostgreSQLInventoryRepository(db)

	// Initialize Application Service
	inventoryService := application.NewInventoryService(inventoryRepo)

	// Context cho Kafka Consumer và gRPC server cho graceful shutdown
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel() // Đảm bảo context được hủy khi main exit

	// Khởi tạo Kafka Product Event Consumer
	productEventConsumer := messaging.NewKafkaProductEventConsumer(
		kafkaBroker,
		productEventsTopic,
		kafkaConsumerGroupID,
		inventoryService,
	)
	defer productEventConsumer.Close() // Đảm bảo đóng consumer khi thoát

	// Khởi động Kafka consumer trong một goroutine
	go productEventConsumer.StartConsuming(rootCtx)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Không thể lắng nghe cổng gRPC", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a gRPC server instance
	s := grpc.NewServer()

	// Register InventoryGRPCServer with the gRPC server
	inventory_client.RegisterInventoryServiceServer(s, inv_grpc_delivery.NewInventoryGRPCServer(inventoryService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Inventory Service (gRPC) đang lắng nghe", zap.String("port", grpcPort))

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

	logger.Logger.Info("Đang tắt Inventory Service một cách nhẹ nhàng...")
	rootCancel() // Hủy context để dừng consumer và các goroutine khác
	s.GracefulStop()
	logger.Logger.Info("Inventory Service đã tắt.")
}
