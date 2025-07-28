// cmd/api-gateway/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Import godotenv để load biến môi trường

	api_gateway_http "github.com/datngth03/ecommerce-go-app/internal/api_gateway/delivery/http"
)

// main là hàm entry point của API Gateway.
func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Không tìm thấy file .env, tiếp tục mà không load biến môi trường.")
	}

	// Lấy cổng HTTP từ biến môi trường "HTTP_PORT", nếu không có thì dùng cổng mặc định 8080
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Lấy địa chỉ các dịch vụ gRPC từ biến môi trường
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051"
		log.Printf("USER_GRPC_ADDR không được đặt, sử dụng mặc định: %s", userSvcAddr)
	}

	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052"
		log.Printf("PRODUCT_GRPC_ADDR không được đặt, sử dụng mặc định: %s", productSvcAddr)
	}

	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053"
		log.Printf("ORDER_GRPC_ADDR không được đặt, sử dụng mặc định: %s", orderSvcAddr)
	}

	cartSvcAddr := os.Getenv("CART_GRPC_ADDR")
	if cartSvcAddr == "" {
		cartSvcAddr = "localhost:50054"
		log.Printf("CART_GRPC_ADDR không được đặt, sử dụng mặc định: %s", cartSvcAddr)
	}

	paymentSvcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentSvcAddr == "" {
		paymentSvcAddr = "localhost:50055"
		log.Printf("PAYMENT_GRPC_ADDR không được đặt, sử dụng mặc định: %s", paymentSvcAddr)
	}

	shippingSvcAddr := os.Getenv("SHIPPING_GRPC_ADDR")
	if shippingSvcAddr == "" {
		shippingSvcAddr = "localhost:50056"
		log.Printf("SHIPPING_GRPC_ADDR không được đặt, sử dụng mặc định: %s", shippingSvcAddr)
	}

	authSvcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authSvcAddr == "" {
		authSvcAddr = "localhost:50057"
		log.Printf("AUTH_GRPC_ADDR không được đặt, sử dụng mặc định: %s", authSvcAddr)
	}

	notificationSvcAddr := os.Getenv("NOTIFICATION_GRPC_ADDR")
	if notificationSvcAddr == "" {
		notificationSvcAddr = "localhost:50058"
		log.Printf("NOTIFICATION_GRPC_ADDR không được đặt, sử dụng mặc định: %s", notificationSvcAddr)
	}

	inventorySvcAddr := os.Getenv("INVENTORY_GRPC_ADDR")
	if inventorySvcAddr == "" {
		inventorySvcAddr = "localhost:50059"
		log.Printf("INVENTORY_GRPC_ADDR không được đặt, sử dụng mặc định: %s", inventorySvcAddr)
	}

	reviewSvcAddr := os.Getenv("REVIEW_GRPC_ADDR")
	if reviewSvcAddr == "" {
		reviewSvcAddr = "localhost:50060"
		log.Printf("REVIEW_GRPC_ADDR không được đặt, sử dụng mặc định: %s", reviewSvcAddr)
	}

	searchSvcAddr := os.Getenv("SEARCH_GRPC_ADDR")
	if searchSvcAddr == "" {
		searchSvcAddr = "localhost:50061" // Cổng mặc định cho Search Service
		log.Printf("SEARCH_GRPC_ADDR không được đặt, sử dụng mặc định: %s", searchSvcAddr)
	}

	recommendationSvcAddr := os.Getenv("RECOMMENDATION_GRPC_ADDR")
	if recommendationSvcAddr == "" {
		log.Fatalf("RECOMMENDATION_GRPC_ADDR environment variable is not set")
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
		log.Fatalf("Không thể khởi tạo Gateway Handlers: %v", err)
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
		log.Printf("API Gateway đang lắng nghe tại cổng :%s...", httpPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Không thể khởi động API Gateway: %v", err)
		}
	}()

	// Chờ tín hiệu dừng ứng dụng (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Đang tắt API Gateway...")

	// Thực hiện tắt server một cách graceful
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("API Gateway tắt không graceful: %v", err)
	}
	log.Println("API Gateway đã tắt.")
}
