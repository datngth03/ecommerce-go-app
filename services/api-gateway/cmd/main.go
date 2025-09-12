package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/ecommerce/services/api-gateway/internal/config"
	"github.com/ecommerce/services/api-gateway/internal/handler"
	"github.com/ecommerce/services/api-gateway/internal/middleware"
	"github.com/ecommerce/services/api-gateway/internal/proxy"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize service proxy
	serviceProxy := proxy.NewServiceProxy(cfg)

	// Initialize handlers
	proxyHandler := handler.NewProxyHandler(serviceProxy)

	// Setup router
	router := mux.NewRouter()

	// Setup middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.RateLimitMiddleware)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure properly in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	router.Use(c.Handler)

	// API Routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// User Service Routes
	userRoutes := api.PathPrefix("/users").Subrouter()
	userRoutes.HandleFunc("/register", proxyHandler.ProxyToUserService).Methods("POST")
	userRoutes.HandleFunc("/login", proxyHandler.ProxyToUserService).Methods("POST")
	userRoutes.HandleFunc("/profile", proxyHandler.ProxyToUserServiceWithAuth).Methods("GET")
	userRoutes.HandleFunc("/refresh", proxyHandler.ProxyToUserService).Methods("POST")
	userRoutes.HandleFunc("/logout", proxyHandler.ProxyToUserService).Methods("POST")
	userRoutes.HandleFunc("", proxyHandler.ProxyToUserServiceWithAuth).Methods("GET") // List users
	userRoutes.HandleFunc("", proxyHandler.ProxyToUserService).Methods("POST")        // Create user
	userRoutes.HandleFunc("/{id:[0-9]+}", proxyHandler.ProxyToUserService).Methods("GET", "PUT", "DELETE")

	// Auth Routes (separate from users)
	authRoutes := api.PathPrefix("/auth").Subrouter()
	authRoutes.HandleFunc("/login", proxyHandler.ProxyToUserService).Methods("POST")
	authRoutes.HandleFunc("/register", proxyHandler.ProxyToUserService).Methods("POST")
	authRoutes.HandleFunc("/refresh", proxyHandler.ProxyToUserService).Methods("POST")
	authRoutes.HandleFunc("/validate", proxyHandler.ProxyToUserService).Methods("POST")
	authRoutes.HandleFunc("/logout", proxyHandler.ProxyToUserServiceWithAuth).Methods("POST")

	// Product Service Routes (example)
	productRoutes := api.PathPrefix("/products").Subrouter()
	productRoutes.HandleFunc("", proxyHandler.ProxyToProductService).Methods("GET", "POST")
	productRoutes.HandleFunc("/{id:[0-9]+}", proxyHandler.ProxyToProductService).Methods("GET", "PUT", "DELETE")

	// Order Service Routes (example)
	orderRoutes := api.PathPrefix("/orders").Subrouter()
	orderRoutes.Use(middleware.AuthMiddleware(serviceProxy)) // All order routes require auth
	orderRoutes.HandleFunc("", proxyHandler.ProxyToOrderService).Methods("GET", "POST")
	orderRoutes.HandleFunc("/{id:[0-9]+}", proxyHandler.ProxyToOrderService).Methods("GET", "PUT", "DELETE")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("API Gateway starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("API Gateway shutting down...")
}
