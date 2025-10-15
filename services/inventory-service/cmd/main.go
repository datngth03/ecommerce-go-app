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

	"github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/events"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/service"
	"github.com/go-redis/redis/v8"
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

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis
	redisClient := initRedis(cfg)

	// Initialize repository
	repo := repository.NewInventoryRepository(db, redisClient)

	// Initialize service
	svc := service.NewInventoryService(repo)

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
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
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			// Check database connection
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Database not ready"))
				return
			}

			// Check Redis connection
			if err := redisClient.Ping(context.Background()).Err(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Redis not ready"))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Printf("Inventory HTTP server listening on port %s", cfg.Server.HTTPPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.HTTPPort), nil); err != nil {
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
