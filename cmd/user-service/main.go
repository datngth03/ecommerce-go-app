// cmd/user-service/main.go
package main

import (
	"context"
	"database/sql" // Thư viện chuẩn để tương tác với DB
	"fmt"
	"log"
	"net"  // Để lắng nghe kết nối mạng
	"os"   // Để đọc biến môi trường
	"time" // Để thiết lập timeout cho DB

	"github.com/joho/godotenv" // Để đọc biến môi trường từ file .env
	_ "github.com/lib/pq"      // PostgreSQL driver

	"google.golang.org/grpc"            // Import thư viện gRPC chuẩn
	"google.golang.org/grpc/reflection" // Cho phép gRPC reflection

	"github.com/datngth03/ecommerce-go-app/internal/user/application"
	user_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/user/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/user/infrastructure/repository" // Import gói repository mới
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"             // Mã gRPC đã tạo
)

// main là hàm entry point của User Service.
func main() {
	// Tải biến môi trường từ file .env (nếu có)
	// Điều này hữu ích cho môi trường phát triển cục bộ
	if err := godotenv.Load(); err != nil {
		log.Printf("Không tìm thấy file .env, đang sử dụng biến môi trường hệ thống: %v", err)
	}

	// Lấy cổng gRPC từ biến môi trường "GRPC_PORT"
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// Lấy chuỗi kết nối DB từ biến môi trường "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Chuỗi kết nối mặc định cho PostgreSQL trong Docker Compose
		databaseURL = "postgres://user:password@localhost:5432/user_service_db?sslmode=disable"
		log.Printf("Không tìm thấy biến môi trường DATABASE_URL, sử dụng mặc định: %s", databaseURL)
	}

	// Khởi tạo kết nối cơ sở dữ liệu PostgreSQL
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Không thể kết nối đến cơ sở dữ liệu: %v", err)
	}
	defer db.Close() // Đảm bảo đóng kết nối DB khi ứng dụng kết thúc

	// Thiết lập các thông số kết nối DB (tùy chọn nhưng được khuyến nghị)
	db.SetMaxOpenConns(25)                 // Số lượng kết nối tối đa
	db.SetMaxIdleConns(25)                 // Số lượng kết nối nhàn rỗi tối đa
	db.SetConnMaxLifetime(5 * time.Minute) // Thời gian sống tối đa của một kết nối

	// Ping DB để kiểm tra kết nối
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatalf("Không thể ping cơ sở dữ liệu: %v", err)
	}
	log.Println("Đã kết nối thành công đến cơ sở dữ liệu PostgreSQL.")

	// Khởi tạo PostgreSQLUserRepository
	userRepo := repository.NewPostgreSQLUserRepository(db)

	// Khởi tạo Application Service
	userService := application.NewUserService(userRepo)

	// Khởi tạo gRPC Server
	grpcServer := user_grpc_delivery.NewUserGRPCServer(userService)

	// Tạo một listener trên cổng đã định nghĩa
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Không thể lắng nghe cổng %s: %v", grpcPort, err)
	}

	// Tạo một instance của gRPC server
	s := grpc.NewServer()

	// Đăng ký UserGRPCServer với gRPC server
	user_client.RegisterUserServiceServer(s, grpcServer)

	// Đăng ký reflection service.
	reflection.Register(s)

	log.Printf("User Service (gRPC) đang lắng nghe tại cổng :%s...", grpcPort)

	// Bắt đầu gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Không thể phục vụ gRPC server: %v", err)
	}
}
