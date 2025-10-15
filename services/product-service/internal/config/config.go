// ...existing code...
package config

import (
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
)

// Config holds product service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	Logging  sharedConfig.LoggingConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Service: sharedConfig.ServiceInfo{
			Name:        sharedConfig.GetEnv("SERVICE_NAME", "product-service"),
			Version:     sharedConfig.GetEnv("SERVICE_VERSION", "1.0.0"),
			Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		},
		Server:   sharedConfig.LoadServerConfig("product-service", "8002", "9002"),
		Database: sharedConfig.LoadDatabaseConfig("product_db"),
		Logging:  sharedConfig.LoadLoggingConfig(),
	}

	return cfg, nil
}

// GetDatabaseDSN returns PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return c.Database.GetDSN()
}

// PrintConfig prints the configuration
func (c *Config) PrintConfig() {
	baseConfig := sharedConfig.Config{
		Service:  c.Service,
		Server:   c.Server,
		Database: c.Database,
		Logging:  c.Logging,
	}
	baseConfig.PrintConfig()
}
