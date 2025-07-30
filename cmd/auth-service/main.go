// cmd/auth-service/main.go
package main

import (
	"context"
	"fmt"
	"net"
	"net/http" // THÊM: Import net/http for metrics server
	"os"
	"os/signal" // For graceful shutdown
	"syscall"   // For graceful shutdown
	"time"

	"github.com/go-redis/redis/v8"                                                         // Redis client
	"github.com/joho/godotenv"                                                             // For .env file
	"github.com/prometheus/client_golang/prometheus/promhttp"                              // THÊM: Import promhttp
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc" // THÊM: OpenTelemetry gRPC instrumentation
	"go.opentelemetry.io/otel"                                                             // THÊM: Import otel để lấy global TracerProvider
	"go.uber.org/zap"                                                                      // THÊM: Add zap for structured logging

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For gRPC client to User Service
	"google.golang.org/grpc/reflection"

	"github.com/datngth03/ecommerce-go-app/internal/auth/application"
	auth_grpc_delivery "github.com/datngth03/ecommerce-go-app/internal/auth/delivery/grpc"
	"github.com/datngth03/ecommerce-go-app/internal/auth/infrastructure/repository" // Import Redis repository
	"github.com/datngth03/ecommerce-go-app/internal/shared/logger"                  // THÊM: Add shared logger
	"github.com/datngth03/ecommerce-go-app/internal/shared/tracing"                 // THÊM: Add shared tracing
	auth_client "github.com/datngth03/ecommerce-go-app/pkg/client/auth"             // Generated Auth gRPC client
	user_client "github.com/datngth03/ecommerce-go-app/pkg/client/user"             // User gRPC client
)

// main is the entry point for the Auth Service.
func main() {
	// Initialize logger at the very beginning
	logger.InitLogger()
	defer logger.SyncLogger() // Ensure all buffered logs are written before exiting

	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		logger.Logger.Info("No .env file found, falling back to system environment variables.", zap.Error(err))
	}

	// Define service name for tracing
	serviceName := "auth-service"

	// Init TracerProvider for OpenTelemetry
	jaegerCollectorURL := os.Getenv("JAEGER_COLLECTOR_URL")
	if jaegerCollectorURL == "" {
		jaegerCollectorURL = "http://localhost:14268/api/traces" // Default Jaeger collector URL
		logger.Logger.Info("JAEGER_COLLECTOR_URL not set, using default.", zap.String("address", jaegerCollectorURL))
	}

	tp, err := tracing.InitTracerProvider(context.Background(), serviceName, jaegerCollectorURL)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize TracerProvider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Logger.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	// Get gRPC port from environment variable "GRPC_PORT_AUTH"
	grpcPort := os.Getenv("GRPC_PORT_AUTH")
	if grpcPort == "" {
		grpcPort = "50057" // Default port for Auth Service
		logger.Logger.Info("GRPC_PORT_AUTH not set, using default.", zap.String("port", grpcPort))
	}

	// Get Metrics port from environment variable, default to 9097
	metricsPort := os.Getenv("METRICS_PORT_AUTH")
	if metricsPort == "" {
		metricsPort = "9107"
		logger.Logger.Info("METRICS_PORT_AUTH not set, using default.", zap.String("port", metricsPort))
	}

	// Get Redis address from environment variable "REDIS_ADDR"
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address from Docker Compose
		logger.Logger.Info("REDIS_ADDR not set, using default.", zap.String("address", redisAddr))
	}

	// Get User Service gRPC address from environment variable "USER_GRPC_ADDR"
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051" // Default port for User Service
		logger.Logger.Info("USER_GRPC_ADDR not set, using default.", zap.String("address", userSvcAddr))
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
	} else {
		logger.Logger.Warn("ACCESS_TOKEN_TTL environment variable parse error, using default.", zap.Error(err))
	}
	if parsedTTL, err := time.ParseDuration(refreshTokenTTLStr); err == nil {
		refreshTokenTTL = parsedTTL
	} else {
		logger.Logger.Warn("REFRESH_TOKEN_TTL environment variable parse error, using default.", zap.Error(err))
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0, // Use default DB
	})

	// Ping Redis to check connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		logger.Logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	logger.Logger.Info("Successfully connected to Redis for Auth Service.")

	// Initialize User Service gRPC client (THÊM INTERCEPTOR CHO TRACING)
	userConn, err := grpc.Dial(
		userSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)
	if err != nil {
		logger.Logger.Fatal("Failed to connect to User Service gRPC", zap.String("address", userSvcAddr), zap.Error(err))
	}
	defer userConn.Close()
	userClient := user_client.NewUserServiceClient(userConn)

	// Initialize Redis Refresh Token Repository
	refreshTokenRepo := repository.NewRedisRefreshTokenRepository(redisClient)

	// Initialize Application Service
	authService := application.NewAuthService(refreshTokenRepo, userClient, jwtSecret, accessTokenTTL, refreshTokenTTL)

	// Create a listener on the defined gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		logger.Logger.Fatal("Failed to listen on gRPC port", zap.String("port", grpcPort), zap.Error(err))
	}

	// Create a new gRPC server instance (THÊM INTERCEPTOR CHO TRACING)
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithTracerProvider(otel.GetTracerProvider()))),
	)

	// Register AuthGRPCServer with the gRPC server
	auth_client.RegisterAuthServiceServer(s, auth_grpc_delivery.NewAuthGRPCServer(authService))

	// Register reflection service.
	reflection.Register(s)

	logger.Logger.Info("Auth Service (gRPC) listening.", zap.String("port", grpcPort))

	// Start the gRPC server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Logger.Fatal("Failed to serve gRPC server", zap.Error(err))
		}
	}()

	// Start HTTP server for Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		metricsSrv := &http.Server{
			Addr: fmt.Sprintf(":%s", metricsPort),
		}
		logger.Logger.Info("Auth Service Metrics server listening.", zap.String("port", metricsPort))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Failed to serve metrics HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // This line blocks until a signal is received

	logger.Logger.Info("Shutting down Auth Service gracefully...")
	s.GracefulStop()
	logger.Logger.Info("Auth Service stopped.")
}
