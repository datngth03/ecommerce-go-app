package main

import (
	"context"
	"database/sql" // Thư viện chuẩn để tương tác với DB
	"fmt"
	"net"       // Để lắng nghe kết nối mạng
	"os"        // Để đọc biến môi trường
	"os/signal" // Import cho graceful shutdown
	"syscall"   // Import cho graceful shutdown
	"time"      // Để thiết lập timeout cho DB

	"github.com/joho/godotenv" // Để đọc biến môi trường từ file .env
	_ "github.com/lib/pq"      // PostgreSQL driver
	"go.uber.org/zap"          // Thêm zap để ghi log có cấu trúc

	"google.golang.org/grpc"            // Import thư viện gRPC chuẩn
	"google.golang.org/grpc/reflection" // Cho phép gRPC reflection

	"github.com/datngth03/ecommerce-go-app/internal/shared/logger" // Thêm logger dùng chung
	"github.com/datngth03/ecommerce-go-app/internal/user/application"
	user_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/user/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/user/infrastructure/repository" // Import gói repository mới
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"             // Mã gRPC đã tạo
)

// main là hàm entry point của User Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi vào output trước khi thoát

	// Tải biến môi trường từ file .env (nếu có)
	// Điều này hữu ích cho môi trường phát triển cục bộ
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("Không tìm thấy file .env hoặc không thể tải.", zap.Error(err))
	}

	// Lấy cổng gRPC từ biến môi trường "GRPC_PORT"
	grpcPort := os.Getenv("GRPC_PORT_USER")
	if grpcPort == "" {
		grpcPort = "50051"
		logger.Logger.Info("GRPC_PORT_USER không được đặt, sử dụng mặc định.", zap.String("port", grpcPort))
	}

	// Lấy chuỗi kết nối DB từ biến môi trường "DATABASE_URL"
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		logger.Logger.Fatal("DATABASE_URL biến môi trường không được đặt.")
	}

	// Khởi tạo kết nối cơ sở dữ liệu PostgreSQL
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Logger.Fatal("Không thể kết nối đến cơ sở dữ liệu.", zap.Error(err))
	}
	defer db.Close() // Đảm bảo đóng kết nối DB khi ứng dụng kết thúc

	// Thiết lập các thông số kết nối DB (tùy chọn nhưng được khuyến nghị)
	db.SetMaxOpenConns(25)                 // Số lượng kết nối tối đa
	db.SetMaxIdleConns(25)                 // Số lượng kết nối nhàn rỗi tối đa
	db.SetConnMaxLifetime(5 * time.Minute) // Thời gian sống tối đa của một kết nối

	// Ping DB để kiểm tra kết nối
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err = db.PingContext(pingCtx); err != nil {
		logger.Logger.Fatal("Không thể ping cơ sở dữ liệu.", zap.Error(err))
	}
	logger.Logger.Info("Đã kết nối thành công đến cơ sở dữ liệu PostgreSQL.")

	// Khởi tạo PostgreSQLUserRepository
	userRepo := repository.NewPostgreSQLUserRepository(db)

	// Khởi tạo Application Service
	userService := application.NewUserService(userRepo)

	// Khởi tạo gRPC Server
	grpcServer := user_grpc_delivery.NewUserGRPCServer(userService)

	// Tạo một listener trên cổng đã định nghĩa
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Không thể lắng nghe cổng gRPC.", zap.String("port", grpcPort), zap.Error(err))
	}

	// Tạo một instance của gRPC server
	s := grpc.NewServer()

	// Đăng ký UserGRPCServer với gRPC server
	user_client.RegisterUserServiceServer(s, grpcServer)

	// Đăng ký reflection service.
	reflection.Register(s)

	logger.Logger.Info("User Service (gRPC) đang lắng nghe tại cổng.", zap.String("port", grpcPort))

	// Bắt đầu gRPC server trong một goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Không thể phục vụ gRPC server.", zap.Error(err))
		}
	}()

	// Graceful shutdown
	// Tạo channel để nhận tín hiệu OS
	quit := make(chan os.Signal, 1)
	// Thông báo channel khi nhận tín hiệu ngắt (Ctrl+C) và kết thúc
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// Chặn cho đến khi nhận được tín hiệu
	<-quit

	logger.Logger.Info("Đang tắt User Service một cách graceful...")
	s.GracefulStop() // Dừng gRPC server một cách graceful
	logger.Logger.Info("User Service đã tắt.")
}
