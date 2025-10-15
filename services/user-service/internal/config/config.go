package config

import (
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
)

// Config holds user service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	Redis    sharedConfig.RedisConfig
	Auth     sharedConfig.AuthConfig
	Logging  sharedConfig.LoggingConfig
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Service: sharedConfig.ServiceInfo{
			Name:        sharedConfig.GetEnv("SERVICE_NAME", "user-service"),
			Version:     sharedConfig.GetEnv("SERVICE_VERSION", "1.0.0"),
			Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		},
		Server:   sharedConfig.LoadServerConfig("user-service", "8001", "9001"),
		Database: sharedConfig.LoadDatabaseConfig("users_db"),
		Redis:    sharedConfig.LoadRedisConfig(),
		Auth:     sharedConfig.LoadAuthConfig(),
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

// PrintConfig prints the configuration
func (c *Config) PrintConfig() {
	baseConfig := sharedConfig.Config{
		Service:  c.Service,
		Server:   c.Server,
		Database: c.Database,
		Redis:    c.Redis,
		Auth:     c.Auth,
		Logging:  c.Logging,
	}
	baseConfig.PrintConfig()
}
