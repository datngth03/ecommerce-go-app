// internal/cart/infrastructure/repository/redis_cart_repo.go
package repository

import (
	"context"
	"encoding/json" // For marshaling/unmarshaling Cart struct to/from JSON
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8" // Redis client for Go

	"github.com/datngth03/ecommerce-go-app/internal/cart/domain"
)

// RedisCartRepository implements the domain.CartRepository interface
// for Redis database operations.
type RedisCartRepository struct {
	client *redis.Client // Redis client instance
	// Optional: TTL for cart items to expire
	cartTTL time.Duration
}

// NewRedisCartRepository creates a new instance of RedisCartRepository.
func NewRedisCartRepository(client *redis.Client, cartTTL time.Duration) *RedisCartRepository {
	if cartTTL == 0 {
		cartTTL = 24 * time.Hour // Default TTL for carts
	}
	return &RedisCartRepository{
		client:  client,
		cartTTL: cartTTL,
	}
}

// cartKey generates the Redis key for a given user's cart.
func (r *RedisCartRepository) cartKey(userID string) string {
	return fmt.Sprintf("cart:%s", userID)
}

// SaveCart saves or updates a user's shopping cart in Redis.
// The entire Cart object is marshaled to JSON and stored.
func (r *RedisCartRepository) SaveCart(ctx context.Context, cart *domain.Cart) error {
	cartJSON, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart to JSON: %w", err)
	}

	key := r.cartKey(cart.UserID)
	// Set the cart in Redis with an expiration time
	if err := r.client.Set(ctx, key, cartJSON, r.cartTTL).Err(); err != nil {
		return fmt.Errorf("failed to save cart to Redis: %w", err)
	}
	return nil
}

// GetCart retrieves a user's shopping cart from Redis by user ID.
// Returns nil and no error if cart is not found.
func (r *RedisCartRepository) GetCart(ctx context.Context, userID string) (*domain.Cart, error) {
	key := r.cartKey(userID)
	cartJSON, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cart not found, return nil without error
		}
		return nil, fmt.Errorf("failed to get cart from Redis: %w", err)
	}

	var cart domain.Cart
	if err := json.Unmarshal(cartJSON, &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart from JSON: %w", err)
	}
	return &cart, nil
}

// DeleteCart removes a user's shopping cart from Redis.
func (r *RedisCartRepository) DeleteCart(ctx context.Context, userID string) error {
	key := r.cartKey(userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cart from Redis: %w", err)
	}
	return nil
}

