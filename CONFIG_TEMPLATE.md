# Service Configuration Template
# Sử dụng template này cho tất cả các microservices trong hệ thống

## Common Configuration Structure

```go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	Service  ServiceInfo      // Service metadata
	Server   ServerConfig     // HTTP & gRPC server settings
	Database DatabaseConfig   // PostgreSQL connection
	Redis    RedisConfig      // Redis cache connection
	RabbitMQ RabbitMQConfig   // Message queue connection
	Services ExternalServices // External service endpoints
	Auth     AuthConfig       // Authentication (if needed)
	Logging  LoggingConfig    // Logging configuration
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
}

// RabbitMQConfig contains RabbitMQ connection settings
type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	VHost    string
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
	JWTSecret           string
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	ResetTokenTTL       time.Duration
	EnableAuth          bool
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	Output     string // stdout, file
	FilePath   string
}
```

## Environment Variables Convention

### Server
- `SERVICE_NAME` - Service name (e.g., "user-service", "product-service")
- `SERVICE_VERSION` - Service version (e.g., "1.0.0")
- `ENVIRONMENT` - Environment (development, staging, production)
- `HTTP_PORT` - HTTP server port (e.g., "8081")
- `GRPC_PORT` - gRPC server port (e.g., "9001")
- `SERVER_HOST` - Server host (default: "0.0.0.0")
- `READ_TIMEOUT` - HTTP read timeout in seconds (default: "30")
- `WRITE_TIMEOUT` - HTTP write timeout in seconds (default: "30")
- `SHUTDOWN_TIMEOUT` - Graceful shutdown timeout in seconds (default: "15")

### Database
- `DB_HOST` - PostgreSQL host (default: "localhost")
- `DB_PORT` - PostgreSQL port (default: "5432")
- `DB_USER` - Database user (default: "postgres")
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name (e.g., "users_db", "products_db")
- `DB_SSL_MODE` - SSL mode (disable, require, verify-ca, verify-full)
- `DB_MAX_OPEN_CONNS` - Max open connections (default: "25")
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: "5")
- `DB_CONN_MAX_LIFETIME` - Connection max lifetime in minutes (default: "5")

### Redis
- `REDIS_HOST` - Redis host (default: "localhost")
- `REDIS_PORT` - Redis port (default: "6379")
- `REDIS_PASSWORD` - Redis password (default: "")
- `REDIS_DB` - Redis database number (default: "0")
- `REDIS_POOL_SIZE` - Connection pool size (default: "10")
- `REDIS_MIN_IDLE_CONNS` - Minimum idle connections (default: "5")

### RabbitMQ
- `RABBITMQ_HOST` - RabbitMQ host (default: "localhost")
- `RABBITMQ_PORT` - RabbitMQ port (default: "5672")
- `RABBITMQ_USER` - RabbitMQ user (default: "guest")
- `RABBITMQ_PASSWORD` - RabbitMQ password (default: "guest")
- `RABBITMQ_VHOST` - RabbitMQ virtual host (default: "/")

### External Services
- `USER_SERVICE_GRPC` - User service gRPC address (e.g., "localhost:9001")
- `USER_SERVICE_HTTP` - User service HTTP address (e.g., "http://localhost:8001")
- `USER_SERVICE_TIMEOUT` - Timeout in seconds (default: "30")
- `USER_SERVICE_ENABLED` - Enable/disable service (default: "true")

- `PRODUCT_SERVICE_GRPC` - Product service gRPC address (e.g., "localhost:9002")
- `PRODUCT_SERVICE_HTTP` - Product service HTTP address (e.g., "http://localhost:8002")
- `PRODUCT_SERVICE_TIMEOUT` - Timeout in seconds (default: "30")
- `PRODUCT_SERVICE_ENABLED` - Enable/disable service (default: "true")

- `ORDER_SERVICE_GRPC` - Order service gRPC address (e.g., "localhost:9003")
- `ORDER_SERVICE_HTTP` - Order service HTTP address (e.g., "http://localhost:8003")
- `ORDER_SERVICE_TIMEOUT` - Timeout in seconds (default: "30")
- `ORDER_SERVICE_ENABLED` - Enable/disable service (default: "true")

(Pattern continues for Payment, Inventory, Notification services)

### Authentication (for services that need it)
- `JWT_SECRET` - JWT signing secret
- `JWT_ACCESS_TOKEN_TTL` - Access token TTL in minutes (default: "15")
- `JWT_REFRESH_TOKEN_TTL` - Refresh token TTL in hours (default: "168")
- `JWT_RESET_TOKEN_TTL` - Password reset token TTL in minutes (default: "30")
- `ENABLE_AUTH` - Enable authentication (default: "true")

### Logging
- `LOG_LEVEL` - Log level (debug, info, warn, error) (default: "info")
- `LOG_FORMAT` - Log format (json, text) (default: "json")
- `LOG_OUTPUT` - Log output (stdout, file) (default: "stdout")
- `LOG_FILE_PATH` - Log file path (default: "/var/log/service.log")

## Port Convention

| Service       | HTTP Port | gRPC Port |
|--------------|-----------|-----------|
| API Gateway  | 8000      | -         |
| User Service | 8001      | 9001      |
| Product Service | 8002   | 9002      |
| Order Service | 8003     | 9003      |
| Payment Service | 8004   | 9004      |
| Inventory Service | 8005 | 9005      |
| Notification Service | 8006 | 9006 |

## Helper Functions

```go
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
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

// PrintConfig logs configuration (với masking cho sensitive data)
func (c *Config) PrintConfig() {
	fmt.Printf("=== %s Configuration ===\n", c.Service.Name)
	fmt.Printf("Version: %s\n", c.Service.Version)
	fmt.Printf("Environment: %s\n", c.Service.Environment)
	fmt.Printf("HTTP Port: %s\n", c.Server.HTTPPort)
	fmt.Printf("gRPC Port: %s\n", c.Server.GRPCPort)
	fmt.Printf("Database: %s:%s/%s\n", c.Database.Host, c.Database.Port, c.Database.DBName)
	fmt.Printf("Redis: %s:%s\n", c.Redis.Host, c.Redis.Port)
	fmt.Printf("RabbitMQ: %s:%s\n", c.RabbitMQ.Host, c.RabbitMQ.Port)
	fmt.Println("===========================")
}
```

## Usage Example

```go
// In your service's main.go
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	// Print configuration
	cfg.PrintConfig()
	
	// Use configuration
	db, err := connectPostgres(cfg.GetDatabaseDSN())
	redisClient, err := connectRedis(cfg.GetRedisAddr())
	rabbitMQ, err := connectRabbitMQ(cfg.GetRabbitMQURL())
}
```

## Notes

1. **Mỗi service chỉ cần include các config section cần thiết**
2. **Environment variables should follow 12-factor app principles**
3. **Sensitive data (passwords, secrets) phải được mask khi log**
4. **Defaults should be sensible cho development environment**
5. **Production config phải được set qua environment variables**
