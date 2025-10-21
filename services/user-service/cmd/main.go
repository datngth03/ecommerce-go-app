package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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
	log.Printf("User Service v%s starting in %s mode...", cfg.Service.Version, cfg.Service.Environment)

	// 2. Initialize Database Connection
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), &gorm.Config{
		Logger: nil, // Use default logger or configure custom
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("✓ PostgreSQL connection established")

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("✓ Database connection closed")
		}
	}()

	// 3. Check Database Connection (migrations should be run externally via 'make migrate-up')
	log.Println("Verifying database connection...")
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connection verified")

	// 4. Initialize Redis Connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✓ Redis connection established")

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		} else {
			log.Println("✓ Redis connection closed")
		}
	}()

	// 5. Initialize Repositories
	userRepo := repository.NewSQLUserRepository(sqlDB)
	tokenRepo := repository.NewRedisTokenRepository(redisClient)

	// 6. Initialize Services
	authService := service.NewAuthService(
		userRepo,
		tokenRepo,
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
		cfg.Auth.ResetTokenTTL,
	)
	userService := service.NewUserService(userRepo, authService)
	log.Println("✓ Services initialized")

	// 7. Initialize gRPC Server
	grpcServer := grpc.NewServer()

	// Register User Service
	userGRPCServer := rpc.NewGRPCServer(userService, authService)
	pb.RegisterUserServiceServer(grpcServer, userGRPCServer)

	// Register Health Check Service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user_service.UserService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	// 8. Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.Server.GRPCPort, err)
		}

		log.Printf("✓ User gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 9. Start HTTP Health Check Server
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Server.HTTPPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"healthy","service":"user-service"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	}

	go func() {
		log.Printf("✓ HTTP health check server listening on port %s", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 10. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println("✓ User Service is running. Press Ctrl+C to exit...")
	<-quit

	log.Println("Shutting down User Service...")

	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("✓ HTTP server stopped")

	// Stop gRPC server gracefully
	grpcServer.GracefulStop()
	log.Println("✓ gRPC server stopped")

	log.Println("✓ User Service shutdown completed")
}
