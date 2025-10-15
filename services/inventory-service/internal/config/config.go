package config

import (
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
)

// Config holds inventory service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	Redis    sharedConfig.RedisConfig
	RabbitMQ sharedConfig.RabbitMQConfig
	Services sharedConfig.ExternalServices
	Logging  sharedConfig.LoggingConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Service: sharedConfig.ServiceInfo{
			Name:        sharedConfig.GetEnv("SERVICE_NAME", "inventory-service"),
			Version:     sharedConfig.GetEnv("SERVICE_VERSION", "1.0.0"),
			Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		},
		Server:   sharedConfig.LoadServerConfig("inventory-service", "8005", "9005"),
		Database: sharedConfig.LoadDatabaseConfig("inventory_db"),
		Redis:    sharedConfig.LoadRedisConfig(),
		RabbitMQ: sharedConfig.LoadRabbitMQConfig(),
		Services: sharedConfig.LoadExternalServices(),
		Logging:  sharedConfig.LoadLoggingConfig(),
	}

	return cfg, nil
}

// GetDatabaseDSN returns PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return c.Database.GetDSN()
}

// GetRedisAddr returns Redis address
func (c *Config) GetRedisAddr() string {
	return c.Redis.GetAddr()
}

// GetRabbitMQURL returns RabbitMQ connection URL
func (c *Config) GetRabbitMQURL() string {
	baseConfig := sharedConfig.Config{
		RabbitMQ: c.RabbitMQ,
	}
	return baseConfig.GetRabbitMQURL()
}

// PrintConfig prints the configuration
func (c *Config) PrintConfig() {
	baseConfig := sharedConfig.Config{
		Service:  c.Service,
		Server:   c.Server,
		Database: c.Database,
		Redis:    c.Redis,
		RabbitMQ: c.RabbitMQ,
		Services: c.Services,
		Logging:  c.Logging,
	}
	baseConfig.PrintConfig()
}
