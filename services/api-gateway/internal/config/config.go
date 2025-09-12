package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Services ServiceConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type ServiceConfig struct {
	UserService         ServiceEndpoint
	ProductService      ServiceEndpoint
	OrderService        ServiceEndpoint
	PaymentService      ServiceEndpoint
	InventoryService    ServiceEndpoint
	NotificationService ServiceEndpoint
}

type ServiceEndpoint struct {
	HTTPHost string
	HTTPPort string
	GRPCHost string
	GRPCPort string
	Timeout  time.Duration
}

type AuthConfig struct {
	JWTSecret string
}

func Load() (*Config, error) {
	readTimeout, _ := strconv.Atoi(getEnv("READ_TIMEOUT", "15"))
	writeTimeout, _ := strconv.Atoi(getEnv("WRITE_TIMEOUT", "15"))
	serviceTimeout, _ := strconv.Atoi(getEnv("SERVICE_TIMEOUT", "10"))

	return &Config{
		Server: ServerConfig{
			Port:         getEnv("GATEWAY_PORT", "8080"),
			Host:         getEnv("GATEWAY_HOST", "0.0.0.0"),
			ReadTimeout:  time.Duration(readTimeout) * time.Second,
			WriteTimeout: time.Duration(writeTimeout) * time.Second,
		},
		Services: ServiceConfig{
			UserService: ServiceEndpoint{
				HTTPHost: getEnv("USER_SERVICE_HTTP_HOST", "localhost"),
				HTTPPort: getEnv("USER_SERVICE_HTTP_PORT", "8001"),
				GRPCHost: getEnv("USER_SERVICE_GRPC_HOST", "localhost"),
				GRPCPort: getEnv("USER_SERVICE_GRPC_PORT", "9001"),
				Timeout:  time.Duration(serviceTimeout) * time.Second,
			},
			ProductService: ServiceEndpoint{
				HTTPHost: getEnv("PRODUCT_SERVICE_HTTP_HOST", "localhost"),
				HTTPPort: getEnv("PRODUCT_SERVICE_HTTP_PORT", "8002"),
				GRPCHost: getEnv("PRODUCT_SERVICE_GRPC_HOST", "localhost"),
				GRPCPort: getEnv("PRODUCT_SERVICE_GRPC_PORT", "9002"),
				Timeout:  time.Duration(serviceTimeout) * time.Second,
			},
			OrderService: ServiceEndpoint{
				HTTPHost: getEnv("ORDER_SERVICE_HTTP_HOST", "localhost"),
				HTTPPort: getEnv("ORDER_SERVICE_HTTP_PORT", "8003"),
				GRPCHost: getEnv("ORDER_SERVICE_GRPC_HOST", "localhost"),
				GRPCPort: getEnv("ORDER_SERVICE_GRPC_PORT", "9003"),
				Timeout:  time.Duration(serviceTimeout) * time.Second,
			},
			PaymentService: ServiceEndpoint{
				HTTPHost: getEnv("PAYMENT_SERVICE_HTTP_HOST", "localhost"),
				HTTPPort: getEnv("PAYMENT_SERVICE_HTTP_PORT", "8004"),
				GRPCHost: getEnv("PAYMENT_SERVICE_GRPC_HOST", "localhost"),
				GRPCPort: getEnv("PAYMENT_SERVICE_GRPC_PORT", "9004"),
				Timeout:  time.Duration(serviceTimeout) * time.Second,
			},
			InventoryService: ServiceEndpoint{
				HTTPHost: getEnv("INVENTORY_SERVICE_HTTP_HOST", "localhost"),
				HTTPPort: getEnv("INVENTORY_SERVICE_HTTP_PORT", "8005"),
				GRPCHost: getEnv("INVENTORY_SERVICE_GRPC_HOST", "localhost"),
				GRPCPort: getEnv("INVENTORY_SERVICE_GRPC_PORT", "9005"),
				Timeout:  time.Duration(serviceTimeout) * time.Second,
			},
			NotificationService: ServiceEndpoint{
				HTTPHost: getEnv("NOTIFICATION_SERVICE_HTTP_HOST", "localhost"),
				HTTPPort: getEnv("NOTIFICATION_SERVICE_HTTP_PORT", "8006"),
				GRPCHost: getEnv("NOTIFICATION_SERVICE_GRPC_HOST", "localhost"),
				GRPCPort: getEnv("NOTIFICATION_SERVICE_GRPC_PORT", "9006"),
				Timeout:  time.Duration(serviceTimeout) * time.Second,
			},
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
		},
	}, nil
}

func (s ServiceEndpoint) GetHTTPURL() string {
	return "http://" + s.HTTPHost + ":" + s.HTTPPort
}

func (s ServiceEndpoint) GetGRPCAddr() string {
	return s.GRPCHost + ":" + s.GRPCPort
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
