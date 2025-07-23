// internal/auth/infrastructure/repository/redis_refresh_token_repo.go
package repository

import (
	"context"
	"encoding/json" // For marshaling/unmarshaling RefreshToken struct to/from JSON
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8" // Redis client for Go

	"github.com/datngth03/ecommerce-go-app/internal/auth/domain"
)

// RedisRefreshTokenRepository implements the domain.RefreshTokenRepository interface
// for Redis database operations.
type RedisRefreshTokenRepository struct {
	client *redis.Client // Redis client instance
}

// NewRedisRefreshTokenRepository creates a new instance of RedisRefreshTokenRepository.
func NewRedisRefreshTokenRepository(client *redis.Client) *RedisRefreshTokenRepository {
	return &RedisRefreshTokenRepository{
		client: client,
	}
}

// refreshTokenKey generates the Redis key for a refresh token.
func (r *RedisRefreshTokenRepository) refreshTokenKey(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

// Save stores a refresh token in Redis.
// The RefreshToken object is marshaled to JSON and stored with its expiration time.
func (r *RedisRefreshTokenRepository) Save(ctx context.Context, token *domain.RefreshToken) error {
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token to JSON: %w", err)
	}

	key := r.refreshTokenKey(token.Token)
	// Calculate TTL based on ExpiresAt
	ttl := token.ExpiresAt.Sub(time.Now())
	if ttl <= 0 {
		return errors.New("refresh token expiration time is in the past")
	}

	if err := r.client.Set(ctx, key, tokenJSON, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save refresh token to Redis: %w", err)
	}
	return nil
}

// FindByToken retrieves a refresh token from Redis by its value.
func (r *RedisRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	key := r.refreshTokenKey(token)
	tokenJSON, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("refresh token not found") // Specific error for not found
		}
		return nil, fmt.Errorf("failed to get refresh token from Redis: %w", err)
	}

	var refreshToken domain.RefreshToken
	if err := json.Unmarshal(tokenJSON, &refreshToken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh token from JSON: %w", err)
	}
	return &refreshToken, nil
}

// Delete removes a refresh token from Redis.
func (r *RedisRefreshTokenRepository) Delete(ctx context.Context, token string) error {
	key := r.refreshTokenKey(token)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token from Redis: %w", err)
	}
	return nil
}

// DeleteByUserID removes all refresh tokens for a given user.
// This is a more complex operation in Redis and might require scanning keys.
// For simplicity, this implementation is a placeholder.
// In a real application, you might store refresh tokens in a HASH or SET per user.
func (r *RedisRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID string) error {
	// TODO: Implement actual deletion by user ID.
	// This would involve finding all refresh tokens associated with the user.
	// For now, we'll just log a warning.
	log.Printf("WARNING: DeleteByUserID not fully implemented for RedisRefreshTokenRepository. User ID: %s", userID)
	return nil // Placeholder
}
