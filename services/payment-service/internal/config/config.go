package config

import (
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
)

// Config holds payment service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	RabbitMQ sharedConfig.RabbitMQConfig
	Services sharedConfig.ExternalServices
	Logging  sharedConfig.LoggingConfig
	Payment  PaymentConfig
}

// PaymentConfig contains payment-specific settings
type PaymentConfig struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	PayPalClientID      string
	PayPalClientSecret  string
	Currency            string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Service: sharedConfig.ServiceInfo{
			Name:        sharedConfig.GetEnv("SERVICE_NAME", "payment-service"),
			Version:     sharedConfig.GetEnv("SERVICE_VERSION", "1.0.0"),
			Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		},
		Server:   sharedConfig.LoadServerConfig("payment-service", "8004", "9004"),
		Database: sharedConfig.LoadDatabaseConfig("payment_db"),
		RabbitMQ: sharedConfig.LoadRabbitMQConfig(),
		Services: sharedConfig.LoadExternalServices(),
		Logging:  sharedConfig.LoadLoggingConfig(),
		Payment: PaymentConfig{
			StripeSecretKey:     sharedConfig.GetEnv("STRIPE_SECRET_KEY", ""),
			StripeWebhookSecret: sharedConfig.GetEnv("STRIPE_WEBHOOK_SECRET", ""),
			PayPalClientID:      sharedConfig.GetEnv("PAYPAL_CLIENT_ID", ""),
			PayPalClientSecret:  sharedConfig.GetEnv("PAYPAL_CLIENT_SECRET", ""),
			Currency:            sharedConfig.GetEnv("PAYMENT_CURRENCY", "USD"),
		},
	}

	return cfg, nil
}

// GetDatabaseDSN returns PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return c.Database.GetDSN()
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
		RabbitMQ: c.RabbitMQ,
		Services: c.Services,
		Logging:  c.Logging,
	}
	baseConfig.PrintConfig()
}
