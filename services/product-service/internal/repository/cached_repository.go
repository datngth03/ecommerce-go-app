package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/cache"
)

// CachedProductRepository wraps ProductRepository with Redis caching
type CachedProductRepository struct {
	repo  ProductRepository
	cache *cache.RedisCache
}

// Cache TTL constants for products
const (
	ProductCacheTTL      = 5 * time.Minute  // Individual product cache
	ProductListCacheTTL  = 3 * time.Minute  // Product list cache
	CategoryCacheTTL     = 10 * time.Minute // Categories change less frequently
	SearchResultCacheTTL = 2 * time.Minute  // Search results cache
)

// NewCachedProductRepository creates a cached product repository
func NewCachedProductRepository(repo ProductRepository, cache *cache.RedisCache) *CachedProductRepository {
	return &CachedProductRepository{
		repo:  repo,
		cache: cache,
	}
}

// Create creates a new product and invalidates related caches
func (r *CachedProductRepository) Create(ctx context.Context, product *models.Product) error {
	if err := r.repo.Create(ctx, product); err != nil {
		return err
	}

	// Invalidate list caches
	r.invalidateProductCaches(ctx, product.CategoryID)

	return nil
}

// GetByID retrieves a product by ID with caching
func (r *CachedProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:id:%s", id)

	var product models.Product

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		return &product, nil
	}

	if !cache.IsCacheMiss(err) {
		// Log cache error but continue to DB
		fmt.Printf("Cache error for product ID %s: %v\n", id, err)
	}

	// Cache miss or error - fetch from DB
	dbProduct, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, dbProduct, ProductCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache product ID %s: %v\n", id, err)
	}

	return dbProduct, nil
}

// GetBySlug retrieves a product by slug with caching
func (r *CachedProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:slug:%s", slug)

	var product models.Product

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		return &product, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for product slug %s: %v\n", slug, err)
	}

	// Fetch from DB
	dbProduct, err := r.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, dbProduct, ProductCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache product slug %s: %v\n", slug, err)
	}

	return dbProduct, nil
}

// Update updates a product and invalidates its caches
func (r *CachedProductRepository) Update(ctx context.Context, product *models.Product) error {
	if err := r.repo.Update(ctx, product); err != nil {
		return err
	}

	// Invalidate product caches
	r.cache.Delete(ctx,
		fmt.Sprintf("product:id:%s", product.ID),
		fmt.Sprintf("product:slug:%s", product.Slug),
	)

	// Invalidate list caches
	r.invalidateProductCaches(ctx, product.CategoryID)

	return nil
}

// Delete deletes a product and invalidates its caches
func (r *CachedProductRepository) Delete(ctx context.Context, id string) error {
	// Get product first to know category for cache invalidation
	product, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate product caches
	r.cache.Delete(ctx,
		fmt.Sprintf("product:id:%s", id),
		fmt.Sprintf("product:slug:%s", product.Slug),
	)

	// Invalidate list caches
	r.invalidateProductCaches(ctx, product.CategoryID)

	return nil
}

// List retrieves products with caching
func (r *CachedProductRepository) List(ctx context.Context, req *models.ListProductsRequest) ([]models.Product, int64, error) {
	cacheKey := fmt.Sprintf("products:list:page:%d:pagesize:%d:category:%s",
		req.Page, req.PageSize, req.CategoryID)

	var cachedResult struct {
		Products []models.Product
		Total    int64
	}

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Products, cachedResult.Total, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for product list: %v\n", err)
	}

	// Fetch from DB
	products, total, err := r.repo.List(ctx, req)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result
	cachedResult = struct {
		Products []models.Product
		Total    int64
	}{
		Products: products,
		Total:    total,
	}

	if err := r.cache.Set(ctx, cacheKey, cachedResult, ProductListCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache product list: %v\n", err)
	}

	return products, total, nil
}

// ListByCategoryID retrieves products by category with caching
func (r *CachedProductRepository) ListByCategoryID(ctx context.Context, categoryID string, req *models.ListProductsRequest) ([]models.Product, int64, error) {
	cacheKey := fmt.Sprintf("products:category:%s:page:%d:pagesize:%d",
		categoryID, req.Page, req.PageSize)

	var cachedResult struct {
		Products []models.Product
		Total    int64
	}

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Products, cachedResult.Total, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for category products: %v\n", err)
	}

	// Fetch from DB
	products, total, err := r.repo.ListByCategoryID(ctx, categoryID, req)
	if err != nil {
		return nil, 0, err
	}

	// Cache the result
	cachedResult = struct {
		Products []models.Product
		Total    int64
	}{
		Products: products,
		Total:    total,
	}

	if err := r.cache.Set(ctx, cacheKey, cachedResult, ProductListCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache category products: %v\n", err)
	}

	return products, total, nil
}

