package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for a microservice
type Config struct {
	Service  ServiceInfo
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
	Services ExternalServices
	Auth     AuthConfig
	Logging  LoggingConfig
}

// ServiceInfo contains service metadata
type ServiceInfo struct {
	Name        string
	Version     string
	Environment string // development, staging, production
}

// ServerConfig contains HTTP and gRPC server settings
type ServerConfig struct {
	HTTPPort        string
	GRPCPort        string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	TLS             TLSConfig
}

// TLSConfig contains TLS/SSL certificate settings
type TLSConfig struct {
	Enabled  bool
	CertFile string // Path to server certificate (.pem)
	KeyFile  string // Path to server private key (.pem)
	CAFile   string // Path to CA certificate for client verification (.pem)
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	Enabled      bool
}

// RabbitMQConfig contains RabbitMQ connection settings
type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
	Enabled  bool
}

// ExternalServices contains addresses of other microservices
type ExternalServices struct {
	UserService         ServiceEndpoint
	ProductService      ServiceEndpoint
	OrderService        ServiceEndpoint
	PaymentService      ServiceEndpoint
	InventoryService    ServiceEndpoint
	NotificationService ServiceEndpoint
}

// ServiceEndpoint represents a microservice endpoint
type ServiceEndpoint struct {
	GRPCAddr string
	HTTPAddr string
	Timeout  time.Duration
	Enabled  bool
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	ResetTokenTTL   time.Duration
	Enabled         bool
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level    string // debug, info, warn, error
	Format   string // json, text
	Output   string // stdout, file
	FilePath string
}

// GetEnv retrieves environment variable with default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt retrieves environment variable as int with default value
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool retrieves environment variable as bool with default value
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvAsDuration retrieves environment variable as duration (in seconds) with default value
func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

// GetEnvAsDurationMinutes retrieves environment variable as duration (in minutes) with default value
func GetEnvAsDurationMinutes(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if minutes, err := strconv.Atoi(value); err == nil {
			return time.Duration(minutes) * time.Minute
		}
	}
	return defaultValue
}

// GetEnvAsDurationHours retrieves environment variable as duration (in hours) with default value
func GetEnvAsDurationHours(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if hours, err := strconv.Atoi(value); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}
	return defaultValue
}

// GetDSN builds PostgreSQL connection string (alias for GetDatabaseDSN)
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.DBName,
		d.SSLMode,
	)
}

// GetAddr returns Redis address (alias for GetRedisAddr)
func (r *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// GetDatabaseDSN builds PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// GetRedisAddr returns Redis address
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// GetRabbitMQURL returns RabbitMQ connection URL
func (c *Config) GetRabbitMQURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		c.RabbitMQ.User,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.VHost,
	)
}

