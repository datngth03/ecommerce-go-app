package main

import (
	"context"
	// "fmt" // Import fmt cho lỗi
	// "log" // Vẫn giữ để log lỗi khởi tạo logger nếu có. Sẽ được thay thế bằng logger.Logger.Fatal
	"net/http"
	"os"
	"os/signal" // Import cho graceful shutdown
	"syscall"   // Import cho graceful shutdown
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Import godotenv để load biến môi trường
	"go.uber.org/zap"          // Import zap để dùng zap.Error

	api_gateway_http "github.com/datngth03/ecommerce-go-app/internal/api_gateway/delivery/http"
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger" // THÊM: Import logger mới
)

// main là hàm entry point của API Gateway.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load .env file
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log, không dùng log.Println nữa
		logger.Logger.Info("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.", zap.Error(err))
	}

	// Lấy cổng HTTP từ biến môi trường "HTTP_PORT", nếu không có thì dùng cổng mặc định 8080
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
		logger.Logger.Info("HTTP_PORT không được đặt, sử dụng mặc định", zap.String("port", httpPort))
	}

	// Lấy địa chỉ các dịch vụ gRPC từ biến môi trường
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051"
		logger.Logger.Warn("USER_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", userSvcAddr))
	}

	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052"
		logger.Logger.Warn("PRODUCT_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", productSvcAddr))
	}

	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053"
		logger.Logger.Warn("ORDER_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", orderSvcAddr))
	}

	cartSvcAddr := os.Getenv("CART_GRPC_ADDR")
	if cartSvcAddr == "" {
		cartSvcAddr = "localhost:50054"
		logger.Logger.Warn("CART_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", cartSvcAddr))
	}

	paymentSvcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentSvcAddr == "" {
		paymentSvcAddr = "localhost:50055"
		logger.Logger.Warn("PAYMENT_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", paymentSvcAddr))
	}

	shippingSvcAddr := os.Getenv("SHIPPING_GRPC_ADDR")
	if shippingSvcAddr == "" {
		shippingSvcAddr = "localhost:50056"
		logger.Logger.Warn("SHIPPING_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", shippingSvcAddr))
	}

	authSvcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authSvcAddr == "" {
		authSvcAddr = "localhost:50057"
		logger.Logger.Warn("AUTH_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", authSvcAddr))
	}

	notificationSvcAddr := os.Getenv("NOTIFICATION_GRPC_ADDR")
	if notificationSvcAddr == "" {
		notificationSvcAddr = "localhost:50058"
		logger.Logger.Warn("NOTIFICATION_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", notificationSvcAddr))
	}

	inventorySvcAddr := os.Getenv("INVENTORY_GRPC_ADDR")
	if inventorySvcAddr == "" {
		inventorySvcAddr = "localhost:50059"
		logger.Logger.Warn("INVENTORY_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", inventorySvcAddr))
	}

	reviewSvcAddr := os.Getenv("REVIEW_GRPC_ADDR")
	if reviewSvcAddr == "" {
		reviewSvcAddr = "localhost:50060"
		logger.Logger.Warn("REVIEW_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", reviewSvcAddr))
	}

	searchSvcAddr := os.Getenv("SEARCH_GRPC_ADDR")
	if searchSvcAddr == "" {
		searchSvcAddr = "localhost:50061" // Cổng mặc định cho Search Service
		logger.Logger.Warn("SEARCH_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", searchSvcAddr))
	}

	recommendationSvcAddr := os.Getenv("RECOMMENDATION_GRPC_ADDR")
	if recommendationSvcAddr == "" {
		recommendationSvcAddr = "localhost:50062" // Mặc định recommendationSvcAddr
		logger.Logger.Warn("RECOMMENDATION_GRPC_ADDR không được đặt, sử dụng mặc định", zap.String("address", recommendationSvcAddr))
	}

	// Khởi tạo các handler của API Gateway với các gRPC client
	// Truyền tất cả các địa chỉ dịch vụ gRPC cần thiết
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
		// Dùng logger.Logger.Fatal để thoát nếu không thể khởi tạo Gateway Handlers
		logger.Logger.Fatal("Không thể khởi tạo Gateway Handlers", zap.Error(err))
	}
	defer handlers.CloseConnections() // Đảm bảo đóng tất cả các kết nối gRPC khi thoát ứng dụng

	// Khởi tạo Gin router
	router := gin.Default()

	// Đăng ký các routes
	api_gateway_http.RegisterRoutes(router, handlers)

	// Cấu hình HTTP server
	srv := &http.Server{
		Addr:    ":" + httpPort,
		Handler: router,
	}

	// Chạy HTTP server trong một goroutine riêng
	go func() {
		logger.Logger.Info("API Gateway đang lắng nghe", zap.String("port", httpPort))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Dùng logger.Logger.Fatal nếu server thất bại (non-recoverable error)
			logger.Logger.Fatal("Không thể khởi động API Gateway", zap.Error(err))
		}
	}()

	// Chờ tín hiệu dừng ứng dụng (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block cho đến khi nhận được tín hiệu
	logger.Logger.Info("Đang tắt API Gateway...")

	// Thực hiện tắt server một cách graceful
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Đảm bảo context được hủy
	if err := srv.Shutdown(ctx); err != nil {
		// Dùng logger.Logger.Fatal nếu shutdown không graceful
		logger.Logger.Fatal("API Gateway tắt không graceful", zap.Error(err))
	}
	logger.Logger.Info("API Gateway đã tắt.")
}
