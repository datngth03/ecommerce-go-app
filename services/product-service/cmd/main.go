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

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/repository"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/rpc"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/service"

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
	log.Printf("Product Service v%s starting in %s mode...", cfg.Service.Version, cfg.Service.Environment)

	// 2. Initialize Database Connection
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

	// 3. Initialize Repository
	repos, err := repository.NewPostgresRepository(&repository.RepositoryOptions{
		Database: db,
	})
	if err != nil {
		log.Fatalf("Failed to initialize repositories: %v", err)
	}
	log.Println(" Repositories initialized")

	// 4. Initialize Services
	productService := service.NewProductService(repos)
	categoryService := service.NewCategoryService(repos)
	log.Println(" Services initialized")

	// 5. Initialize gRPC Server
	grpcServer := grpc.NewServer()

	// Register Product Service
	productGRPCServer := rpc.NewProductGRPCServer(productService, categoryService)
	pb.RegisterProductServiceServer(grpcServer, productGRPCServer)

	// Register Health Check Service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("product_service.ProductService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for debugging
	reflection.Register(grpcServer)

	// 6. Start gRPC Server
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.GRPCPort))
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %s: %v", cfg.Server.GRPCPort, err)
		}

		log.Printf(" Product gRPC server listening on port %s", cfg.Server.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// 7. Start HTTP Health Check Server
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Server.HTTPPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"healthy","service":"product-service"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}),
	}

	go func() {
		log.Printf(" HTTP health check server listening on port %s", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Println(" Product Service is running. Press Ctrl+C to exit...")
	<-quit

	log.Println("Shutting down Product Service...")

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

	log.Println(" Product Service shutdown completed")
}
