package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/cache"
)

// CachedInventoryRepository wraps InventoryRepository with Redis caching
type CachedInventoryRepository struct {
	repo  InventoryRepository
	cache *cache.RedisCache
}

// Cache TTL constants for inventory
const (
	StockCacheTTL       = 1 * time.Minute  // Stock levels need near real-time accuracy
	AvailabilityTTL     = 30 * time.Second // Availability checks even shorter
	ReservationCacheTTL = 2 * time.Minute  // Reservations cached briefly
	MovementHistoryTTL  = 5 * time.Minute  // History changes less frequently
)

// NewCachedInventoryRepository creates a cached inventory repository
func NewCachedInventoryRepository(repo InventoryRepository, cache *cache.RedisCache) *CachedInventoryRepository {
	return &CachedInventoryRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetStock retrieves stock information with short TTL caching
func (r *CachedInventoryRepository) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	cacheKey := fmt.Sprintf("stock:product:%s", productID)

	var stock models.Stock

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &stock)
	if err == nil {
		return &stock, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for stock product %s: %v\n", productID, err)
	}

	// Cache miss - fetch from DB
	dbStock, err := r.repo.GetStock(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Cache with short TTL (1 minute) for real-time accuracy
	if err := r.cache.Set(ctx, cacheKey, dbStock, StockCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache stock for product %s: %v\n", productID, err)
	}

	return dbStock, nil
}

// UpdateStock updates stock and invalidates cache immediately
func (r *CachedInventoryRepository) UpdateStock(ctx context.Context, productID string, quantity int32, reason string) (*models.Stock, error) {
	// Update in database
	updatedStock, err := r.repo.UpdateStock(ctx, productID, quantity, reason)
	if err != nil {
		return nil, err
	}

	// Invalidate stock cache immediately after update
	cacheKeys := []string{
		fmt.Sprintf("stock:product:%s", productID),
		fmt.Sprintf("availability:product:%s:*", productID), // Pattern for all availability checks
	}

	for _, key := range cacheKeys {
		if err := r.cache.DeletePattern(ctx, key); err != nil {
			fmt.Printf("Warning: failed to invalidate cache for %s: %v\n", key, err)
		}
	}

	// Cache the new stock value
	cacheKey := fmt.Sprintf("stock:product:%s", productID)
	if err := r.cache.Set(ctx, cacheKey, updatedStock, StockCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache updated stock: %v\n", err)
	}

	return updatedStock, nil
}

// CheckAvailability checks stock availability with very short TTL
func (r *CachedInventoryRepository) CheckAvailability(ctx context.Context, productID string, quantity int32) (bool, error) {
	cacheKey := fmt.Sprintf("availability:product:%s:qty:%d", productID, quantity)

	var available bool

	// Try cache first (very short TTL)
	err := r.cache.Get(ctx, cacheKey, &available)
	if err == nil {
		return available, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for availability check: %v\n", err)
	}

	// Check from DB
	available, err = r.repo.CheckAvailability(ctx, productID, quantity)
	if err != nil {
		return false, err
	}

	// Cache with very short TTL (30 seconds) for accuracy
	if err := r.cache.Set(ctx, cacheKey, available, AvailabilityTTL); err != nil {
		fmt.Printf("Warning: failed to cache availability: %v\n", err)
	}

	return available, nil
}

// CreateReservation creates a reservation and invalidates related caches
func (r *CachedInventoryRepository) CreateReservation(ctx context.Context, orderID, productID string, quantity int32) (*models.Reservation, error) {
	// Create in database
	reservation, err := r.repo.CreateReservation(ctx, orderID, productID, quantity)
	if err != nil {
		return nil, err
	}

	// Invalidate stock and availability caches (reservation affects availability)
	keysToInvalidate := []string{
		fmt.Sprintf("stock:product:%s", productID),
		fmt.Sprintf("reservation:order:%s", orderID),
	}

	if err := r.cache.Delete(ctx, keysToInvalidate...); err != nil {
		fmt.Printf("Warning: failed to invalidate caches after reservation: %v\n", err)
	}

	// Invalidate all availability checks for this product
	if err := r.cache.DeletePattern(ctx, fmt.Sprintf("availability:product:%s:*", productID)); err != nil {
		fmt.Printf("Warning: failed to invalidate availability pattern: %v\n", err)
	}

	return reservation, nil
}

// GetReservation retrieves reservations with caching
func (r *CachedInventoryRepository) GetReservation(ctx context.Context, orderID string) ([]*models.Reservation, error) {
	cacheKey := fmt.Sprintf("reservation:order:%s", orderID)

	var reservations []*models.Reservation

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &reservations)
	if err == nil {
		return reservations, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for reservations: %v\n", err)
	}

	// Fetch from DB
	dbReservations, err := r.repo.GetReservation(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Cache reservations
	if err := r.cache.Set(ctx, cacheKey, dbReservations, ReservationCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache reservations: %v\n", err)
	}

	return dbReservations, nil
}

