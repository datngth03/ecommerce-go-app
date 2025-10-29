package clients

import (
	"fmt"
	"log"

	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
)

// Clients holds all gRPC client connections with connection pooling
type Clients struct {
	User         *UserClient
	Product      *ProductClient
	Order        *OrderClient
	Payment      *PaymentClient
	Inventory    *InventoryClient
	Notification *NotificationClient
	poolManager  *grpcpool.Manager
}

// NewClients initializes all gRPC clients with connection pooling from config
func NewClients(cfg *config.Config) (*Clients, error) {
	log.Println("🔌 Initializing gRPC clients with connection pooling...")

	// Initialize connection pool manager
	poolManager := grpcpool.NewManager()

	// Create pools for all services
	serviceConfig := &grpcpool.ServicePoolConfig{
		UserServiceTarget:         cfg.Services.UserService.GRPCAddr,
		ProductServiceTarget:      cfg.Services.ProductService.GRPCAddr,
		OrderServiceTarget:        cfg.Services.OrderService.GRPCAddr,
		PaymentServiceTarget:      cfg.Services.PaymentService.GRPCAddr,
		InventoryServiceTarget:    cfg.Services.InventoryService.GRPCAddr,
		NotificationServiceTarget: cfg.Services.NotificationService.GRPCAddr,
		DefaultPoolSize:           5, // 5 connections per service
	}

	if err := poolManager.CreateCommonPools(serviceConfig); err != nil {
		return nil, fmt.Errorf("failed to create connection pools: %w", err)
	}
	log.Println("✅ Connection pools created for all services")

	// Initialize User Client with connection pool
	userPool, exists := poolManager.Get("user-service")
	if !exists {
		poolManager.Close()
		return nil, fmt.Errorf("user service pool not found")
	}
	userClient, err := NewUserClientWithPool(userPool, cfg.Services.UserService.Timeout)
	if err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create user client: %w", err)
	}
	log.Printf("✅ User client initialized with pool (%s)", cfg.Services.UserService.GRPCAddr)

	// Initialize Product Client with connection pool
	productPool, exists := poolManager.Get("product-service")
	if !exists {
		poolManager.Close()
		return nil, fmt.Errorf("product service pool not found")
	}
	productClient, err := NewProductClientWithPool(productPool, cfg.Services.ProductService.Timeout)
	if err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create product client: %w", err)
	}
	log.Printf("✅ Product client initialized with pool (%s)", cfg.Services.ProductService.GRPCAddr)

	// Initialize Order Client with connection pool
	orderPool, exists := poolManager.Get("order-service")
	if !exists {
		poolManager.Close()
		return nil, fmt.Errorf("order service pool not found")
	}
	orderClient, err := NewOrderClientWithPool(orderPool, cfg.Services.OrderService.Timeout)
	if err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create order client: %w", err)
	}
	log.Printf("✅ Order client initialized with pool (%s)", cfg.Services.OrderService.GRPCAddr)

	// Initialize Payment Client with connection pool
	paymentPool, exists := poolManager.Get("payment-service")
	if !exists {
		poolManager.Close()
		return nil, fmt.Errorf("payment service pool not found")
	}
	paymentClient, err := NewPaymentClientWithPool(paymentPool, cfg.Services.PaymentService.Timeout)
	if err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create payment client: %w", err)
	}
	log.Printf("✅ Payment client initialized with pool (%s)", cfg.Services.PaymentService.GRPCAddr)

	// Initialize Inventory Client with connection pool
	inventoryPool, exists := poolManager.Get("inventory-service")
	if !exists {
		poolManager.Close()
		return nil, fmt.Errorf("inventory service pool not found")
	}
	inventoryClient, err := NewInventoryClientWithPool(inventoryPool, cfg.Services.InventoryService.Timeout)
	if err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create inventory client: %w", err)
	}
	log.Printf("✅ Inventory client initialized with pool (%s)", cfg.Services.InventoryService.GRPCAddr)

	// Initialize Notification Client with connection pool
	notificationPool, exists := poolManager.Get("notification-service")
	if !exists {
		poolManager.Close()
		return nil, fmt.Errorf("notification service pool not found")
	}
	notificationClient, err := NewNotificationClientWithPool(notificationPool, cfg.Services.NotificationService.Timeout)
	if err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create notification client: %w", err)
	}
	log.Printf("✅ Notification client initialized with pool (%s)", cfg.Services.NotificationService.GRPCAddr)

	log.Println("✅ All gRPC clients initialized successfully with connection pooling")

	return &Clients{
		User:         userClient,
		Product:      productClient,
		Order:        orderClient,
		Payment:      paymentClient,
		Inventory:    inventoryClient,
		Notification: notificationClient,
		poolManager:  poolManager,
	}, nil
}

// Close closes all client connections and connection pools gracefully
func (c *Clients) Close() error {
	log.Println("🔌 Closing gRPC client connections and pools...")

	// Close the pool manager (which closes all pools and connections)
	if c.poolManager != nil {
		if err := c.poolManager.Close(); err != nil {
			log.Printf("❌ Error closing pool manager: %v", err)
			return fmt.Errorf("failed to close pool manager: %w", err)
		}
	}

	log.Println("✅ All gRPC clients and pools closed")
	return nil
}

// GetPoolStats returns statistics for all connection pools
func (c *Clients) GetPoolStats() map[string]*grpcpool.PoolStats {
	if c.poolManager == nil {
		return nil
	}
	return c.poolManager.GetAllStats()
}
