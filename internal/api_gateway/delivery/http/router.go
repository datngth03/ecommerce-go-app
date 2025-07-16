// internal/api_gateway/delivery/http/router.go
package http

import (
	"net/http" // Thư viện HTTP chuẩn của Go

	"github.com/gin-gonic/gin" // Import Gin Gonic framework
)

// RegisterRoutes đăng ký tất cả các routes cho API Gateway.
// Nó nhận một con trỏ đến gin.Engine (router chính) làm đối số.
func RegisterRoutes(router *gin.Engine) {
	// Định nghĩa một route GET cho đường dẫn "/health"
	// Đây là một endpoint phổ biến để kiểm tra trạng thái hoạt động của service
	router.GET("/health", healthCheckHandler)

	// Bạn có thể thêm các routes khác ở đây khi phát triển các chức năng mới
	// Ví dụ: router.POST("/users", userHandler.CreateUser)
	//        router.GET("/products", productHandler.GetProducts)
}

// healthCheckHandler là hàm xử lý cho endpoint "/health".
// Nó trả về một phản hồi JSON đơn giản để cho biết service đang hoạt động.
func healthCheckHandler(c *gin.Context) {
	// gin.Context cung cấp các phương thức để xử lý yêu cầu và phản hồi
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"message": "API Gateway đang hoạt động!",
	})
}
