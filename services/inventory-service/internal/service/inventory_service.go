package service

import (
	"context"
	"fmt"

	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/repository"
)

// InventoryService handles inventory business logic
type InventoryService struct {
	repo repository.InventoryRepository
}

// NewInventoryService creates a new inventory service
func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{
		repo: repo,
	}
}

// GetStock retrieves current stock for a product
func (s *InventoryService) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	if productID == "" {
		return nil, fmt.Errorf("product_id is required")
	}

	return s.repo.GetStock(ctx, productID)
}

// UpdateStock updates stock quantity
func (s *InventoryService) UpdateStock(ctx context.Context, productID string, quantity int32, reason string) (*models.Stock, error) {
	if productID == "" {
		return nil, fmt.Errorf("product_id is required")
	}

	if quantity == 0 {
		return nil, fmt.Errorf("quantity cannot be zero")
	}

	return s.repo.UpdateStock(ctx, productID, quantity, reason)
}

// ReserveStock reserves stock for an order
func (s *InventoryService) ReserveStock(ctx context.Context, orderID string, items []struct {
	ProductID string
	Quantity  int32
}) (string, error) {
	if orderID == "" {
		return "", fmt.Errorf("order_id is required")
	}

	if len(items) == 0 {
		return "", fmt.Errorf("items are required")
	}

	// Check availability for all items first
	for _, item := range items {
		available, err := s.repo.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return "", fmt.Errorf("failed to check availability for %s: %w", item.ProductID, err)
		}

		if !available {
			stock, _ := s.repo.GetStock(ctx, item.ProductID)
			return "", fmt.Errorf("insufficient stock for product %s: need %d, have %d",
				item.ProductID, item.Quantity, stock.Available)
		}
	}

	// Reserve all items
	for _, item := range items {
		_, err := s.repo.CreateReservation(ctx, orderID, item.ProductID, item.Quantity)
		if err != nil {
			// Rollback: release already reserved items
			s.repo.ReleaseReservation(ctx, orderID, "Reservation failed")
			return "", fmt.Errorf("failed to reserve stock for %s: %w", item.ProductID, err)
		}
	}

	return orderID, nil
}

// ReleaseStock releases reserved stock
func (s *InventoryService) ReleaseStock(ctx context.Context, orderID string, reason string) error {
	if orderID == "" {
		return fmt.Errorf("order_id is required")
	}

	return s.repo.ReleaseReservation(ctx, orderID, reason)
}

// CommitStock commits reserved stock
func (s *InventoryService) CommitStock(ctx context.Context, orderID string) error {
	if orderID == "" {
		return fmt.Errorf("order_id is required")
	}

	return s.repo.CommitReservation(ctx, orderID)
}

// CheckAvailability checks if products are available
func (s *InventoryService) CheckAvailability(ctx context.Context, items []struct {
	ProductID string
	Quantity  int32
}) (bool, []map[string]interface{}, error) {
	unavailable := []map[string]interface{}{}

	for _, item := range items {
		available, err := s.repo.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return false, nil, fmt.Errorf("failed to check availability: %w", err)
		}

		if !available {
			stock, _ := s.repo.GetStock(ctx, item.ProductID)
			unavailable = append(unavailable, map[string]interface{}{
				"product_id": item.ProductID,
				"requested":  item.Quantity,
				"available":  stock.Available,
			})
		}
	}

	return len(unavailable) == 0, unavailable, nil
}

// GetStockHistory retrieves stock movement history
func (s *InventoryService) GetStockHistory(ctx context.Context, productID string, limit, offset int) ([]*models.StockMovement, int, error) {
	if productID == "" {
		return nil, 0, fmt.Errorf("product_id is required")
	}

	if limit <= 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	return s.repo.GetMovementHistory(ctx, productID, limit, offset)
}
