package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	HTTPPort string
	GRPCPort string
	Host     string
	Timeout  time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration // Sửa lại
	RefreshTokenTTL time.Duration // Thêm vào
	ResetTokenTTL   time.Duration // Thêm vào
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func Load() (*Config, error) {
	timeout, err := strconv.Atoi(getEnv("SERVER_TIMEOUT", "30"))
	if err != nil {
		timeout = 30
	}

	accessTokenTTL, _ := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_TTL_MINUTES", "15"))
	refreshTokenTTL, _ := strconv.Atoi(getEnv("JWT_REFRESH_TOKEN_TTL_HOURS", "168"))
	resetTokenTTL, _ := strconv.Atoi(getEnv("JWT_RESET_TOKEN_TTL_MINUTES", "30"))

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	return &Config{
		Server: ServerConfig{
			HTTPPort: getEnv("HTTP_PORT", "8001"),
			GRPCPort: getEnv("GRPC_PORT", "9001"),
			Host:     getEnv("HOST", "0.0.0.0"),
			Timeout:  time.Duration(timeout) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "users_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key"),
			AccessTokenTTL:  time.Duration(accessTokenTTL) * time.Minute,
			RefreshTokenTTL: time.Duration(refreshTokenTTL) * time.Hour,
			ResetTokenTTL:   time.Duration(resetTokenTTL) * time.Minute,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
	}, nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.Port,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
