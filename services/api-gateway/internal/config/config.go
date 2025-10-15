package config

import (
	sharedConfig "github.com/ecommerce-go-app/shared/pkg/config"
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
	}

	return cfg, nil
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
}
