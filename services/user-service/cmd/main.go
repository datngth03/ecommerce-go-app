package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// 2. Setup Database Connections
	// PostgreSQL Connection
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("PostgreSQL connection established")

	// Get underlying sql.DB for repository
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Ensure database connection is closed on exit
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		} else {
			log.Println("Database connection closed")
		}
	}()

	// Redis Connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Redis connection established")

	// Ensure Redis connection is closed on exit
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("Redis connection closed")
		}
	}()

	// 3. Database Migration
	log.Println("Running database migrations...")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// 4. Dependency Injection
	// Initialize Repositories
	userRepo := repository.NewSQLUserRepository(sqlDB)
	tokenRepo := repository.NewRedisTokenRepository(redisClient)

	// Initialize Services
	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenTTL,
		cfg.JWT.RefreshTokenTTL,
		cfg.JWT.ResetTokenTTL,
	)
	userService := service.NewUserService(userRepo, authService)

	log.Println("Services initialized successfully")

	// 5. Setup gRPC Server
	grpcServer := rpc.NewGRPCServer(userService, authService)

	// Create listener
	lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Server.GRPCPort, err)
	}

	// Create gRPC server
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, grpcServer)

	// Register reflection service on gRPC server
	reflection.Register(s)

	log.Printf("gRPC server configured to listen on port %s", cfg.Server.GRPCPort)

	// 6. Start gRPC Server in a goroutine
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("User service is running. Press Ctrl+C to exit...")
	<-quit // Wait for shutdown signal

	log.Println("Received shutdown signal, initiating graceful shutdown...")

	// Stop gRPC server gracefully
	log.Println("Shutting down gRPC server...")
	s.GracefulStop()
	log.Println("gRPC server stopped")

	log.Println("User service shutdown completed")
}
