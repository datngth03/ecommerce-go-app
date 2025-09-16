package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	// "time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/handler"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "gorm.io/gorm"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Setup Database Connections
	// PostgreSQL Connection
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("PostgreSQL connection established.")

	// Auto Migrate (Tự động tạo/cập nhật bảng dựa trên struct models.User)
	db.AutoMigrate(&models.User{})

	// Redis Connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connection established.")

	// 3. Database Migration
	// Tự động tạo/cập nhật bảng 'users' dựa trên struct models.User
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed.")

	// 4. Dependency Injection (Wiring everything together)
	// Khởi tạo Repositories
	sqlDB, err := db.DB()
	userRepo := repository.NewSQLUserRepository(sqlDB)
	tokenRepo := repository.NewRedisTokenRepository(redisClient)

	// Khởi tạo Services
	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.ResetTokenTTL,
	)
	userService := service.NewUserService(userRepo, authService)

	// 5. Setup gRPC Server
	grpcServer := handler.NewGRPCServer(userService, authService)

	lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Server.GRPCPort, err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, grpcServer)

	// Đăng ký reflection service trên gRPC server.
	reflection.Register(s)

	// Chạy gRPC server trong một goroutine riêng
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Chờ tín hiệu shutdown

	log.Println("Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("gRPC server stopped.")

	// Đóng các kết nối
	sqlDB.Close()
	redisClient.Close()
	log.Println("Database connections closed.")
}
