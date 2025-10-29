// ...existing code...
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

// Config holds product service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	Logging  sharedConfig.LoggingConfig
	Security SecurityConfig
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
