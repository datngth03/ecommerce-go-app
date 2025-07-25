// internal/inventory/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/inventory/domain"
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Generated Inventory gRPC client
)

// InventoryService defines the application service interface for inventory management.
// InventoryService định nghĩa interface dịch vụ ứng dụng cho quản lý tồn kho.
type InventoryService interface {
	GetStockQuantity(ctx context.Context, req *inventory_client.GetStockQuantityRequest) (*inventory_client.StockQuantityResponse, error)
	IncreaseStock(ctx context.Context, req *inventory_client.IncreaseStockRequest) (*inventory_client.StockQuantityResponse, error)
	DecreaseStock(ctx context.Context, req *inventory_client.DecreaseStockRequest) (*inventory_client.StockQuantityResponse, error)
	ReserveStock(ctx context.Context, req *inventory_client.ReserveStockRequest) (*inventory_client.StockQuantityResponse, error)
	ReleaseStock(ctx context.Context, req *inventory_client.ReleaseStockRequest) (*inventory_client.StockQuantityResponse, error)
	SetStock(ctx context.Context, req *inventory_client.SetStockRequest) (*inventory_client.StockQuantityResponse, error)
}

// inventoryService implements the InventoryService interface.
// inventoryService triển khai interface InventoryService.
type inventoryService struct {
	inventoryRepo domain.InventoryRepository
	// TODO: Add other dependencies like event publisher (e.g., for Product Service to update its stock_quantity)
}

// NewInventoryService creates a new instance of InventoryService.
// NewInventoryService tạo một thể hiện mới của InventoryService.
func NewInventoryService(repo domain.InventoryRepository) InventoryService {
	return &inventoryService{
		inventoryRepo: repo,
	}
}

// GetStockQuantity retrieves the current stock and reserved quantity for a product.
// GetStockQuantity lấy số lượng tồn kho hiện tại và số lượng đã đặt trước của một sản phẩm.
func (s *inventoryService) GetStockQuantity(ctx context.Context, req *inventory_client.GetStockQuantityRequest) (*inventory_client.StockQuantityResponse, error) {
	if req.GetProductId() == "" {
		return nil, errors.New("product ID is required")
	}

	item, err := s.inventoryRepo.FindByProductID(ctx, req.GetProductId())
	if err != nil {
		if errors.Is(err, errors.New("inventory item not found")) { // Assuming specific error from repo
			// If item not found, return 0 stock
			return &inventory_client.StockQuantityResponse{
				Item: &inventory_client.InventoryItem{
					ProductId:        req.GetProductId(),
					Quantity:         0,
					ReservedQuantity: 0,
					LastUpdatedAt:    time.Now().Format(time.RFC3339),
				},
				Message: "Product inventory not found, assuming 0 stock",
			}, nil
		}
		return nil, fmt.Errorf("failed to retrieve inventory item: %w", err)
	}

	return &inventory_client.StockQuantityResponse{
		Item: &inventory_client.InventoryItem{
			ProductId:        item.ProductID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			LastUpdatedAt:    item.LastUpdatedAt.Format(time.RFC3339),
		},
		Message: "Stock quantity retrieved successfully",
	}, nil
}

// IncreaseStock increases the available stock quantity for a product.
// IncreaseStock tăng số lượng tồn kho có sẵn cho một sản phẩm.
func (s *inventoryService) IncreaseStock(ctx context.Context, req *inventory_client.IncreaseStockRequest) (*inventory_client.StockQuantityResponse, error) {
	if req.GetProductId() == "" || req.GetQuantity() <= 0 {
		return nil, errors.New("product ID and positive quantity are required")
	}

	item, err := s.inventoryRepo.FindByProductID(ctx, req.GetProductId())
	if err != nil {
		if errors.Is(err, errors.New("inventory item not found")) {
			// If not found, create a new one
			item = domain.NewInventoryItem(req.GetProductId(), req.GetQuantity())
		} else {
			return nil, fmt.Errorf("failed to retrieve inventory item: %w", err)
		}
	} else {
		item.IncreaseQuantity(req.GetQuantity())
	}

	if err := s.inventoryRepo.Save(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to save inventory item: %w", err)
	}

	// TODO: Publish StockUpdated event (e.g., for Product Service to update its denormalized stock)

	return &inventory_client.StockQuantityResponse{
		Item: &inventory_client.InventoryItem{
			ProductId:        item.ProductID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			LastUpdatedAt:    item.LastUpdatedAt.Format(time.RFC3339),
		},
		Message: "Stock increased successfully",
	}, nil
}

