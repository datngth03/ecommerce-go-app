package main

import (
	"context"
	"fmt" // Kept for logging initial logger errors, if any. Will be replaced by logger.Logger.Fatal
	"net"
	"os"
	"os/signal" // Import for graceful shutdown
	"syscall"   // Import for graceful shutdown
	"time"

	"github.com/go-redis/redis/v8" // Redis client
	"github.com/joho/godotenv"     // For .env file
	"go.uber.org/zap"              // Import zap to use zap.Error

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/auth/application"
	auth_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/auth/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/auth/infrastructure/repository" // Import Redis repository
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                  // THÊM: Import logger mới
	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth"             // Generated Auth gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"             // User gRPC client
)

// main is the entry point for the Auth Service.
func main() {
	// Khởi tạo logger ngay từ đầu
	logger.InitLogger()
	defer logger.SyncLogger() // Đảm bảo tất cả log được ghi trước khi ứng dụng thoát

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		// Dùng logger mới để log, không dùng log.Printf nữa
		logger.Logger.Info("No .env file found or failed to load", zap.Error(err))
	}

	// Get gRPC port from environment variable "GRPC_PORT_AUTH"
	grpcPort := os.Getenv("GRPC_PORT_AUTH")
	if grpcPort == "" {
		grpcPort = "50057" // Default port for Auth Service
		logger.Logger.Info("GRPC_PORT_AUTH environment variable not set, using default", zap.String("port", grpcPort))
	}

	// Get Redis address from environment variable "REDIS_ADDR"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address from Docker Compose
		logger.Logger.Info("REDIS_ADDR environment variable not found, using default", zap.String("address", redisAddr))
	}

	// Get User Service gRPC address from environment variable "USER_GRPC_ADDR"
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051" // Default port for User Service
		logger.Logger.Info("USER_GRPC_ADDR environment variable not found, using default", zap.String("address", userSvcAddr))
	}

	// Get JWT Secret from environment variable "JWT_SECRET"
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretjwtkeythatissafetouseinproductionandverylong" // Default secret
		logger.Logger.Warn("JWT_SECRET environment variable not set, using default secret. CHANGE THIS IN PRODUCTION!")
	}

	// Get Token TTLs from environment variables
	accessTokenTTLStr := os.Getenv("ACCESS_TOKEN_TTL")
	refreshTokenTTLStr := os.Getenv("REFRESH_TOKEN_TTL")

	accessTokenTTL := 15 * time.Minute    // Default
	refreshTokenTTL := 7 * 24 * time.Hour // Default (7 days)

	if parsedTTL, err := time.ParseDuration(accessTokenTTLStr); err == nil {
		accessTokenTTL = parsedTTL
	} else if accessTokenTTLStr != "" {
		logger.Logger.Warn("Invalid ACCESS_TOKEN_TTL format, using default", zap.String("value", accessTokenTTLStr), zap.Error(err))
	}

	if parsedTTL, err := time.ParseDuration(refreshTokenTTLStr); err == nil {
		refreshTokenTTL = parsedTTL
	} else if refreshTokenTTLStr != "" {
		logger.Logger.Warn("Invalid REFRESH_TOKEN_TTL format, using default", zap.String("value", refreshTokenTTLStr), zap.Error(err))
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // Use default DB
	})

	// Ping Redis to check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Release resources associated with this context
	if err := redisClient.Ping(ctx).Err(); err != nil {
		// Dùng logger.Logger.Fatal để thoát nếu kết nối Redis thất bại
		logger.Logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to Redis for Auth Service.")

	// Initialize User Service gRPC client
	userConn, err := grpc.Dial(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// Dùng logger.Logger.Fatal để thoát nếu kết nối User Service thất bại
		logger.Logger.Fatal("Failed to connect to User Service gRPC", zap.String("address", userSvcAddr), zap.Error(err))
	}
	defer userConn.Close() // Đảm bảo đóng kết nối gRPC

	userClient := user_client.NewUserServiceClient(userConn)

	// Initialize Redis Refresh Token Repository
	refreshTokenRepo := repository.NewRedisRefreshTokenRepository(redisClient)

	// Initialize Application Service
	authService := application.NewAuthService(refreshTokenRepo, userClient, jwtSecret, accessTokenTTL, refreshTokenTTL)

	// Initialize gRPC Server
	grpcServer := auth_grpc_delivery.NewAuthGRPCServer(authService)

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		// Dùng logger.Logger.Fatal để thoát nếu không thể lắng nghe cổng
		logger.Logger.Fatal("Failed to listen on port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register AuthGRPCServer with the gRPC server
	auth_client.RegisterAuthServiceServer(s, grpcServer)

	// Register reflection service (useful for gRPC client tools like grpcurl)
	reflection.Register(s)

	logger.Logger.Info("Auth Service (gRPC) listening", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			// Dùng logger.Logger.Fatal nếu server thất bại (non-recoverable error)
			logger.Logger.Fatal("Failed to serve gRPC server", zap.Error(err))
		}
	}()

	// Graceful shutdown: Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	// Bind SIGINT (Ctrl+C) and SIGTERM (kill command) to the quit channel
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received

	logger.Logger.Info("Shutting down Auth Service gracefully...")
	s.GracefulStop() // Gracefully stop the gRPC server
	logger.Logger.Info("Auth Service stopped.")
}
