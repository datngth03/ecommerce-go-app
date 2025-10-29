package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/user_service"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/config"

	// "github.com/datngth03/ecommerce-go-app/services/user-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/middleware"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/service"

	sharedCache "github.com/datngth03/ecommerce-go-app/shared/pkg/cache"
	sharedMiddleware "github.com/datngth03/ecommerce-go-app/shared/pkg/middleware"
	sharedTLS "github.com/datngth03/ecommerce-go-app/shared/pkg/tlsutil"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
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

	// 2. Initialize Distributed Tracing
	tracerCleanup, err := sharedTracing.InitTracer(sharedTracing.TracerConfig{
		ServiceName:    cfg.Service.Name,
		ServiceVersion: cfg.Service.Version,
		Environment:    cfg.Service.Environment,
		JaegerEndpoint: os.Getenv("JAEGER_ENDPOINT"),
		Enabled:        os.Getenv("TRACING_ENABLED") == "true",
	})
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerCleanup(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

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

	// 4.5. Initialize Redis Cache for user data caching
	redisPort, _ := strconv.Atoi(cfg.Redis.Port)
	userCache, err := sharedCache.NewRedisCache(sharedCache.CacheConfig{
		Host:     cfg.Redis.Host,
		Port:     redisPort,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Prefix:   "users", // Service-specific prefix
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize user cache: %v (continuing without advanced caching)", err)
		userCache = nil
	} else {
		log.Println("✓ User cache layer initialized")
		defer func() {
			if err := userCache.Close(); err != nil {
				log.Printf("Error closing user cache: %v", err)
			}
		}()
	}

	// 5. Initialize Repositories
	userRepo := repository.NewSQLUserRepository(sqlDB)

	// Wrap with caching layer if available
	var finalUserRepo repository.UserRepositoryInterface = userRepo
	if userCache != nil {
		finalUserRepo = repository.NewCachedUserRepository(userRepo, userCache)
		log.Println("✓ User repository initialized with caching")
	} else {
		log.Println("✓ User repository initialized (without caching)")
	}

	tokenRepo := repository.NewRedisTokenRepository(redisClient)

	// 6. Initialize Services
	authService := service.NewAuthService(
		finalUserRepo,
		tokenRepo,
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTokenTTL,
		cfg.Auth.RefreshTokenTTL,
		cfg.Auth.ResetTokenTTL,
	)
	userService := service.NewUserService(finalUserRepo, authService)
	log.Println("✓ Services initialized")

	// 7. Initialize gRPC Server with Tracing Interceptor and TLS
	var grpcServerOpts []grpc.ServerOption
	grpcServerOpts = append(grpcServerOpts, grpc.UnaryInterceptor(sharedTracing.UnaryServerInterceptor()))

	// Enable TLS if configured
	if cfg.Server.TLS.Enabled {
		tlsCreds, err := sharedTLS.ServerTLSConfig(cfg.Server.TLS.CertFile, cfg.Server.TLS.KeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS credentials: %v", err)
		}
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(tlsCreds))
		log.Printf("✓ TLS enabled for gRPC server (cert: %s)", cfg.Server.TLS.CertFile)
	} else {
		log.Println("⚠️  TLS disabled - using insecure connection")
	}

	grpcServer := grpc.NewServer(grpcServerOpts...)

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

	// 9. Start HTTP Server with Gin
	router := gin.New()

	// Add tracing middleware FIRST (to capture all requests)
	router.Use(sharedTracing.GinMiddleware(cfg.Service.Name))

	// Initialize security middleware
	var securityMiddlewares []gin.HandlerFunc

	// Rate limiting middleware
	if cfg.Security.RateLimit.Enabled {
		rateLimiter := sharedMiddleware.NewIPRateLimiter(
			rate.Limit(cfg.Security.RateLimit.RequestsPerSecond),
			cfg.Security.RateLimit.BurstSize,
		)
		securityMiddlewares = append(securityMiddlewares, sharedMiddleware.RateLimitMiddleware(rateLimiter))
	}

	// Security headers middleware
	securityMiddlewares = append(securityMiddlewares, sharedMiddleware.SecurityHeadersMiddleware())

	// CORS middleware
	if cfg.Security.CORS.Enabled {
		securityMiddlewares = append(securityMiddlewares, sharedMiddleware.CORSMiddleware(cfg.Security.CORS.AllowedOrigins))
	}

	// Timeout middleware
	securityMiddlewares = append(securityMiddlewares, sharedMiddleware.TimeoutMiddleware(cfg.Security.RequestTimeout))

	// Add security middleware first
	for _, mw := range securityMiddlewares {
		router.Use(mw)
	}

	// Enhanced input validation middlewares (5MB max request size)
	for _, mw := range sharedMiddleware.EnhancedValidationMiddlewares(5 * 1024 * 1024) {
		router.Use(mw)
	}

	// Response compression middleware
	router.Use(sharedMiddleware.CompressionMiddleware())

	// Add other middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.PrometheusMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "user-service",
		})
	})

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.HTTPPort),
		Handler: router,
	}

	go func() {
		log.Printf("✓ HTTP server with metrics listening on port %s", cfg.Server.HTTPPort)
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
