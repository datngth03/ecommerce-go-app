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

	"github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/events"

	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/middleware"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/service"

	sharedCache "github.com/datngth03/ecommerce-go-app/shared/pkg/cache"
	sharedMiddleware "github.com/datngth03/ecommerce-go-app/shared/pkg/middleware"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
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

	// Initialize Redis
	redisClient := initRedis(cfg)

	// Initialize Redis cache for inventory data
	redisPort, _ := strconv.Atoi(cfg.Redis.Port)
	inventoryCache, err := sharedCache.NewRedisCache(sharedCache.CacheConfig{
		Host:     cfg.Redis.Host,
		Port:     redisPort,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Prefix:   "inventory", // Service-specific prefix
	})
	if err != nil {
		log.Printf("Warning: Failed to initialize inventory cache: %v (continuing without advanced caching)", err)
		inventoryCache = nil
	} else {
		log.Println("✓ Inventory cache layer initialized")
		defer func() {
			if err := inventoryCache.Close(); err != nil {
				log.Printf("Error closing inventory cache: %v", err)
			}
		}()
	}

	// Initialize repository
	repo := repository.NewInventoryRepository(db, redisClient)

	// Wrap with caching layer if available
	var finalRepo repository.InventoryRepository = repo
	if inventoryCache != nil {
		finalRepo = repository.NewCachedInventoryRepository(repo, inventoryCache)
		log.Println("✓ Inventory repository initialized with caching")
	} else {
		log.Println("✓ Inventory repository initialized (without caching)")
	}

	// Initialize service
	svc := service.NewInventoryService(finalRepo)

	// Initialize gRPC server with tracing interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(sharedTracing.UnaryServerInterceptor()),
	)
	inventoryServer := rpc.NewInventoryServer(svc)
	inventory_service.RegisterInventoryServiceServer(grpcServer, inventoryServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("inventory_service.InventoryService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for grpcurl
	reflection.Register(grpcServer)

	// Initialize event subscriber
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subscriber, err := events.NewEventSubscriber(svc, cfg.GetRabbitMQURL())
	if err != nil {
		log.Printf("Warning: Failed to initialize event subscriber: %v", err)
	} else {
		err = subscriber.Start(ctx)
		if err != nil {
			log.Printf("Warning: Failed to start event subscriber: %v", err)
		}
		defer subscriber.Close()
	}

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		log.Printf("Inventory gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Start HTTP health check server
	go func() {
		// Set Gin mode
		if cfg.Service.Environment == "production" {
			gin.SetMode(gin.ReleaseMode)
		}

		router := gin.New()

		// Add distributed tracing middleware FIRST
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
		router.Use(middleware.PrometheusGinMiddleware())

		// Health check endpoint
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "healthy",
				"service": "inventory-service",
			})
		})

		// Readiness check endpoint
		router.GET("/ready", func(c *gin.Context) {
			// Check database connection
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not ready",
					"error":  "Database not ready",
				})
				return
			}

			// Check Redis connection
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not ready",
					"error":  "Redis not ready",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": "ready",
			})
		})

		// Prometheus metrics endpoint
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))

		httpAddr := fmt.Sprintf(":%s", cfg.Server.HTTPPort)
		log.Printf("Inventory HTTP server listening on port %s", cfg.Server.HTTPPort)
		log.Printf("Health check: http://localhost:%s/health", cfg.Server.HTTPPort)
		log.Printf("Ready check: http://localhost:%s/ready", cfg.Server.HTTPPort)
		log.Printf("Metrics endpoint: http://localhost:%s/metrics", cfg.Server.HTTPPort)

		httpServer := &http.Server{
			Addr:         httpAddr,
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Inventory Service...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	// Close database
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	// Close Redis
	redisClient.Close()

	log.Println("Inventory Service stopped")
}

// initDB initializes database connection
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

	// Set connection pool settings
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

// initRedis initializes Redis connection
func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Connected to Redis")
	}

	return client
}
