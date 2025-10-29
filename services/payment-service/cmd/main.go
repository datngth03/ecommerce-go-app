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

	"golang.org/x/time/rate"

	"github.com/datngth03/ecommerce-go-app/proto/payment_service"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/service"
	sharedMiddleware "github.com/datngth03/ecommerce-go-app/shared/pkg/middleware"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize distributed tracing
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

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize gRPC Clients with Connection Pooling
	clients, err := client.NewClients(cfg)
	if err != nil {
		log.Fatalf("Failed to create clients: %v", err)
	}
	log.Println("✓ gRPC clients with connection pooling initialized")

	defer func() {
		if err := clients.Close(); err != nil {
			log.Printf("Error closing clients: %v", err)
		} else {
			log.Println("✓ gRPC clients closed")
		}
	}()

	// Initialize repository
	repo := repository.NewPaymentRepository(db)

	// Initialize service
	svc := service.NewPaymentService(repo)

	// Initialize gRPC server with tracing interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(sharedTracing.UnaryServerInterceptor()),
	)
	paymentServer := rpc.NewPaymentServer(svc)
	payment_service.RegisterPaymentServiceServer(grpcServer, paymentServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("payment_service.PaymentService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		log.Printf("Payment gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Start HTTP server with Gin and Prometheus
	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()

		// Add distributed tracing middleware FIRST
		router.Use(sharedTracing.GinMiddleware(cfg.Service.Name))

		// Enhanced input validation middlewares (5MB max request size)
		for _, mw := range sharedMiddleware.EnhancedValidationMiddlewares(5 * 1024 * 1024) {
			router.Use(mw)
		}

		// Response compression middleware
		router.Use(sharedMiddleware.CompressionMiddleware())

		router.Use(gin.Recovery())

		// Security middleware
		var securityMiddleware []gin.HandlerFunc
		if cfg.Security.RateLimit.Enabled {
			rateLimiter := sharedMiddleware.NewIPRateLimiter(
				rate.Limit(cfg.Security.RateLimit.RequestsPerSecond),
				cfg.Security.RateLimit.BurstSize,
			)
			securityMiddleware = append(securityMiddleware, sharedMiddleware.RateLimitMiddleware(rateLimiter))
			log.Printf("✓ Rate limiting enabled: %.1f req/s, burst: %d",
				cfg.Security.RateLimit.RequestsPerSecond, cfg.Security.RateLimit.BurstSize)
		}

		securityMiddleware = append(securityMiddleware, sharedMiddleware.SecurityHeadersMiddleware())

		if cfg.Security.CORS.Enabled {
			securityMiddleware = append(securityMiddleware,
				sharedMiddleware.CORSMiddleware(cfg.Security.CORS.AllowedOrigins))
			log.Printf("✓ CORS enabled for origins: %v", cfg.Security.CORS.AllowedOrigins)
		}

		securityMiddleware = append(securityMiddleware,
			sharedMiddleware.TimeoutMiddleware(cfg.Security.RequestTimeout))

		router.Use(securityMiddleware...)
		router.Use(metrics.PrometheusGinMiddleware())

		// Health check endpoints
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"service": "payment-service",
			})
		})

		router.GET("/ready", func(c *gin.Context) {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not ready",
					"error":  "database not ready",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		})

		// Prometheus metrics endpoint
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))

		log.Printf("✓ Payment HTTP server listening on port %s", cfg.Server.HTTPPort)
		if err := router.Run(fmt.Sprintf(":%s", cfg.Server.HTTPPort)); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Payment Service...")
	grpcServer.GracefulStop()

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	log.Println("Payment Service stopped")
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to PostgreSQL database")
	return db, nil
}
