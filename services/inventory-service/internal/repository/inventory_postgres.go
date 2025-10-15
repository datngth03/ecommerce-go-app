package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/models"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inventoryRepository struct {
	db          *gorm.DB
	redisClient *redis.Client
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *gorm.DB, redisClient *redis.Client) InventoryRepository {
	return &inventoryRepository{
		db:          db,
		redisClient: redisClient,
	}
}

// GetStock retrieves stock for a product (with Redis caching)
func (r *inventoryRepository) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("stock:%s", productID)
	cached, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var stock models.Stock
		if err := json.Unmarshal([]byte(cached), &stock); err == nil {
			return &stock, nil
		}
	}

	// Not in cache, fetch from DB
	var stock models.Stock
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Initialize stock for new product
			stock = models.Stock{
				ProductID:   productID,
				Available:   0,
				Reserved:    0,
				Total:       0,
				WarehouseID: "default",
			}
			if err := r.db.WithContext(ctx).Create(&stock).Error; err != nil {
				return nil, fmt.Errorf("failed to create stock: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get stock: %w", err)
		}
	}

	// Cache for 5 minutes
	if data, err := json.Marshal(stock); err == nil {
		r.redisClient.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return &stock, nil
}

// UpdateStock updates stock quantity (with transaction)
func (r *inventoryRepository) UpdateStock(ctx context.Context, productID string, quantity int32, reason string) (*models.Stock, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lock row for update
	var stock models.Stock
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ?", productID).
		First(&stock).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			// Create new stock
			stock = models.Stock{
				ProductID:   productID,
				Available:   0,
				Reserved:    0,
				Total:       0,
				WarehouseID: "default",
			}
			if err := tx.Create(&stock).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create stock: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to lock stock: %w", err)
		}
	}

	// Calculate new quantities
	beforeTotal := stock.Total
	stock.Total += quantity
	stock.Available = stock.Total - stock.Reserved

	// Validate
	if stock.Total < 0 {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient stock: cannot have negative total")
	}
	if stock.Available < 0 {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient stock: available would be negative")
	}

	// Update stock
	if err := tx.Save(&stock).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// Create movement record
	movementType := models.MovementTypeInbound
	if quantity < 0 {
		movementType = models.MovementTypeOutbound
	}

	movement := &models.StockMovement{
		ProductID:      productID,
		MovementType:   movementType,
		Quantity:       quantity,
		BeforeQuantity: beforeTotal,
		AfterQuantity:  stock.Total,
		ReferenceType:  models.ReferenceTypeAdjustment,
		Reason:         reason,
	}

	if err := tx.Create(movement).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create movement: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("stock:%s", productID)
	r.redisClient.Del(ctx, cacheKey)

	return &stock, nil
}

// CheckAvailability checks if product has enough stock
func (r *inventoryRepository) CheckAvailability(ctx context.Context, productID string, quantity int32) (bool, error) {
	stock, err := r.GetStock(ctx, productID)
	if err != nil {
		return false, err
	}

	return stock.Available >= quantity, nil
}

// CreateReservation reserves stock for an order
func (r *inventoryRepository) CreateReservation(ctx context.Context, orderID, productID string, quantity int32) (*models.Reservation, error) {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Lock stock row
	var stock models.Stock
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ?", productID).
		First(&stock).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to lock stock: %w", err)
	}

	// Check availability
	if stock.Available < quantity {
		tx.Rollback()
		return nil, fmt.Errorf("insufficient stock: need %d, have %d", quantity, stock.Available)
	}

	// Update stock
	beforeAvailable := stock.Available
	stock.Reserved += quantity
	stock.Available -= quantity

	if err := tx.Save(&stock).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// Create reservation
	reservation := &models.Reservation{
		OrderID:     orderID,
		ProductID:   productID,
		Quantity:    quantity,
		Status:      models.ReservationStatusPending,
		WarehouseID: stock.WarehouseID,
		ExpiresAt:   time.Now().Add(30 * time.Minute), // 30 min to complete payment
	}

	if err := tx.Create(reservation).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Create movement record
	movement := &models.StockMovement{
		ProductID:      productID,
		MovementType:   models.MovementTypeReserved,
		Quantity:       quantity,
		BeforeQuantity: beforeAvailable,
		AfterQuantity:  stock.Available,
		ReferenceType:  models.ReferenceTypeOrder,
		ReferenceID:    orderID,
		Reason:         "Stock reserved for order",
	}

	if err := tx.Create(movement).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create movement: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("stock:%s", productID)
	r.redisClient.Del(ctx, cacheKey)

	return reservation, nil
}

