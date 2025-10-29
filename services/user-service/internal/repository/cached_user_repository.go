package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/cache"
)

// CachedUserRepository wraps UserRepositoryInterface with Redis caching
type CachedUserRepository struct {
	repo  UserRepositoryInterface
	cache *cache.RedisCache
}

// Cache TTL constants for users
const (
	UserCacheTTL    = 5 * time.Minute  // Individual user cache
	ProfileCacheTTL = 10 * time.Minute // User profiles (less frequent changes)
	EmailLookupTTL  = 5 * time.Minute  // Email to user mapping
)

// NewCachedUserRepository creates a cached user repository
func NewCachedUserRepository(repo UserRepositoryInterface, cache *cache.RedisCache) *CachedUserRepository {
	return &CachedUserRepository{
		repo:  repo,
		cache: cache,
	}
}

// Create creates a new user and caches it
func (r *CachedUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	// Create in database
	createdUser, err := r.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Cache the newly created user
	cacheKeyID := fmt.Sprintf("user:id:%d", createdUser.ID)
	cacheKeyEmail := fmt.Sprintf("user:email:%s", createdUser.Email)

	// Cache by ID
	if err := r.cache.Set(ctx, cacheKeyID, createdUser, UserCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache new user ID %d: %v\n", createdUser.ID, err)
	}

	// Cache by email
	if err := r.cache.Set(ctx, cacheKeyEmail, createdUser, EmailLookupTTL); err != nil {
		fmt.Printf("Warning: failed to cache new user email %s: %v\n", createdUser.Email, err)
	}

	return createdUser, nil
}

// GetByID retrieves a user by ID with caching
func (r *CachedUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:id:%d", id)

	var user models.User

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for user ID %d: %v\n", id, err)
	}

	// Cache miss - fetch from DB
	dbUser, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, dbUser, UserCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache user ID %d: %v\n", id, err)
	}

	return dbUser, nil
}

// GetByEmail retrieves a user by email with caching
func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)

	var user models.User

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for user email %s: %v\n", email, err)
	}

	// Cache miss - fetch from DB
	dbUser, err := r.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Cache the result (both by email and by ID)
	if err := r.cache.Set(ctx, cacheKey, dbUser, EmailLookupTTL); err != nil {
		fmt.Printf("Warning: failed to cache user email %s: %v\n", email, err)
	}

	// Also cache by ID for future lookups
	cacheKeyID := fmt.Sprintf("user:id:%d", dbUser.ID)
	if err := r.cache.Set(ctx, cacheKeyID, dbUser, UserCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache user ID %d: %v\n", dbUser.ID, err)
	}

	return dbUser, nil
}

// Update updates a user and invalidates its caches
func (r *CachedUserRepository) Update(ctx context.Context, updateData *models.UserUpdateData) (*models.User, error) {
	// Get current user data for cache invalidation
	currentUser, err := r.repo.GetByID(ctx, updateData.ID)
	if err != nil {
		return nil, err
	}

	// Update in database
	updatedUser, err := r.repo.Update(ctx, updateData)
	if err != nil {
		return nil, err
	}

	// Invalidate caches
	keysToDelete := []string{
		fmt.Sprintf("user:id:%d", updatedUser.ID),
		fmt.Sprintf("user:email:%s", currentUser.Email),
	}

	// If email changed, also invalidate new email cache
	if updatedUser.Email != currentUser.Email {
		keysToDelete = append(keysToDelete, fmt.Sprintf("user:email:%s", updatedUser.Email))
	}

	if err := r.cache.Delete(ctx, keysToDelete...); err != nil {
		fmt.Printf("Warning: failed to invalidate user caches: %v\n", err)
	}

	// Cache the updated user
	cacheKeyID := fmt.Sprintf("user:id:%d", updatedUser.ID)
	cacheKeyEmail := fmt.Sprintf("user:email:%s", updatedUser.Email)

	if err := r.cache.Set(ctx, cacheKeyID, updatedUser, UserCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache updated user: %v\n", err)
	}

	if err := r.cache.Set(ctx, cacheKeyEmail, updatedUser, EmailLookupTTL); err != nil {
		fmt.Printf("Warning: failed to cache updated user email: %v\n", err)
	}

	return updatedUser, nil
}

// Delete deletes a user and invalidates its caches
func (r *CachedUserRepository) Delete(ctx context.Context, id int64) error {
	// Get user first for cache invalidation
	user, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate caches
	keysToDelete := []string{
		fmt.Sprintf("user:id:%d", id),
		fmt.Sprintf("user:email:%s", user.Email),
	}

	if err := r.cache.Delete(ctx, keysToDelete...); err != nil {
		fmt.Printf("Warning: failed to invalidate deleted user caches: %v\n", err)
	}

	return nil
}

// UpdatePassword updates a user's password and invalidates caches
func (r *CachedUserRepository) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	if err := r.repo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}

	// Invalidate user cache (password change might affect user data)
	cacheKey := fmt.Sprintf("user:id:%d", userID)
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: failed to invalidate user cache after password update: %v\n", err)
	}

	return nil
}

// ExistsByEmail checks if user exists by email (no caching - security sensitive)
func (r *CachedUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// Don't cache existence checks - they're security-sensitive
	// and need to be real-time accurate
	return r.repo.ExistsByEmail(ctx, email)
}

// GetProfile retrieves a user profile (extended cache TTL)
func (r *CachedUserRepository) GetProfile(ctx context.Context, id int64) (*models.User, error) {
	cacheKey := fmt.Sprintf("user:profile:%d", id)

	var user models.User

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for user profile %d: %v\n", id, err)
	}

	// Fetch from DB
	dbUser, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache profile with longer TTL (profiles change less frequently)
	if err := r.cache.Set(ctx, cacheKey, dbUser, ProfileCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache user profile: %v\n", err)
	}

	return dbUser, nil
}

// InvalidateUserCache manually invalidates all caches for a user
func (r *CachedUserRepository) InvalidateUserCache(ctx context.Context, userID int64) error {
	// Get user to know email for invalidation
	user, err := r.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	keysToDelete := []string{
		fmt.Sprintf("user:id:%d", userID),
		fmt.Sprintf("user:email:%s", user.Email),
		fmt.Sprintf("user:profile:%d", userID),
	}

	return r.cache.Delete(ctx, keysToDelete...)
}

// WarmupCache pre-populates cache with frequently accessed users
func (r *CachedUserRepository) WarmupCache(ctx context.Context, userIDs []int64) error {
	for _, id := range userIDs {
		user, err := r.repo.GetByID(ctx, id)
		if err != nil {
			fmt.Printf("Warning: failed to warmup cache for user %d: %v\n", id, err)
			continue
		}

		cacheKeyID := fmt.Sprintf("user:id:%d", id)
		cacheKeyEmail := fmt.Sprintf("user:email:%s", user.Email)

		// Cache by ID
		if err := r.cache.Set(ctx, cacheKeyID, user, UserCacheTTL); err != nil {
			fmt.Printf("Warning: failed to warmup cache (ID) for user %d: %v\n", id, err)
		}

		// Cache by email
		if err := r.cache.Set(ctx, cacheKeyEmail, user, EmailLookupTTL); err != nil {
			fmt.Printf("Warning: failed to warmup cache (email) for user %d: %v\n", id, err)
		}
	}

	return nil
}
