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

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/events"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/service"
	sharedMiddleware "github.com/datngth03/ecommerce-go-app/shared/pkg/middleware"
	sharedTLS "github.com/datngth03/ecommerce-go-app/shared/pkg/tlsutil"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Order Service v%s starting in %s mode...", cfg.Service.Version, cfg.Service.Environment)

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

	// 3. Initialize Database Connection
	db, err := repository.ConnectPostgres(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println(" PostgreSQL connection established")

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println(" Database connection closed")
		}
	}()

	// 3. Initialize Redis Connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println(" Redis connection established")

	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing Redis: %v", err)
		} else {
			log.Println(" Redis connection closed")
		}
	}()

	// 4. Initialize Repositories
	orderRepo := repository.NewOrderPostgresRepository(db)
	cartRepo := repository.NewCartPostgresRepository(db, redisClient)
	log.Println("✓ Repositories initialized")

	// 5. Initialize RabbitMQ Publisher
	publisher, err := events.NewPublisher(cfg.GetRabbitMQURL())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	log.Println("✓ RabbitMQ connection established")

	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("Error closing RabbitMQ publisher: %v", err)
		} else {
			log.Println("✓ RabbitMQ publisher closed")
		}
	}()

	// 6. Initialize gRPC Clients with Connection Pooling
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

	// 7. Initialize Services
	orderService := service.NewOrderService(orderRepo, cartRepo, clients.Product, clients.User, publisher)
	cartService := service.NewCartService(cartRepo, clients.Product)
	log.Println("✓ Services initialized")

	// 6. Initialize gRPC Server with Tracing Interceptor and TLS
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

	// Register Order Service
	orderGRPCServer := rpc.NewOrderServer(orderService, cartService)
	pb.RegisterOrderServiceServer(grpcServer, orderGRPCServer)

	// Register Health Check Service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("order_service.OrderService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	// 7. Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.Server.GRPCPort, err)
		}

		log.Printf(" Order gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 8. Setup Gin HTTP Server with Prometheus metrics
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add tracing middleware FIRST (to capture all requests)
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

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "order-service",
		})
	})

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.HTTPPort),
		Handler: router,
	}

	go func() {
		log.Printf(" HTTP health check server listening on port %s", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println(" Order Service is running. Press Ctrl+C to exit...")
	<-quit

	log.Println("Shutting down Order Service...")

	// Stop HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println(" HTTP server stopped")

	// Stop gRPC server gracefully
	grpcServer.GracefulStop()
	log.Println(" gRPC server stopped")

	log.Println(" Order Service shutdown completed")
}