// GetReservation retrieves reservations for an order
func (r *inventoryRepository) GetReservation(ctx context.Context, orderID string) ([]*models.Reservation, error) {
	var reservations []*models.Reservation
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&reservations).Error; err != nil {
		return nil, fmt.Errorf("failed to get reservations: %w", err)
	}
	return reservations, nil
}

// CommitReservation commits reserved stock (payment completed)
func (r *inventoryRepository) CommitReservation(ctx context.Context, orderID string) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get reservations
	var reservations []*models.Reservation
	if err := tx.Where("order_id = ? AND status = ?", orderID, models.ReservationStatusPending).
		Find(&reservations).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get reservations: %w", err)
	}

	if len(reservations) == 0 {
		tx.Rollback()
		return fmt.Errorf("no pending reservations found for order %s", orderID)
	}

	// Process each reservation
	for _, res := range reservations {
		// Lock stock
		var stock models.Stock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ?", res.ProductID).
			First(&stock).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to lock stock: %w", err)
		}

		// Update stock (reduce reserved and total)
		beforeTotal := stock.Total
		stock.Reserved -= res.Quantity
		stock.Total -= res.Quantity

		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update stock: %w", err)
		}

		// Update reservation status
		res.Status = models.ReservationStatusCommitted
		if err := tx.Save(res).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update reservation: %w", err)
		}

		// Create movement record
		movement := &models.StockMovement{
			ProductID:      res.ProductID,
			MovementType:   models.MovementTypeCommitted,
			Quantity:       -res.Quantity,
			BeforeQuantity: beforeTotal,
			AfterQuantity:  stock.Total,
			ReferenceType:  models.ReferenceTypeOrder,
			ReferenceID:    orderID,
			Reason:         "Stock committed (order completed)",
		}

		if err := tx.Create(movement).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create movement: %w", err)
		}

		// Invalidate cache
		cacheKey := fmt.Sprintf("stock:%s", res.ProductID)
		r.redisClient.Del(ctx, cacheKey)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReleaseReservation releases reserved stock (order cancelled)
func (r *inventoryRepository) ReleaseReservation(ctx context.Context, orderID string, reason string) error {
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get reservations
	var reservations []*models.Reservation
	if err := tx.Where("order_id = ? AND status = ?", orderID, models.ReservationStatusPending).
		Find(&reservations).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get reservations: %w", err)
	}

	if len(reservations) == 0 {
		tx.Rollback()
		return fmt.Errorf("no pending reservations found for order %s", orderID)
	}

	// Process each reservation
	for _, res := range reservations {
		// Lock stock
		var stock models.Stock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("product_id = ?", res.ProductID).
			First(&stock).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to lock stock: %w", err)
		}

		// Update stock (reduce reserved, increase available)
		beforeAvailable := stock.Available
		stock.Reserved -= res.Quantity
		stock.Available += res.Quantity

		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update stock: %w", err)
		}

		// Update reservation status
		res.Status = models.ReservationStatusReleased
		if err := tx.Save(res).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update reservation: %w", err)
		}

		// Create movement record
		movement := &models.StockMovement{
			ProductID:      res.ProductID,
			MovementType:   models.MovementTypeReleased,
			Quantity:       res.Quantity,
			BeforeQuantity: beforeAvailable,
			AfterQuantity:  stock.Available,
			ReferenceType:  models.ReferenceTypeOrder,
			ReferenceID:    orderID,
			Reason:         reason,
		}

		if err := tx.Create(movement).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create movement: %w", err)
		}

		// Invalidate cache
		cacheKey := fmt.Sprintf("stock:%s", res.ProductID)
		r.redisClient.Del(ctx, cacheKey)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateMovement creates a stock movement record
func (r *inventoryRepository) CreateMovement(ctx context.Context, movement *models.StockMovement) error {
	if err := r.db.WithContext(ctx).Create(movement).Error; err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}
	return nil
}

// GetMovementHistory retrieves stock movement history
func (r *inventoryRepository) GetMovementHistory(ctx context.Context, productID string, limit, offset int) ([]*models.StockMovement, int, error) {
	var movements []*models.StockMovement
	var total int64

	query := r.db.WithContext(ctx).Where("product_id = ?", productID)

	// Get total count
	if err := query.Model(&models.StockMovement{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count movements: %w", err)
	}

	// Get paginated results
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&movements).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get movements: %w", err)
	}

	return movements, int(total), nil
}
