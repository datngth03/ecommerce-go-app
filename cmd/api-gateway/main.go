// cmd/api-gateway/main.go
package main

import (
	"fmt"
	"log"
	"os" // To read environment variables

	"github.com/gin-gonic/gin" // Import Gin Gonic framework
	"github.com/joho/godotenv" // To read environment variables from .env file

	// Import router and handlers of the API Gateway from internal
	router_http "github.com/datngth03/ecommerce-go-app/internal/api_gateway/delivery/http"
)

func main() {
	// Load environment variables from .env file (if any)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Get HTTP port for the Gateway from environment variable "HTTP_PORT"
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	// Get gRPC address for User Service from environment variable "USER_GRPC_ADDR"
	userSvcAddr := os.Getenv("USER_GRPC_ADDR")
	if userSvcAddr == "" {
		userSvcAddr = "localhost:50051" // Default port for User Service
		log.Printf("No USER_GRPC_ADDR environment variable found, using default: %s", userSvcAddr)
	}

	// Get gRPC address for Product Service from environment variable "PRODUCT_GRPC_ADDR"
	productSvcAddr := os.Getenv("PRODUCT_GRPC_ADDR")
	if productSvcAddr == "" {
		productSvcAddr = "localhost:50052" // Default port for Product Service
		log.Printf("No PRODUCT_GRPC_ADDR environment variable found, using default: %s", productSvcAddr)
	}

	// Set Gin to Release Mode for performance optimization and to suppress debug logs
	gin.SetMode(gin.ReleaseMode)

	// Create a default Gin router
	router := gin.Default()

	// Initialize GatewayHandlers with gRPC clients
	handlers, err := router_http.NewGatewayHandlers(userSvcAddr, productSvcAddr)
	if err != nil {
		log.Fatalf("Failed to initialize Gateway Handlers: %v", err)
	}
	// TODO: Ensure gRPC connections are closed when the application exits
	// Currently, gRPC connections created in NewGatewayHandlers are not explicitly closed.
	// In a production application, you should manage their lifecycle.

	// Register routes for the API Gateway
	router_http.RegisterRoutes(router, handlers)

	// Print server startup message
	fmt.Printf("API Gateway listening on port :%s...\n", httpPort)

	// Start the HTTP server
	if err := router.Run(fmt.Sprintf(":%s", httpPort)); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
