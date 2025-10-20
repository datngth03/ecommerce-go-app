package clients

import (
	"fmt"
	"log"

	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/config"
)

// Clients holds all gRPC client connections
type Clients struct {
	User         *UserClient
	Product      *ProductClient
	Order        *OrderClient
	Payment      *PaymentClient
	Inventory    *InventoryClient
	Notification *NotificationClient
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

	// Initialize Order Client
	orderClient, err := NewOrderClient(
		cfg.Services.OrderService.GRPCAddr,
		cfg.Services.OrderService.Timeout,
	)
	if err != nil {
		userClient.Close()
		productClient.Close()
		return nil, fmt.Errorf("failed to create order client: %w", err)
	}
	log.Printf("✅ Order client connected to %s", cfg.Services.OrderService.GRPCAddr)

	// Initialize Payment Client
	paymentClient, err := NewPaymentClient(
		cfg.Services.PaymentService.GRPCAddr,
		cfg.Services.PaymentService.Timeout,
	)
	if err != nil {
		userClient.Close()
		productClient.Close()
		orderClient.Close()
		return nil, fmt.Errorf("failed to create payment client: %w", err)
	}
	log.Printf("✅ Payment client connected to %s", cfg.Services.PaymentService.GRPCAddr)

	// Initialize Inventory Client
	inventoryClient, err := NewInventoryClient(
		cfg.Services.InventoryService.GRPCAddr,
		cfg.Services.InventoryService.Timeout,
	)
	if err != nil {
		userClient.Close()
		productClient.Close()
		orderClient.Close()
		paymentClient.Close()
		return nil, fmt.Errorf("failed to create inventory client: %w", err)
	}
	log.Printf("✅ Inventory client connected to %s", cfg.Services.InventoryService.GRPCAddr)

	// Initialize Notification Client
	notificationClient, err := NewNotificationClient(
		cfg.Services.NotificationService.GRPCAddr,
		cfg.Services.NotificationService.Timeout,
	)
	if err != nil {
		userClient.Close()
		productClient.Close()
		orderClient.Close()
		paymentClient.Close()
		inventoryClient.Close()
		return nil, fmt.Errorf("failed to create notification client: %w", err)
	}
	log.Printf("✅ Notification client connected to %s", cfg.Services.NotificationService.GRPCAddr)

	log.Println("✅ All gRPC clients initialized successfully")

	return &Clients{
		User:         userClient,
		Product:      productClient,
		Order:        orderClient,
		Payment:      paymentClient,
		Inventory:    inventoryClient,
		Notification: notificationClient,
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

	if c.Order != nil {
		if err := c.Order.Close(); err != nil {
			errs = append(errs, fmt.Errorf("order client: %w", err))
		}
	}

	if c.Payment != nil {
		if err := c.Payment.Close(); err != nil {
			errs = append(errs, fmt.Errorf("payment client: %w", err))
		}
	}

	if c.Inventory != nil {
		if err := c.Inventory.Close(); err != nil {
			errs = append(errs, fmt.Errorf("inventory client: %w", err))
		}
	}

	if c.Notification != nil {
		if err := c.Notification.Close(); err != nil {
			errs = append(errs, fmt.Errorf("notification client: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	log.Println("✅ All gRPC clients closed")
	return nil
}
