package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/ecommerce/proto/user_service"
	"github.com/ecommerce/services/user-service/internal/config"
	"github.com/ecommerce/services/user-service/internal/handler"
	"github.com/ecommerce/services/user-service/internal/middleware"
	"github.com/ecommerce/services/user-service/internal/repository"
	"github.com/ecommerce/services/user-service/internal/rpc"
	"github.com/ecommerce/services/user-service/internal/service"
	"github.com/ecommerce/services/user-service/pkg/validator"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := repository.NewDatabase(cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize validator
	userValidator := validator.NewUserValidator()

	// Initialize services
	userService := service.NewUserService(repos.User, userValidator)
	authService := service.NewAuthService(userService)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, userService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Start gRPC server
	go startGRPCServer(cfg, userService, authService)

	// Start HTTP server
	startHTTPServer(cfg, userHandler, authHandler, authMiddleware)
}

func startHTTPServer(cfg *config.Config, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, authMiddleware *middleware.AuthMiddleware) {
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Auth routes (public)
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/login", authHandler.Login).Methods("POST")
	authRoutes.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")
	authRoutes.HandleFunc("/validate", authHandler.ValidateToken).Methods("POST")
	authRoutes.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	authRoutes.HandleFunc("/profile", authHandler.GetProfile).Methods("GET")

	// User routes
	userRoutes := api.PathPrefix("/users").Subrouter()

	// Public user routes
	userRoutes.HandleFunc("", userHandler.CreateUser).Methods("POST") // Register

	// Protected user routes
	protectedUserRoutes := userRoutes.PathPrefix("").Subrouter()
	protectedUserRoutes.Use(authMiddleware.RequireAuth)
	protectedUserRoutes.HandleFunc("", userHandler.ListUsers).Methods("GET")
	protectedUserRoutes.HandleFunc("/{id:[0-9]+}", userHandler.GetUser).Methods("GET")
	protectedUserRoutes.HandleFunc("/{id:[0-9]+}", userHandler.UpdateUser).Methods("PUT")

	// Admin only routes
	adminRoutes := userRoutes.PathPrefix("").Subrouter()
	adminRoutes.Use(authMiddleware.RequireAuth)
	adminRoutes.Use(authMiddleware.RequireRole("admin"))
	adminRoutes.HandleFunc("/{id:[0-9]+}", userHandler.DeleteUser).Methods("DELETE")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "user-service"}`))
	}).Methods("GET")

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("User Service HTTP server starting on %s:%s", cfg.Server.Host, cfg.Server.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("User Service shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("User Service stopped")
}

func startGRPCServer(cfg *config.Config, userService service.UserService, authService service.AuthService) {
	listener, err := net.Listen("tcp", cfg.Server.Host+":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	userGRPCServer := rpc.NewUserServer(userService, authService)
	user_service.RegisterUserServiceServer(grpcServer, userGRPCServer)

	log.Printf("User Service gRPC server starting on %s:%s", cfg.Server.Host, cfg.Server.GRPCPort)

	// Start serving
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}