// PrintConfig logs configuration (with masking for sensitive data)
func (c *Config) PrintConfig() {
	fmt.Printf("\n=== %s Configuration ===\n", c.Service.Name)
	fmt.Printf("Version: %s\n", c.Service.Version)
	fmt.Printf("Environment: %s\n", c.Service.Environment)
	fmt.Printf("HTTP Port: %s\n", c.Server.HTTPPort)
	fmt.Printf("gRPC Port: %s\n", c.Server.GRPCPort)

	// Database
	fmt.Printf("\nDatabase:\n")
	fmt.Printf("  Host: %s:%s\n", c.Database.Host, c.Database.Port)
	fmt.Printf("  Database: %s\n", c.Database.DBName)
	fmt.Printf("  User: %s\n", c.Database.User)
	fmt.Printf("  Password: %s\n", maskPassword(c.Database.Password))
	fmt.Printf("  SSL Mode: %s\n", c.Database.SSLMode)

	// Redis
	if c.Redis.Enabled {
		fmt.Printf("\nRedis:\n")
		fmt.Printf("  Address: %s:%s\n", c.Redis.Host, c.Redis.Port)
		fmt.Printf("  Database: %d\n", c.Redis.DB)
		fmt.Printf("  Password: %s\n", maskPassword(c.Redis.Password))
	}

	// RabbitMQ
	if c.RabbitMQ.Enabled {
		fmt.Printf("\nRabbitMQ:\n")
		fmt.Printf("  Address: %s:%s\n", c.RabbitMQ.Host, c.RabbitMQ.Port)
		fmt.Printf("  VHost: %s\n", c.RabbitMQ.VHost)
		fmt.Printf("  User: %s\n", c.RabbitMQ.User)
		fmt.Printf("  Password: %s\n", maskPassword(c.RabbitMQ.Password))
	}

	// External Services
	fmt.Printf("\nExternal Services:\n")
	printServiceEndpoint("User Service", c.Services.UserService)
	printServiceEndpoint("Product Service", c.Services.ProductService)
	printServiceEndpoint("Order Service", c.Services.OrderService)
	printServiceEndpoint("Payment Service", c.Services.PaymentService)
	printServiceEndpoint("Inventory Service", c.Services.InventoryService)
	printServiceEndpoint("Notification Service", c.Services.NotificationService)

	// Auth
	if c.Auth.Enabled {
		fmt.Printf("\nAuthentication:\n")
		fmt.Printf("  JWT Secret: %s\n", maskPassword(c.Auth.JWTSecret))
		fmt.Printf("  Access Token TTL: %v\n", c.Auth.AccessTokenTTL)
		fmt.Printf("  Refresh Token TTL: %v\n", c.Auth.RefreshTokenTTL)
	}

	// Logging
	fmt.Printf("\nLogging:\n")
	fmt.Printf("  Level: %s\n", c.Logging.Level)
	fmt.Printf("  Format: %s\n", c.Logging.Format)
	fmt.Printf("  Output: %s\n", c.Logging.Output)

	fmt.Println("===========================\n")
}

// Helper function to mask passwords
func maskPassword(password string) string {
	if password == "" {
		return "(empty)"
	}
	if len(password) <= 4 {
		return "****"
	}
	return password[:2] + "****" + password[len(password)-2:]
}

// Helper function to print service endpoint
func printServiceEndpoint(name string, endpoint ServiceEndpoint) {
	if endpoint.Enabled {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    gRPC: %s\n", endpoint.GRPCAddr)
		fmt.Printf("    HTTP: %s\n", endpoint.HTTPAddr)
		fmt.Printf("    Timeout: %v\n", endpoint.Timeout)
	}
}

// LoadServerConfig loads common server configuration
func LoadServerConfig(serviceName, defaultHTTPPort, defaultGRPCPort string) ServerConfig {
	return ServerConfig{
		HTTPPort:        GetEnv("HTTP_PORT", defaultHTTPPort),
		GRPCPort:        GetEnv("GRPC_PORT", defaultGRPCPort),
		Host:            GetEnv("SERVER_HOST", "0.0.0.0"),
		ReadTimeout:     GetEnvAsDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    GetEnvAsDuration("WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: GetEnvAsDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		TLS:             LoadTLSConfig(serviceName),
	}
}

// LoadTLSConfig loads TLS configuration from environment variables
func LoadTLSConfig(serviceName string) TLSConfig {
	enabled := GetEnvAsBool("TLS_ENABLED", false)

	// Default certificate paths (relative to project root)
	defaultCertPath := fmt.Sprintf("infrastructure/certs/%s/server-cert.pem", serviceName)
	defaultKeyPath := fmt.Sprintf("infrastructure/certs/%s/server-key.pem", serviceName)
	defaultCAPath := "infrastructure/certs/ca/ca-cert.pem"

	return TLSConfig{
		Enabled:  enabled,
		CertFile: GetEnv("TLS_CERT_FILE", defaultCertPath),
		KeyFile:  GetEnv("TLS_KEY_FILE", defaultKeyPath),
		CAFile:   GetEnv("TLS_CA_FILE", defaultCAPath),
	}
}

// LoadDatabaseConfig loads common database configuration
func LoadDatabaseConfig(defaultDBName string) DatabaseConfig {
	return DatabaseConfig{
		Host:            GetEnv("DB_HOST", "localhost"),
		Port:            GetEnv("DB_PORT", "5432"),
		User:            GetEnv("DB_USER", "postgres"),
		Password:        GetEnv("DB_PASSWORD", "postgres123"),
		DBName:          GetEnv("DB_NAME", defaultDBName),
		SSLMode:         GetEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    GetEnvAsInt("DB_MAX_OPEN_CONNS", 100), // Increased from 25 for load testing
		MaxIdleConns:    GetEnvAsInt("DB_MAX_IDLE_CONNS", 25),  // Increased from 5 for load testing
		ConnMaxLifetime: GetEnvAsDurationMinutes("DB_CONN_MAX_LIFETIME", 5*time.Minute),
	}
}

// LoadRedisConfig loads common Redis configuration
func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         GetEnv("REDIS_HOST", "localhost"),
		Port:         GetEnv("REDIS_PORT", "6379"),
		Password:     GetEnv("REDIS_PASSWORD", ""),
		DB:           GetEnvAsInt("REDIS_DB", 0),
		PoolSize:     GetEnvAsInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: GetEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
		Enabled:      GetEnvAsBool("REDIS_ENABLED", true),
	}
}

