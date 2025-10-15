package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type CartPostgresRepository struct {
	db          *sql.DB
	redisClient *redis.Client
}

func NewCartPostgresRepository(db *sql.DB, redisClient *redis.Client) *CartPostgresRepository {
	return &CartPostgresRepository{
		db:          db,
		redisClient: redisClient,
	}
}

// Get retrieves cart from Redis cache first, fallback to PostgreSQL
func (r *CartPostgresRepository) Get(ctx context.Context, userID int64) (*models.Cart, error) {
	// Try Redis cache first
	cacheKey := fmt.Sprintf("cart:user:%d", userID)
	cachedData, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cart models.Cart
		if err := json.Unmarshal([]byte(cachedData), &cart); err == nil {
			return &cart, nil
		}
	}

	// Fallback to PostgreSQL
	cart := &models.Cart{UserID: userID, Items: []models.CartItem{}}

	query := `
		SELECT id, user_id, created_at, updated_at
		FROM carts WHERE user_id = $1`

	err = r.db.QueryRowContext(ctx, query, userID).Scan(
		&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		// Create new cart
		return r.createCart(ctx, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Get cart items
	itemQuery := `
		SELECT id, cart_id, product_id, product_name, quantity, price, created_at, updated_at
		FROM cart_items WHERE cart_id = $1`

	rows, err := r.db.QueryContext(ctx, itemQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.CartItem
		err = rows.Scan(&item.ID, &item.CartID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.Price, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}
		cart.Items = append(cart.Items, item)
	}

	// Cache in Redis for 1 hour
	r.cacheCart(ctx, cart)

	return cart, nil
}

// AddItem adds or updates item in cart
func (r *CartPostgresRepository) AddItem(ctx context.Context, userID int64, item *models.CartItem) (*models.Cart, error) {
	cart, err := r.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if item already exists
	var existingItem *models.CartItem
	for i := range cart.Items {
		if cart.Items[i].ProductID == item.ProductID {
			existingItem = &cart.Items[i]
			break
		}
	}

	if existingItem != nil {
		// Update quantity
		existingItem.Quantity += item.Quantity
		query := `
			UPDATE cart_items 
			SET quantity = $1, updated_at = NOW()
			WHERE id = $2`
		_, err = r.db.ExecContext(ctx, query, existingItem.Quantity, existingItem.ID)
	} else {
		// Insert new item
		item.ID = uuid.New().String()
		item.CartID = cart.ID
		query := `
			INSERT INTO cart_items (id, cart_id, product_id, product_name, quantity, price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`
		_, err = r.db.ExecContext(ctx, query,
			item.ID, item.CartID, item.ProductID, item.ProductName, item.Quantity, item.Price)
		if err == nil {
			cart.Items = append(cart.Items, *item)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}

	// Update cart timestamp
	r.db.ExecContext(ctx, "UPDATE carts SET updated_at = NOW() WHERE id = $1", cart.ID)

	// Invalidate cache
	r.invalidateCache(ctx, userID)

	return r.Get(ctx, userID)
}

// UpdateItem updates item quantity in cart
func (r *CartPostgresRepository) UpdateItem(ctx context.Context, userID int64, productID string, quantity int32) (*models.Cart, error) {
	cart, err := r.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	var itemID string
	for _, item := range cart.Items {
		if item.ProductID == productID {
			itemID = item.ID
			break
		}
	}

	if itemID == "" {
		return nil, fmt.Errorf("item not found in cart")
	}

	query := `
		UPDATE cart_items 
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2 AND cart_id = $3`

	result, err := r.db.ExecContext(ctx, query, quantity, itemID, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("item not found")
	}

	// Invalidate cache
	r.invalidateCache(ctx, userID)

	return r.Get(ctx, userID)
}

// RemoveItem removes item from cart
func (r *CartPostgresRepository) RemoveItem(ctx context.Context, userID int64, productID string) (*models.Cart, error) {
	cart, err := r.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	query := `
		DELETE FROM cart_items 
		WHERE cart_id = $1 AND product_id = $2`

	_, err = r.db.ExecContext(ctx, query, cart.ID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove item from cart: %w", err)
	}

	// Invalidate cache
	r.invalidateCache(ctx, userID)

	return r.Get(ctx, userID)
}

// Clear removes all items from cart
func (r *CartPostgresRepository) Clear(ctx context.Context, userID int64) error {
	cart, err := r.Get(ctx, userID)
	if err != nil {
		return err
	}

	query := `DELETE FROM cart_items WHERE cart_id = $1`
	_, err = r.db.ExecContext(ctx, query, cart.ID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	// Invalidate cache
	r.invalidateCache(ctx, userID)

	return nil
}

// Helper methods

func (r *CartPostgresRepository) createCart(ctx context.Context, userID int64) (*models.Cart, error) {
	cart := &models.Cart{
		ID:     uuid.New().String(),
		UserID: userID,
		Items:  []models.CartItem{},
	}

	query := `
		INSERT INTO carts (id, user_id, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query, cart.ID, cart.UserID).Scan(
		&cart.CreatedAt, &cart.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}

	return cart, nil
}

func (r *CartPostgresRepository) cacheCart(ctx context.Context, cart *models.Cart) {
	data, err := json.Marshal(cart)
	if err != nil {
		return
	}
	cacheKey := fmt.Sprintf("cart:user:%d", cart.UserID)
	r.redisClient.Set(ctx, cacheKey, data, time.Hour)
}

func (r *CartPostgresRepository) invalidateCache(ctx context.Context, userID int64) {
	cacheKey := fmt.Sprintf("cart:user:%d", userID)
	r.redisClient.Del(ctx, cacheKey)
}

// ConnectRedis creates a Redis client connection
func ConnectRedis(addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return client, nil
}
