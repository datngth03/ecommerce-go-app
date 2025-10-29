package grpcpool

import (
	"fmt"
	"sync"

	"google.golang.org/grpc/credentials"
)

// Manager manages multiple connection pools for different services
type Manager struct {
	pools map[string]*ConnectionPool
	mu    sync.RWMutex
}

// NewManager creates a new connection pool manager
func NewManager() *Manager {
	return &Manager{
		pools: make(map[string]*ConnectionPool),
	}
}

// GetOrCreate gets an existing pool or creates a new one
func (m *Manager) GetOrCreate(name string, config *PoolConfig) (*ConnectionPool, error) {
	m.mu.RLock()
	if pool, exists := m.pools[name]; exists {
		m.mu.RUnlock()
		return pool, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists := m.pools[name]; exists {
		return pool, nil
	}

	// Create new pool
	pool, err := NewConnectionPool(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool for %s: %w", name, err)
	}

	m.pools[name] = pool
	return pool, nil
}

// Get retrieves a pool by name
func (m *Manager) Get(name string) (*ConnectionPool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pool, exists := m.pools[name]
	return pool, exists
}

// Close closes all connection pools
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, pool := range m.pools {
		if err := pool.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close pool %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing pools: %v", errs)
	}

	return nil
}

// GetAllStats returns statistics for all pools
func (m *Manager) GetAllStats() map[string]*PoolStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]*PoolStats, len(m.pools))
	for name, pool := range m.pools {
		stats[name] = pool.GetStats()
	}

	return stats
}

// List returns the names of all registered pools
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.pools))
	for name := range m.pools {
		names = append(names, name)
	}

	return names
}

// ServicePoolConfig contains configuration for common service pools
type ServicePoolConfig struct {
	UserServiceTarget           string
	UserServiceTLSCreds         credentials.TransportCredentials
	ProductServiceTarget        string
	ProductServiceTLSCreds      credentials.TransportCredentials
	OrderServiceTarget          string
	OrderServiceTLSCreds        credentials.TransportCredentials
	PaymentServiceTarget        string
	PaymentServiceTLSCreds      credentials.TransportCredentials
	InventoryServiceTarget      string
	InventoryServiceTLSCreds    credentials.TransportCredentials
	NotificationServiceTarget   string
	NotificationServiceTLSCreds credentials.TransportCredentials

	DefaultPoolSize int
	TLSEnabled      bool
}

// CreateCommonPools creates connection pools for all common services
func (m *Manager) CreateCommonPools(config *ServicePoolConfig) error {
	if config.DefaultPoolSize <= 0 {
		config.DefaultPoolSize = 5
	}

	// Map of service name -> (target, tlsCreds)
	services := map[string]struct {
		target   string
		tlsCreds credentials.TransportCredentials
	}{
		"user-service":         {config.UserServiceTarget, config.UserServiceTLSCreds},
		"product-service":      {config.ProductServiceTarget, config.ProductServiceTLSCreds},
		"order-service":        {config.OrderServiceTarget, config.OrderServiceTLSCreds},
		"payment-service":      {config.PaymentServiceTarget, config.PaymentServiceTLSCreds},
		"inventory-service":    {config.InventoryServiceTarget, config.InventoryServiceTLSCreds},
		"notification-service": {config.NotificationServiceTarget, config.NotificationServiceTLSCreds},
	}

	for name, svc := range services {
		if svc.target == "" {
			continue // Skip if target not configured
		}

		poolConfig := DefaultPoolConfig(svc.target)
		poolConfig.PoolSize = config.DefaultPoolSize
		poolConfig.TLSEnabled = config.TLSEnabled
		poolConfig.TLSCreds = svc.tlsCreds // Mỗi service có TLS credentials riêng

		if _, err := m.GetOrCreate(name, poolConfig); err != nil {
			return fmt.Errorf("failed to create pool for %s: %w", name, err)
		}
	}

	return nil
}
