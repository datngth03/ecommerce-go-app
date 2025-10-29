package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
)

// Config holds API Gateway specific configuration
type Config struct {
	Service   sharedConfig.ServiceInfo
	Server    sharedConfig.ServerConfig
	Services  sharedConfig.ExternalServices
	Auth      sharedConfig.AuthConfig
	RateLimit RateLimitConfig
	Logging   sharedConfig.LoggingConfig
	External  ExternalConfig
	Security  SecurityConfig
}

// SecurityConfig contains security middleware settings
type SecurityConfig struct {
	RateLimit      SecurityRateLimitConfig
	CORS           CORSConfig
	RequestTimeout time.Duration
}

// SecurityRateLimitConfig contains rate limiting settings for security middleware
type SecurityRateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	Enabled           bool
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	AllowedOrigins []string
	Enabled        bool
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled        bool
	RequestsPerMin int
	BurstSize      int
}

// ExternalConfig contains external API configurations
type ExternalConfig struct {
	Stripe StripeConfig
	SMTP   SMTPConfig
	Twilio TwilioConfig
}

// StripeConfig contains Stripe API settings
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

// SMTPConfig contains email SMTP settings
type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// TwilioConfig contains Twilio SMS settings
type TwilioConfig struct {
	AccountSID string
	AuthToken  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Service: sharedConfig.ServiceInfo{
			Name:        sharedConfig.GetEnv("SERVICE_NAME", "api-gateway"),
			Version:     sharedConfig.GetEnv("SERVICE_VERSION", "1.0.0"),
			Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		},
		Server:   sharedConfig.LoadServerConfig("api-gateway", "8000", ""),
		Services: sharedConfig.LoadExternalServices(),
		Auth:     sharedConfig.LoadAuthConfig(),
		Logging:  sharedConfig.LoadLoggingConfig(),
		RateLimit: RateLimitConfig{
			Enabled:        sharedConfig.GetEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMin: sharedConfig.GetEnvAsInt("RATE_LIMIT_REQUESTS_PER_MIN", 100),
			BurstSize:      sharedConfig.GetEnvAsInt("RATE_LIMIT_BURST_SIZE", 20),
		},
		External: ExternalConfig{
			Stripe: StripeConfig{
				SecretKey:     sharedConfig.GetEnv("STRIPE_SECRET_KEY", ""),
				WebhookSecret: sharedConfig.GetEnv("STRIPE_WEBHOOK_SECRET", ""),
			},
			SMTP: SMTPConfig{
				Host:     sharedConfig.GetEnv("SMTP_HOST", "smtp.gmail.com"),
				Port:     sharedConfig.GetEnv("SMTP_PORT", "587"),
				User:     sharedConfig.GetEnv("SMTP_USER", ""),
				Password: sharedConfig.GetEnv("SMTP_PASSWORD", ""),
			},
			Twilio: TwilioConfig{
				AccountSID: sharedConfig.GetEnv("TWILIO_ACCOUNT_SID", ""),
				AuthToken:  sharedConfig.GetEnv("TWILIO_AUTH_TOKEN", ""),
			},
		},
		Security: LoadSecurityConfig(),
	}

	return cfg, nil
}

// LoadSecurityConfig loads security middleware configuration
func LoadSecurityConfig() SecurityConfig {
	rateLimitRPS := 50.0
	if rpsStr := sharedConfig.GetEnv("RATE_LIMIT_RPS", "50.0"); rpsStr != "" {
		if parsed, err := strconv.ParseFloat(rpsStr, 64); err == nil {
			rateLimitRPS = parsed
		}
	}

	// Load CORS origins from environment
	corsOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
	if corsEnv := sharedConfig.GetEnv("CORS_ALLOWED_ORIGINS", ""); corsEnv != "" {
		origins := strings.Split(corsEnv, ",")
		corsOrigins = make([]string, 0, len(origins))
		for _, origin := range origins {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				corsOrigins = append(corsOrigins, trimmed)
			}
		}
	}

	return SecurityConfig{
		RateLimit: SecurityRateLimitConfig{
			Enabled:           sharedConfig.GetEnvAsBool("SECURITY_RATE_LIMIT_ENABLED", true),
			RequestsPerSecond: rateLimitRPS,
			BurstSize:         sharedConfig.GetEnvAsInt("SECURITY_RATE_LIMIT_BURST", 100),
		},
		CORS: CORSConfig{
			Enabled:        sharedConfig.GetEnvAsBool("SECURITY_CORS_ENABLED", true),
			AllowedOrigins: corsOrigins,
		},
		RequestTimeout: sharedConfig.GetEnvAsDuration("SECURITY_REQUEST_TIMEOUT", 30*time.Second),
	}
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Service.Environment == "production"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Service.Environment == "development"
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return c.Server.Host + ":" + c.Server.HTTPPort
}

// PrintConfig prints the configuration
func (c *Config) PrintConfig() {
	baseConfig := sharedConfig.Config{
		Service:  c.Service,
		Server:   c.Server,
		Services: c.Services,
		Auth:     c.Auth,
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
