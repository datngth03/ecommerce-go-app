// cmd/auth-service/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-redis/redis/v8" // Redis client
	"github.com/joho/godotenv"     // For .env file

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/auth/application"
	auth_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/auth/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/auth/infrastructure/repository" // Import Redis repository
	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth"             // Generated Auth gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"             // User gRPC client
)

// main is the entry point for the Auth Service.
func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get gRPC port from environment variable "GRPC_PORT_AUTH"
	grpcPort := os.Getenv("GRPC_PORT_AUTH")
	if grpcPort == "" {
		grpcPort = "50057" // Default port for Auth Service
	}

	// Get Redis address from environment variable "REDIS_ADDR"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address from Docker Compose
		log.Printf("No REDIS_ADDR environment variable found, using default: %s", redisAddr)
	}

	// Get User Service gRPC address from environment variable "USER_GRPC_ADDR"
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051" // Default port for User Service
		log.Printf("No USER_GRPC_ADDR environment variable found, using default: %s", userSvcAddr)
	}

	// Get JWT Secret from environment variable "JWT_SECRET"
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretjwtkeythatissafetouseinproductionandverylong" // Default secret
		log.Println("WARNING: JWT_SECRET environment variable not set, using default secret. CHANGE THIS IN PRODUCTION!")
	}

	// Get Token TTLs from environment variables
	accessTokenTTLStr := os.Getenv("ACCESS_TOKEN_TTL")
	refreshTokenTTLStr := os.Getenv("REFRESH_TOKEN_TTL")

	accessTokenTTL := 15 * time.Minute    // Default
	refreshTokenTTL := 7 * 24 * time.Hour // Default (7 days)

	if parsedTTL, err := time.ParseDuration(accessTokenTTLStr); err == nil {
		accessTokenTTL = parsedTTL
	}
	if parsedTTL, err := time.ParseDuration(refreshTokenTTLStr); err == nil {
		refreshTokenTTL = parsedTTL
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // Use default DB
	})

	// Ping Redis to check connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis for Auth Service.")

	// Initialize User Service gRPC client
	userConn, err := grpc.Dial(userSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service gRPC: %v", err)
	}
	defer userConn.Close()
	userClient := user_client.NewUserServiceClient(userConn)

	// Initialize Redis Refresh Token Repository
	refreshTokenRepo := repository.NewRedisRefreshTokenRepository(redisClient)

	// Initialize Application Service
	authService := application.NewAuthService(refreshTokenRepo, userClient, jwtSecret, accessTokenTTL, refreshTokenTTL)

	// Initialize gRPC Server
	grpcServer := auth_grpc_delivery.NewAuthGRPCServer(authService) // We will create this next

	// Create a listener on the defined port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	// Create a new gRPC server instance
	s := grpc.NewServer()

	// Register AuthGRPCServer with the gRPC server
	auth_client.RegisterAuthServiceServer(s, grpcServer)

	// Register reflection service.
	reflection.Register(s)

	log.Printf("Auth Service (gRPC) listening on port :%s...", grpcPort)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