// DecreaseStock decreases the available stock quantity for a product.
// DecreaseStock giảm số lượng tồn kho có sẵn cho một sản phẩm.
func (s *inventoryService) DecreaseStock(ctx context.Context, req *inventory_client.DecreaseStockRequest) (*inventory_client.StockQuantityResponse, error) {
	if req.GetProductId() == "" || req.GetQuantity() <= 0 {
		return nil, errors.New("product ID and positive quantity are required")
	}

	item, err := s.inventoryRepo.FindByProductID(ctx, req.GetProductId())
	if err != nil {
		if errors.Is(err, errors.New("inventory item not found")) {
			return nil, errors.New("inventory item not found for decrease operation")
		}
		return nil, fmt.Errorf("failed to retrieve inventory item: %w", err)
	}

	if err := item.DecreaseQuantity(req.GetQuantity()); err != nil {
		return nil, fmt.Errorf("failed to decrease stock: %w", err)
	}

	if err := s.inventoryRepo.Save(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to save inventory item: %w", err)
	}

	// TODO: Publish StockUpdated event

	return &inventory_client.StockQuantityResponse{
		Item: &inventory_client.InventoryItem{
			ProductId:        item.ProductID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			LastUpdatedAt:    item.LastUpdatedAt.Format(time.RFC3339),
		},
		Message: "Stock decreased successfully",
	}, nil
}

// ReserveStock reserves a certain amount of stock for a product.
// ReserveStock đặt trước một số lượng tồn kho nhất định cho một sản phẩm.
func (s *inventoryService) ReserveStock(ctx context.Context, req *inventory_client.ReserveStockRequest) (*inventory_client.StockQuantityResponse, error) {
	if req.GetProductId() == "" || req.GetQuantity() <= 0 {
		return nil, errors.New("product ID and positive quantity are required")
	}

	item, err := s.inventoryRepo.FindByProductID(ctx, req.GetProductId())
	if err != nil {
		if errors.Is(err, errors.New("inventory item not found")) {
			return nil, errors.New("inventory item not found for reservation")
		}
		return nil, fmt.Errorf("failed to retrieve inventory item: %w", err)
	}

	if err := item.ReserveQuantity(req.GetQuantity()); err != nil {
		return nil, fmt.Errorf("failed to reserve stock: %w", err)
	}

	if err := s.inventoryRepo.Save(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to save inventory item: %w", err)
	}

	// TODO: Publish StockReserved event

	return &inventory_client.StockQuantityResponse{
		Item: &inventory_client.InventoryItem{
			ProductId:        item.ProductID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			LastUpdatedAt:    item.LastUpdatedAt.Format(time.RFC3339),
		},
		Message: "Stock reserved successfully",
	}, nil
}

// ReleaseStock releases a certain amount of reserved stock back to available.
// ReleaseStock giải phóng một số lượng tồn kho đã đặt trước trở lại trạng thái có sẵn.
func (s *inventoryService) ReleaseStock(ctx context.Context, req *inventory_client.ReleaseStockRequest) (*inventory_client.StockQuantityResponse, error) {
	if req.GetProductId() == "" || req.GetQuantity() <= 0 {
		return nil, errors.New("product ID and positive quantity are required")
	}

	item, err := s.inventoryRepo.FindByProductID(ctx, req.GetProductId())
	if err != nil {
		if errors.Is(err, errors.New("inventory item not found")) {
			return nil, errors.New("inventory item not found for release operation")
		}
		return nil, fmt.Errorf("failed to retrieve inventory item: %w", err)
	}

	if err := item.ReleaseReservedQuantity(req.GetQuantity()); err != nil {
		return nil, fmt.Errorf("failed to release reserved stock: %w", err)
	}

	if err := s.inventoryRepo.Save(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to save inventory item: %w", err)
	}

	// TODO: Publish StockReleased event

	return &inventory_client.StockQuantityResponse{
		Item: &inventory_client.InventoryItem{
			ProductId:        item.ProductID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			LastUpdatedAt:    item.LastUpdatedAt.Format(time.RFC3339),
		},
		Message: "Reserved stock released successfully",
	}, nil
}

// SetStock sets the total available stock quantity directly for a product.
// SetStock đặt trực tiếp tổng số lượng tồn kho có sẵn cho một sản phẩm.
func (s *inventoryService) SetStock(ctx context.Context, req *inventory_client.SetStockRequest) (*inventory_client.StockQuantityResponse, error) {
	if req.GetProductId() == "" || req.GetQuantity() < 0 {
		return nil, errors.New("product ID and non-negative quantity are required")
	}

	item, err := s.inventoryRepo.FindByProductID(ctx, req.GetProductId())
	if err != nil {
		if errors.Is(err, errors.New("inventory item not found")) {
			// If not found, create a new one
			item = domain.NewInventoryItem(req.GetProductId(), req.GetQuantity())
		} else {
			return nil, fmt.Errorf("failed to retrieve inventory item: %w", err)
		}
	} else {
		item.SetQuantity(req.GetQuantity())
	}

	if err := s.inventoryRepo.Save(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to save inventory item: %w", err)
	}

	// TODO: Publish StockUpdated event

	return &inventory_client.StockQuantityResponse{
		Item: &inventory_client.InventoryItem{
			ProductId:        item.ProductID,
			Quantity:         item.Quantity,
			ReservedQuantity: item.ReservedQuantity,
			LastUpdatedAt:    item.LastUpdatedAt.Format(time.RFC3339),
		},
		Message: "Stock set successfully",
	}, nil
}
