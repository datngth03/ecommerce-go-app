package config

import (
	"strconv"
	"strings"
	"time"

	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	RateLimit      RateLimitConfig
	CORS           CORSConfig
	RequestTimeout time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enabled           bool
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	Enabled        bool
}

// Config holds payment service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	RabbitMQ sharedConfig.RabbitMQConfig
	Services sharedConfig.ExternalServices
	Logging  sharedConfig.LoggingConfig
	Payment  PaymentConfig
	Security SecurityConfig
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
		Security: LoadSecurityConfig(),
	}

	return cfg, nil
}

// LoadSecurityConfig loads security configuration from environment
func LoadSecurityConfig() SecurityConfig {
	// Parse rate limit RPS
	rpsStr := sharedConfig.GetEnv("SECURITY_RATE_LIMIT_RPS", "50.0")
	rps, err := strconv.ParseFloat(rpsStr, 64)
	if err != nil {
		rps = 50.0
	}

	// Parse rate limit burst
	burstStr := sharedConfig.GetEnv("SECURITY_RATE_LIMIT_BURST", "100")
	burst, err := strconv.Atoi(burstStr)
	if err != nil {
		burst = 100
	}

	// Parse CORS origins
	originsStr := sharedConfig.GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	// Parse timeout
	timeoutStr := sharedConfig.GetEnv("SECURITY_REQUEST_TIMEOUT", "30s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 30 * time.Second
	}

	return SecurityConfig{
		RateLimit: RateLimitConfig{
			RequestsPerSecond: rps,
			BurstSize:         burst,
			Enabled:           sharedConfig.GetEnv("SECURITY_RATE_LIMIT_ENABLED", "true") == "true",
		},
		CORS: CORSConfig{
			AllowedOrigins: origins,
			Enabled:        sharedConfig.GetEnv("SECURITY_CORS_ENABLED", "true") == "true",
		},
		RequestTimeout: timeout,
	}
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
