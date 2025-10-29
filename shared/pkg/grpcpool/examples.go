package grpcpool

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

// Example: Using connection pool in a service client

// UserServiceClient demonstrates how to use the connection pool
type ExampleUserServiceClient struct {
	pool *ConnectionPool
}

// NewExampleUserServiceClient creates a new user service client with connection pooling
func NewExampleUserServiceClient(target string) (*ExampleUserServiceClient, error) {
	config := DefaultPoolConfig(target)
	config.PoolSize = 5 // 5 connections for user service

	pool, err := NewConnectionPool(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &ExampleUserServiceClient{
		pool: pool,
	}, nil
}

// GetConnection returns a connection from the pool
func (c *ExampleUserServiceClient) GetConnection() *grpc.ClientConn {
	return c.pool.Get()
}

// GetHealthyConnection returns a healthy connection with context timeout
func (c *ExampleUserServiceClient) GetHealthyConnection(ctx context.Context) (*grpc.ClientConn, error) {
	return c.pool.GetHealthy(ctx)
}

// Close closes the connection pool
func (c *ExampleUserServiceClient) Close() error {
	return c.pool.Close()
}

// Example usage in a handler:
/*
func (h *Handler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Get a connection from the pool
	conn := h.userServiceClient.GetConnection()

	// Create the service client
	client := pb.NewUserServiceClient(conn)

	// Make the call
	return client.GetUser(ctx, req)
}
*/

// Example: Using Manager for multiple services

// APIGatewayClients manages all service clients with connection pooling
type ExampleAPIGatewayClients struct {
	manager *Manager
}

// NewExampleAPIGatewayClients creates clients for all services
func NewExampleAPIGatewayClients() (*ExampleAPIGatewayClients, error) {
	manager := NewManager()

	// Create pools for all services
	serviceConfig := &ServicePoolConfig{
		UserServiceTarget:         "user-service:50051",
		ProductServiceTarget:      "product-service:50052",
		OrderServiceTarget:        "order-service:50053",
		PaymentServiceTarget:      "payment-service:50054",
		InventoryServiceTarget:    "inventory-service:50055",
		NotificationServiceTarget: "notification-service:50056",
		DefaultPoolSize:           5,
	}

	if err := manager.CreateCommonPools(serviceConfig); err != nil {
		return nil, fmt.Errorf("failed to create service pools: %w", err)
	}

	return &ExampleAPIGatewayClients{
		manager: manager,
	}, nil
}

// GetUserServiceConn returns a connection to user service
func (c *ExampleAPIGatewayClients) GetUserServiceConn() (*grpc.ClientConn, error) {
	pool, exists := c.manager.Get("user-service")
	if !exists {
		return nil, fmt.Errorf("user service pool not found")
	}
	return pool.Get(), nil
}

// GetProductServiceConn returns a connection to product service
func (c *ExampleAPIGatewayClients) GetProductServiceConn() (*grpc.ClientConn, error) {
	pool, exists := c.manager.Get("product-service")
	if !exists {
		return nil, fmt.Errorf("product service pool not found")
	}
	return pool.Get(), nil
}

// Close closes all service connections
func (c *ExampleAPIGatewayClients) Close() error {
	return c.manager.Close()
}

// GetStats returns statistics for all service pools
func (c *ExampleAPIGatewayClients) GetStats() map[string]*PoolStats {
	return c.manager.GetAllStats()
}

// Example: Health check endpoint
/*
func (h *Handler) HealthCheck(c *gin.Context) {
	stats := h.clients.GetStats()

	response := gin.H{
		"status": "healthy",
		"services": make(map[string]interface{}),
	}

	allHealthy := true
	for name, stat := range stats {
		serviceStatus := gin.H{
			"healthy_percentage": stat.HealthyPercentage(),
			"ready_connections": stat.ReadyCount,
			"total_connections": stat.PoolSize,
		}

		if !stat.IsHealthy() {
			allHealthy = false
			serviceStatus["status"] = "unhealthy"
		} else {
			serviceStatus["status"] = "healthy"
		}

		response["services"].(map[string]interface{})[name] = serviceStatus
	}

	if !allHealthy {
		response["status"] = "degraded"
		c.JSON(503, response)
		return
	}

	c.JSON(200, response)
}
*/

// Example: Retry with connection pool
func ExampleRetryWithPool(ctx context.Context, pool *ConnectionPool, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		// Get a healthy connection
		_, err := pool.GetHealthy(ctx)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
			continue
		}

		// Try to make the call
		// conn := pool.Get()
		// client := pb.NewServiceClient(conn)
		// resp, err := client.Method(ctx, req)
		// if err == nil {
		// 	return nil
		// }

		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// Example: Using with interceptors
/*
func NewPoolWithInterceptors(target string) (*ConnectionPool, error) {
	config := DefaultPoolConfig(target)

	// Add unary interceptor
	unaryInterceptor := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		log.Printf("gRPC call: %s, duration: %v, error: %v", method, duration, err)
		return err
	}

	// Add to dial options
	config.DialOptions = append(config.DialOptions,
		grpc.WithUnaryInterceptor(unaryInterceptor),
	)

	return NewConnectionPool(config)
}
*/
