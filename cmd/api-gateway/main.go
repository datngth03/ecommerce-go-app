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

	// Get gRPC address for Order Service from environment variable "ORDER_GRPC_ADDR"
	orderSvcAddr := os.Getenv("ORDER_GRPC_ADDR")
	if orderSvcAddr == "" {
		orderSvcAddr = "localhost:50053" // Default port for Order Service
		log.Printf("No ORDER_GRPC_ADDR environment variable found, using default: %s", orderSvcAddr)
	}

	// Get gRPC address for Payment Service from environment variable "PAYMENT_GRPC_ADDR"
	paymentSvcAddr := os.Getenv("PAYMENT_GRPC_ADDR")
	if paymentSvcAddr == "" {
		paymentSvcAddr = "localhost:50055" // Default port for Payment Service
		log.Printf("No PAYMENT_GRPC_ADDR environment variable found, using default: %s", paymentSvcAddr)
	}

	// Get gRPC address for Cart Service from environment variable "CART_GRPC_ADDR"
	cartSvcAddr := os.Getenv("CART_GRPC_ADDR")
	if cartSvcAddr == "" {
		cartSvcAddr = "localhost:50054" // Default port for Cart Service
		log.Printf("No CART_GRPC_ADDR environment variable found, using default: %s", cartSvcAddr)
	}

	// Get gRPC address for Shipping Service from environment variable "SHIPPING_GRPC_ADDR"
	shippingSvcAddr := os.Getenv("SHIPPING_GRPC_ADDR")
	if shippingSvcAddr == "" {
		shippingSvcAddr = "localhost:50056" // Default port for Shipping Service
		log.Printf("No SHIPPING_GRPC_ADDR environment variable found, using default: %s", shippingSvcAddr)
	}

	// Get gRPC address for Auth Service from environment variable "AUTH_GRPC_ADDR"
	authSvcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authSvcAddr == "" {
		authSvcAddr = "localhost:50057" // Default port for Auth Service
		log.Printf("No AUTH_GRPC_ADDR environment variable found, using default: %s", authSvcAddr)
	}

	// Get gRPC address for Notification Service from environment variable "NOTIFICATION_GRPC_ADDR" (THÊM DÒNG NÀY)
	notificationSvcAddr := os.Getenv("NOTIFICATION_GRPC_ADDR") // THÊM DÒNG NÀY
	if notificationSvcAddr == "" {                             // THÊM DÒNG NÀY
		notificationSvcAddr = "localhost:50058"                                                                    // Default port for Notification Service // THÊM DÒNG NÀY
		log.Printf("No NOTIFICATION_GRPC_ADDR environment variable found, using default: %s", notificationSvcAddr) // THÊM DÒNG NÀY
	}

	// Set Gin to Release Mode for performance optimization and to suppress debug logs
	gin.SetMode(gin.ReleaseMode)

	// Create a default Gin router
	router := gin.Default()

	// Initialize GatewayHandlers with all gRPC clients
	handlers, err := router_http.NewGatewayHandlers(userSvcAddr, productSvcAddr, orderSvcAddr, paymentSvcAddr, cartSvcAddr, shippingSvcAddr, authSvcAddr, notificationSvcAddr)
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
