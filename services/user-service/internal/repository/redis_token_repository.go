package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/user-service/pkg/utils"
	"github.com/redis/go-redis/v9"
)

// RedisTokenRepository implements TokenRepositoryInterface using Redis.
type RedisTokenRepository struct {
	client *redis.Client
}

// NewRedisTokenRepository creates a new instance of RedisTokenRepository.
func NewRedisTokenRepository(client *redis.Client) TokenRepositoryInterface {
	return &RedisTokenRepository{
		client: client,
	}
}

// --- Key Generation Helpers ---
func keyRefreshToken(token string) string {
	return fmt.Sprintf("refresh_token:%s", token)
}

func keyUserTokensSet(userID int64) string {
	return fmt.Sprintf("user_tokens:%d", userID)
}

func keyBlacklist(token string) string {
	return fmt.Sprintf("blacklist:%s", token)
}

func keyResetToken(token string) string {
	return fmt.Sprintf("reset_token:%s", token)
}

// =================================
// Refresh Token Implementation
// =================================

// StoreRefreshToken stores a refresh token and adds it to the user's token set.
func (r *RedisTokenRepository) StoreRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	pipe := r.client.Pipeline()

	// 1. Store the main token with userID as value and a TTL.
	ttl := time.Until(expiresAt)
	pipe.Set(ctx, keyRefreshToken(token), userID, ttl)

	// 2. Add the token to the user's set to track all their tokens.
	pipe.SAdd(ctx, keyUserTokensSet(userID), token)

	// Execute both commands atomically.
	_, err := pipe.Exec(ctx)
	return err
}

// GetRefreshToken retrieves refresh token data from Redis.
func (r *RedisTokenRepository) GetRefreshToken(ctx context.Context, token string) (*utils.RefreshTokenData, error) {
	key := keyRefreshToken(token)
	userIDStr, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("token not found")
		}
		return nil, err
	}

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	return &utils.RefreshTokenData{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// DeleteRefreshToken deletes a refresh token and removes it from the user's token set.
func (r *RedisTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	// Get userID from the token before deleting it to know which set to remove from.
	data, err := r.GetRefreshToken(ctx, token)
	if err != nil {
		if errors.Is(err, errors.New("token not found")) {
			return nil // If token doesn't exist, the goal is achieved.
		}
		return err
	}

	pipe := r.client.Pipeline()
	// 1. Delete the token key.
	pipe.Del(ctx, keyRefreshToken(token))
	// 2. Remove the token from the user's set.
	pipe.SRem(ctx, keyUserTokensSet(data.UserID), token)

	_, err = pipe.Exec(ctx)
	return err
}

// DeleteAllUserRefreshTokens deletes all refresh tokens for a specific user.
func (r *RedisTokenRepository) DeleteAllUserRefreshTokens(ctx context.Context, userID int64) error {
	setKey := keyUserTokensSet(userID)
	// 1. Get all tokens from the user's set.
	tokens, err := r.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	// 2. Delete each individual token key.
	for _, token := range tokens {
		pipe.Del(ctx, keyRefreshToken(token))
	}
	// 3. Delete the set itself.
	pipe.Del(ctx, setKey)

	_, err = pipe.Exec(ctx)
	return err
}

// =================================
// Access Token Blacklist Implementation
// =================================

// IsTokenBlacklisted checks if an access token exists in the blacklist.
func (r *RedisTokenRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	val, err := r.client.Exists(ctx, keyBlacklist(token)).Result()
	if err != nil {
		return false, err
	}
	return val == 1, nil
}

// BlacklistToken adds an access token to the blacklist with a TTL.
func (r *RedisTokenRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	// The value doesn't matter, only the key's existence.
	return r.client.Set(ctx, keyBlacklist(token), "1", ttl).Err()
}

// =================================
// Password Reset Implementation
// =================================

// StorePasswordResetToken stores a password reset token.
func (r *RedisTokenRepository) StorePasswordResetToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	return r.client.Set(ctx, keyResetToken(token), userID, ttl).Err()
}

// GetPasswordResetToken retrieves data for a given password reset token.
func (r *RedisTokenRepository) GetPasswordResetToken(ctx context.Context, token string) (*utils.PasswordResetTokenData, error) {
	key := keyResetToken(token)
	userIDStr, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("token not found")
		}
		return nil, err
	}

	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	return &utils.PasswordResetTokenData{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// DeletePasswordResetToken removes a password reset token after it has been used.
func (r *RedisTokenRepository) DeletePasswordResetToken(ctx context.Context, token string) error {
	err := r.client.Del(ctx, keyResetToken(token)).Err()
	if err == redis.Nil {
		return nil
	}
	return err
}
