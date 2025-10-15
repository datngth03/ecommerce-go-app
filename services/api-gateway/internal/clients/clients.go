package clients

import (
	"fmt"
	"log"

	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/config"
)

// Clients holds all gRPC client connections
type Clients struct {
	User    *UserClient
	Product *ProductClient
	// Order    *OrderClient    // TODO: Implement when ready
	// Payment  *PaymentClient  // TODO: Implement when ready
}

// NewClients initializes all gRPC clients from config
func NewClients(cfg *config.Config) (*Clients, error) {
	log.Println("🔌 Initializing gRPC clients...")

	// Initialize User Client
	userClient, err := NewUserClient(
		cfg.Services.UserService.GRPCAddr,
		cfg.Services.UserService.Timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user client: %w", err)
	}
	log.Printf("✅ User client connected to %s", cfg.Services.UserService.GRPCAddr)

	// Initialize Product Client
	productClient, err := NewProductClient(
		cfg.Services.ProductService.GRPCAddr,
		cfg.Services.ProductService.Timeout,
	)
	if err != nil {
		userClient.Close()
		return nil, fmt.Errorf("failed to create product client: %w", err)
	}
	log.Printf("✅ Product client connected to %s", cfg.Services.ProductService.GRPCAddr)

	log.Println("✅ All gRPC clients initialized successfully")

	return &Clients{
		User:    userClient,
		Product: productClient,
	}, nil
}

// Close closes all client connections gracefully
func (c *Clients) Close() error {
	log.Println("🔌 Closing gRPC client connections...")

	var errs []error

	if c.User != nil {
		if err := c.User.Close(); err != nil {
			errs = append(errs, fmt.Errorf("user client: %w", err))
		}
	}

	if c.Product != nil {
		if err := c.Product.Close(); err != nil {
			errs = append(errs, fmt.Errorf("product client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	log.Println("✅ All gRPC clients closed")
	return nil
}
