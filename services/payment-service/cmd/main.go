package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/datngth03/ecommerce-go-app/proto/payment_service"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/payment-service/internal/service"
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

	// Initialize repository
	repo := repository.NewPaymentRepository(db)

	// Initialize service
	svc := service.NewPaymentService(repo)

	// Initialize gRPC server
	grpcServer := grpc.NewServer()
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

	// Start HTTP health check server
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Database not ready"))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Printf("Payment HTTP server listening on port %s", cfg.Server.HTTPPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Server.HTTPPort), nil); err != nil {
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
