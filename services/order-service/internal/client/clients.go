package client

import (
	"fmt"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
)

// Clients manages all gRPC clients with connection pooling
type Clients struct {
	poolManager *grpcpool.Manager
	config      *config.Config

	// Individual clients with pool support
	User         *UserClient
	Product      *ProductClient
	Inventory    *InventoryClient
	Payment      *PaymentClient
	Notification *NotificationClient
}

// NewClients creates and initializes all gRPC clients with connection pooling
func NewClients(cfg *config.Config) (*Clients, error) {
	poolManager := grpcpool.NewManager()

	clients := &Clients{
		poolManager: poolManager,
		config:      cfg,
	}

	// Create connection pools for all services
	if err := clients.createPools(); err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to create connection pools: %w", err)
	}

	// Initialize individual clients
	if err := clients.initializeClients(); err != nil {
		poolManager.Close()
		return nil, fmt.Errorf("failed to initialize clients: %w", err)
	}

	return clients, nil
}

// createPools creates connection pools for all required services
func (c *Clients) createPools() error {
	services := map[string]string{
		"user":         c.config.Services.UserService.GRPCAddr,
		"product":      c.config.Services.ProductService.GRPCAddr,
		"inventory":    c.config.Services.InventoryService.GRPCAddr,
		"payment":      c.config.Services.PaymentService.GRPCAddr,
		"notification": c.config.Services.NotificationService.GRPCAddr,
	}

	for serviceName, address := range services {
		poolConfig := grpcpool.DefaultPoolConfig(address)
		if _, err := c.poolManager.GetOrCreate(serviceName, poolConfig); err != nil {
			return fmt.Errorf("failed to create pool for %s service: %w", serviceName, err)
		}
	}

	return nil
}

// initializeClients initializes all client instances with pool support
func (c *Clients) initializeClients() error {
	var err error

	// User Client
	c.User, err = NewUserClientWithPool(c.config.Services.UserService, c.poolManager)
	if err != nil {
		return fmt.Errorf("failed to create user client: %w", err)
	}

	// Product Client
	c.Product, err = NewProductClientWithPool(c.config.Services.ProductService, c.poolManager)
	if err != nil {
		return fmt.Errorf("failed to create product client: %w", err)
	}

	// Inventory Client
	c.Inventory, err = NewInventoryClientWithPool(c.config.Services.InventoryService, c.poolManager)
	if err != nil {
		return fmt.Errorf("failed to create inventory client: %w", err)
	}

	// Payment Client
	c.Payment, err = NewPaymentClientWithPool(c.config.Services.PaymentService, c.poolManager)
	if err != nil {
		return fmt.Errorf("failed to create payment client: %w", err)
	}

	// Notification Client
	c.Notification, err = NewNotificationClientWithPool(c.config.Services.NotificationService, c.poolManager)
	if err != nil {
		return fmt.Errorf("failed to create notification client: %w", err)
	}

	return nil
}

// Close gracefully closes all connection pools
func (c *Clients) Close() error {
	if c.poolManager != nil {
		return c.poolManager.Close()
	}
	return nil
}

// GetPoolStats returns statistics for all connection pools
func (c *Clients) GetPoolStats() map[string]*grpcpool.PoolStats {
	if c.poolManager != nil {
		return c.poolManager.GetAllStats()
	}
	return nil
}
