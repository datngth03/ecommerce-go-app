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

// Config holds notification service specific configuration
type Config struct {
	Service  sharedConfig.ServiceInfo
	Server   sharedConfig.ServerConfig
	Database sharedConfig.DatabaseConfig
	RabbitMQ sharedConfig.RabbitMQConfig
	Services sharedConfig.ExternalServices
	Logging  sharedConfig.LoggingConfig
	Email    EmailConfig
	SMS      SMSConfig
	Security SecurityConfig
}

// EmailConfig contains email service settings
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

// SMSConfig contains SMS service settings
type SMSConfig struct {
	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Service: sharedConfig.ServiceInfo{
			Name:        sharedConfig.GetEnv("SERVICE_NAME", "notification-service"),
			Version:     sharedConfig.GetEnv("SERVICE_VERSION", "1.0.0"),
			Environment: sharedConfig.GetEnv("ENVIRONMENT", "development"),
		},
		Server:   sharedConfig.LoadServerConfig("notification-service", "8006", "9006"),
		Database: sharedConfig.LoadDatabaseConfig("notification_db"),
		RabbitMQ: sharedConfig.LoadRabbitMQConfig(),
		Services: sharedConfig.LoadExternalServices(),
		Logging:  sharedConfig.LoadLoggingConfig(),
		Email: EmailConfig{
			SMTPHost:     sharedConfig.GetEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     sharedConfig.GetEnv("SMTP_PORT", "587"),
			SMTPUser:     sharedConfig.GetEnv("SMTP_USER", ""),
			SMTPPassword: sharedConfig.GetEnv("SMTP_PASSWORD", ""),
			FromAddress:  sharedConfig.GetEnv("EMAIL_FROM_ADDRESS", "noreply@ecommerce.com"),
			FromName:     sharedConfig.GetEnv("EMAIL_FROM_NAME", "E-Commerce"),
		},
		SMS: SMSConfig{
			TwilioAccountSID: sharedConfig.GetEnv("TWILIO_ACCOUNT_SID", ""),
			TwilioAuthToken:  sharedConfig.GetEnv("TWILIO_AUTH_TOKEN", ""),
			TwilioFromNumber: sharedConfig.GetEnv("TWILIO_FROM_NUMBER", ""),
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
