// cmd/api-gateway/main.go
package main

import (
	"fmt"
	"log"
	"os" // Để đọc biến môi trường

	"github.com/gin-gonic/gin" // Import Gin Gonic framework

	// Import router của API Gateway từ internal
	// Lưu ý: Tên gói mặc định của đường dẫn này là "http"
	router_http "github.com/datngth03/ecommerce-go-app/internal/api_gateway/delivery/http"
)

func main() {
	// Lấy cổng từ biến môi trường "PORT", nếu không có thì dùng cổng mặc định 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Thiết lập Gin ở chế độ Release để tối ưu hiệu suất và không in ra log debug
	// Trong môi trường phát triển, bạn có thể dùng gin.DebugMode
	gin.SetMode(gin.ReleaseMode)

	// Tạo một Gin router mặc định
	router := gin.Default()

	// Đăng ký các routes cho API Gateway
	// Chúng ta gọi hàm RegisterRoutes từ gói cục bộ đã được đổi tên thành router_http
	router_http.RegisterRoutes(router)

	// In ra thông báo khởi động server
	fmt.Printf("API Gateway đang lắng nghe tại cổng :%s...\n", port)

	// Khởi động server HTTP
	// log.Fatal sẽ ghi log lỗi và thoát chương trình nếu server không thể khởi động
	if err := router.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("Không thể khởi động API Gateway: %v", err)
	}
}