// LoadRabbitMQConfig loads common RabbitMQ configuration
func LoadRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host:     GetEnv("RABBITMQ_HOST", "localhost"),
		Port:     GetEnv("RABBITMQ_PORT", "5672"),
		User:     GetEnv("RABBITMQ_USER", "guest"),
		Password: GetEnv("RABBITMQ_PASSWORD", "guest"),
		VHost:    GetEnv("RABBITMQ_VHOST", "/"),
		Enabled:  GetEnvAsBool("RABBITMQ_ENABLED", true),
	}
}

// LoadAuthConfig loads common auth configuration
func LoadAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:       GetEnv("JWT_SECRET", "your-secret-key"),
		AccessTokenTTL:  GetEnvAsDurationMinutes("JWT_ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTokenTTL: GetEnvAsDurationHours("JWT_REFRESH_TOKEN_TTL", 168*time.Hour),
		ResetTokenTTL:   GetEnvAsDurationMinutes("JWT_RESET_TOKEN_TTL", 30*time.Minute),
		Enabled:         GetEnvAsBool("AUTH_ENABLED", true),
	}
}

// LoadLoggingConfig loads common logging configuration
func LoadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:    GetEnv("LOG_LEVEL", "info"),
		Format:   GetEnv("LOG_FORMAT", "json"),
		Output:   GetEnv("LOG_OUTPUT", "stdout"),
		FilePath: GetEnv("LOG_FILE_PATH", "/var/log/service.log"),
	}
}

// LoadServiceEndpoint loads configuration for an external service
func LoadServiceEndpoint(prefix string, defaultGRPC, defaultHTTP string, defaultTimeout time.Duration) ServiceEndpoint {
	return ServiceEndpoint{
		GRPCAddr: GetEnv(prefix+"_GRPC", defaultGRPC),
		HTTPAddr: GetEnv(prefix+"_HTTP", defaultHTTP),
		Timeout:  GetEnvAsDuration(prefix+"_TIMEOUT", defaultTimeout),
		Enabled:  GetEnvAsBool(prefix+"_ENABLED", true),
	}
}

// LoadExternalServices loads all external service configurations
func LoadExternalServices() ExternalServices {
	return ExternalServices{
		UserService: LoadServiceEndpoint(
			"USER_SERVICE",
			"localhost:9001",
			"http://localhost:8001",
			30*time.Second,
		),
		ProductService: LoadServiceEndpoint(
			"PRODUCT_SERVICE",
			"localhost:9002",
			"http://localhost:8002",
			30*time.Second,
		),
		OrderService: LoadServiceEndpoint(
			"ORDER_SERVICE",
			"localhost:9003",
			"http://localhost:8003",
			30*time.Second,
		),
		PaymentService: LoadServiceEndpoint(
			"PAYMENT_SERVICE",
			"localhost:9004",
			"http://localhost:8004",
			30*time.Second,
		),
		InventoryService: LoadServiceEndpoint(
			"INVENTORY_SERVICE",
			"localhost:9005",
			"http://localhost:8005",
			30*time.Second,
		),
		NotificationService: LoadServiceEndpoint(
			"NOTIFICATION_SERVICE",
			"localhost:9006",
			"http://localhost:8006",
			30*time.Second,
		),
	}
}