// CommitReservation commits a reservation and invalidates all related caches
func (r *CachedInventoryRepository) CommitReservation(ctx context.Context, orderID string) error {
	// Get reservations first to know which products to invalidate
	reservations, err := r.repo.GetReservation(ctx, orderID)
	if err != nil {
		return err
	}

	// Commit in database
	if err := r.repo.CommitReservation(ctx, orderID); err != nil {
		return err
	}

	// Invalidate all related caches
	keysToInvalidate := []string{
		fmt.Sprintf("reservation:order:%s", orderID),
	}

	// Invalidate stock cache for all affected products
	for _, res := range reservations {
		keysToInvalidate = append(keysToInvalidate, fmt.Sprintf("stock:product:%s", res.ProductID))

		// Also invalidate availability patterns
		if err := r.cache.DeletePattern(ctx, fmt.Sprintf("availability:product:%s:*", res.ProductID)); err != nil {
			fmt.Printf("Warning: failed to invalidate availability for product %s: %v\n", res.ProductID, err)
		}
	}

	if err := r.cache.Delete(ctx, keysToInvalidate...); err != nil {
		fmt.Printf("Warning: failed to invalidate caches after commit: %v\n", err)
	}

	return nil
}

// ReleaseReservation releases a reservation and invalidates caches
func (r *CachedInventoryRepository) ReleaseReservation(ctx context.Context, orderID string, reason string) error {
	// Get reservations first to know which products to invalidate
	reservations, err := r.repo.GetReservation(ctx, orderID)
	if err != nil {
		return err
	}

	// Release in database
	if err := r.repo.ReleaseReservation(ctx, orderID, reason); err != nil {
		return err
	}

	// Invalidate all related caches
	keysToInvalidate := []string{
		fmt.Sprintf("reservation:order:%s", orderID),
	}

	// Invalidate stock cache for all affected products
	for _, res := range reservations {
		keysToInvalidate = append(keysToInvalidate, fmt.Sprintf("stock:product:%s", res.ProductID))

		// Also invalidate availability patterns
		if err := r.cache.DeletePattern(ctx, fmt.Sprintf("availability:product:%s:*", res.ProductID)); err != nil {
			fmt.Printf("Warning: failed to invalidate availability for product %s: %v\n", res.ProductID, err)
		}
	}

	if err := r.cache.Delete(ctx, keysToInvalidate...); err != nil {
		fmt.Printf("Warning: failed to invalidate caches after release: %v\n", err)
	}

	return nil
}

// CreateMovement creates a stock movement (no caching - write operation)
func (r *CachedInventoryRepository) CreateMovement(ctx context.Context, movement *models.StockMovement) error {
	if err := r.repo.CreateMovement(ctx, movement); err != nil {
		return err
	}

	// Invalidate movement history cache for this product
	cacheKey := fmt.Sprintf("movements:product:%s:*", movement.ProductID)
	if err := r.cache.DeletePattern(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: failed to invalidate movement history: %v\n", err)
	}

	return nil
}

// GetMovementHistory retrieves stock movement history with caching
func (r *CachedInventoryRepository) GetMovementHistory(ctx context.Context, productID string, limit, offset int) ([]*models.StockMovement, int, error) {
	cacheKey := fmt.Sprintf("movements:product:%s:limit:%d:offset:%d", productID, limit, offset)

	var cachedResult struct {
		Movements []*models.StockMovement
		Total     int
	}

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Movements, cachedResult.Total, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for movement history: %v\n", err)
	}

	// Fetch from DB
	movements, total, err := r.repo.GetMovementHistory(ctx, productID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result
	cachedResult = struct {
		Movements []*models.StockMovement
		Total     int
	}{
		Movements: movements,
		Total:     total,
	}

	if err := r.cache.Set(ctx, cacheKey, cachedResult, MovementHistoryTTL); err != nil {
		fmt.Printf("Warning: failed to cache movement history: %v\n", err)
	}

	return movements, total, nil
}

// InvalidateProductCache manually invalidates all caches for a product
func (r *CachedInventoryRepository) InvalidateProductCache(ctx context.Context, productID string) error {
	patterns := []string{
		fmt.Sprintf("stock:product:%s", productID),
		fmt.Sprintf("availability:product:%s:*", productID),
		fmt.Sprintf("movements:product:%s:*", productID),
	}

	for _, pattern := range patterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// WarmupStockCache pre-populates cache with frequently accessed products
func (r *CachedInventoryRepository) WarmupStockCache(ctx context.Context, productIDs []string) error {
	for _, productID := range productIDs {
		stock, err := r.repo.GetStock(ctx, productID)
		if err != nil {
			fmt.Printf("Warning: failed to warmup stock cache for product %s: %v\n", productID, err)
			continue
		}

		cacheKey := fmt.Sprintf("stock:product:%s", productID)
		if err := r.cache.Set(ctx, cacheKey, stock, StockCacheTTL); err != nil {
			fmt.Printf("Warning: failed to cache stock during warmup: %v\n", err)
		}
	}

	return nil
}