// ExistsByName checks if product exists by name (no caching for existence checks)
func (r *CachedProductRepository) ExistsByName(ctx context.Context, name string, excludeID ...string) (bool, error) {
	return r.repo.ExistsByName(ctx, name, excludeID...)
}

// CountByCategory counts products by category (cached)
func (r *CachedProductRepository) CountByCategory(ctx context.Context, categoryID string) (int64, error) {
	cacheKey := fmt.Sprintf("products:category:%s:count", categoryID)

	var count int64

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &count)
	if err == nil {
		return count, nil
	}

	// Fetch from DB
	count, err = r.repo.CountByCategory(ctx, categoryID)
	if err != nil {
		return 0, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, count, ProductListCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache category count: %v\n", err)
	}

	return count, nil
}

// invalidateProductCaches invalidates all product list caches
func (r *CachedProductRepository) invalidateProductCaches(ctx context.Context, categoryID string) {
	// Delete product list patterns
	patterns := []string{
		"products:list:*",
		fmt.Sprintf("products:category:%s:*", categoryID),
	}

	for _, pattern := range patterns {
		if err := r.cache.DeletePattern(ctx, pattern); err != nil {
			fmt.Printf("Warning: failed to invalidate cache pattern %s: %v\n", pattern, err)
		}
	}
}

// CachedCategoryRepository wraps CategoryRepository with Redis caching
type CachedCategoryRepository struct {
	repo  CategoryRepository
	cache *cache.RedisCache
}

// NewCachedCategoryRepository creates a cached category repository
func NewCachedCategoryRepository(repo CategoryRepository, cache *cache.RedisCache) *CachedCategoryRepository {
	return &CachedCategoryRepository{
		repo:  repo,
		cache: cache,
	}
}

// Create creates a new category and invalidates caches
func (r *CachedCategoryRepository) Create(ctx context.Context, category *models.Category) error {
	if err := r.repo.Create(ctx, category); err != nil {
		return err
	}

	// Invalidate category list cache
	r.cache.DeletePattern(ctx, "categories:*")

	return nil
}

// GetByID retrieves a category by ID with caching
func (r *CachedCategoryRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	cacheKey := fmt.Sprintf("category:id:%s", id)

	var category models.Category

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &category)
	if err == nil {
		return &category, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for category ID %s: %v\n", id, err)
	}

	// Fetch from DB
	dbCategory, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, dbCategory, CategoryCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache category ID %s: %v\n", id, err)
	}

	return dbCategory, nil
}

// GetBySlug retrieves a category by slug with caching
func (r *CachedCategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	cacheKey := fmt.Sprintf("category:slug:%s", slug)

	var category models.Category

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &category)
	if err == nil {
		return &category, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for category slug %s: %v\n", slug, err)
	}

	// Fetch from DB
	dbCategory, err := r.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, dbCategory, CategoryCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache category slug %s: %v\n", slug, err)
	}

	return dbCategory, nil
}

// Update updates a category and invalidates its caches
func (r *CachedCategoryRepository) Update(ctx context.Context, category *models.Category) error {
	if err := r.repo.Update(ctx, category); err != nil {
		return err
	}

	// Invalidate category caches
	r.cache.Delete(ctx,
		fmt.Sprintf("category:id:%s", category.ID),
		fmt.Sprintf("category:slug:%s", category.Slug),
	)
	r.cache.DeletePattern(ctx, "categories:*")

	return nil
}

// Delete deletes a category and invalidates its caches
func (r *CachedCategoryRepository) Delete(ctx context.Context, id string) error {
	// Get category first for slug
	category, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate category caches
	r.cache.Delete(ctx,
		fmt.Sprintf("category:id:%s", id),
		fmt.Sprintf("category:slug:%s", category.Slug),
	)
	r.cache.DeletePattern(ctx, "categories:*")

	return nil
}

// List retrieves all categories with caching
func (r *CachedCategoryRepository) List(ctx context.Context) ([]models.Category, error) {
	cacheKey := "categories:all"

	var categories []models.Category

	// Try cache first
	err := r.cache.Get(ctx, cacheKey, &categories)
	if err == nil {
		return categories, nil
	}

	if !cache.IsCacheMiss(err) {
		fmt.Printf("Cache error for categories list: %v\n", err)
	}

	// Fetch from DB
	categories, err = r.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := r.cache.Set(ctx, cacheKey, categories, CategoryCacheTTL); err != nil {
		fmt.Printf("Warning: failed to cache categories list: %v\n", err)
	}

	return categories, nil
}

// ExistsByName checks if category exists by name (no caching)
func (r *CachedCategoryRepository) ExistsByName(ctx context.Context, name string, excludeID ...string) (bool, error) {
	return r.repo.ExistsByName(ctx, name, excludeID...)
}

// ExistsByID checks if category exists by ID (no caching)
func (r *CachedCategoryRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	return r.repo.ExistsByID(ctx, id)
}
