package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	Security SecurityConfig
}

// SecurityConfig contains security middleware settings
type SecurityConfig struct {
	RateLimit      RateLimitConfig
	CORS           CORSConfig
	RequestTimeout time.Duration
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enabled           bool
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	AllowedOrigins []string
	Enabled        bool
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
		Security: LoadSecurityConfig(),
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

	// Print security config
	fmt.Printf("Security:\n")
	fmt.Printf("  Rate Limit:\n")
	fmt.Printf("    Enabled: %v\n", c.Security.RateLimit.Enabled)
	fmt.Printf("    Requests/Second: %.2f\n", c.Security.RateLimit.RequestsPerSecond)
	fmt.Printf("    Burst Size: %d\n", c.Security.RateLimit.BurstSize)
	fmt.Printf("  CORS:\n")
	fmt.Printf("    Enabled: %v\n", c.Security.CORS.Enabled)
	fmt.Printf("    Allowed Origins: %v\n", c.Security.CORS.AllowedOrigins)
	fmt.Printf("  Request Timeout: %v\n", c.Security.RequestTimeout)
}

// LoadSecurityConfig loads security middleware configuration
func LoadSecurityConfig() SecurityConfig {
	rateLimitRPS := 100.0
	if rpsStr := sharedConfig.GetEnv("RATE_LIMIT_RPS", "100.0"); rpsStr != "" {
		if parsed, err := strconv.ParseFloat(rpsStr, 64); err == nil {
			rateLimitRPS = parsed
		}
	}

	// Load CORS allowed origins from environment
	corsOriginsStr := sharedConfig.GetEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	corsOrigins := strings.Split(corsOriginsStr, ",")
	// Trim spaces
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	return SecurityConfig{
		RateLimit: RateLimitConfig{
			Enabled:           sharedConfig.GetEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: rateLimitRPS,
			BurstSize:         sharedConfig.GetEnvAsInt("RATE_LIMIT_BURST", 200),
		},
		CORS: CORSConfig{
			Enabled:        sharedConfig.GetEnvAsBool("CORS_ENABLED", true),
			AllowedOrigins: corsOrigins,
		},
		RequestTimeout: sharedConfig.GetEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second),
	}
}
