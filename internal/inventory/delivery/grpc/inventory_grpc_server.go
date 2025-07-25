// internal/inventory/delivery/grpc/inventory_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/inventory/application"
	inventory_client "github.com/datngth03/ecommerce-go-app/pkg/client/inventory" // Generated Inventory gRPC client
)

// InventoryGRPCServer implements the inventory_client.InventoryServiceServer interface.
type InventoryGRPCServer struct {
	inventory_client.UnimplementedInventoryServiceServer // Embedded to satisfy all methods
	inventoryService                                   application.InventoryService
}

// NewInventoryGRPCServer creates a new instance of InventoryGRPCServer.
func NewInventoryGRPCServer(svc application.InventoryService) *InventoryGRPCServer {
	return &InventoryGRPCServer{
		inventoryService: svc,
	}
}

// GetStockQuantity implements the gRPC GetStockQuantity method.
func (s *InventoryGRPCServer) GetStockQuantity(ctx context.Context, req *inventory_client.GetStockQuantityRequest) (*inventory_client.StockQuantityResponse, error) {
	log.Printf("Received GetStockQuantity request for Product ID: %s", req.GetProductId())
	resp, err := s.inventoryService.GetStockQuantity(ctx, req)
	if err != nil {
		log.Printf("Error getting stock quantity: %v", err)
		return nil, err
	}
	return resp, nil
}

// IncreaseStock implements the gRPC IncreaseStock method.
func (s *InventoryGRPCServer) IncreaseStock(ctx context.Context, req *inventory_client.IncreaseStockRequest) (*inventory_client.StockQuantityResponse, error) {
	log.Printf("Received IncreaseStock request for Product ID: %s, Quantity: %d", req.GetProductId(), req.GetQuantity())
	resp, err := s.inventoryService.IncreaseStock(ctx, req)
	if err != nil {
		log.Printf("Error increasing stock: %v", err)
		return nil, err
	}
	return resp, nil
}

// DecreaseStock implements the gRPC DecreaseStock method.
func (s *InventoryGRPCServer) DecreaseStock(ctx context.Context, req *inventory_client.DecreaseStockRequest) (*inventory_client.StockQuantityResponse, error) {
	log.Printf("Received DecreaseStock request for Product ID: %s, Quantity: %d", req.GetProductId(), req.GetQuantity())
	resp, err := s.inventoryService.DecreaseStock(ctx, req)
	if err != nil {
		log.Printf("Error decreasing stock: %v", err)
		return nil, err
	}
	return resp, nil
}

// ReserveStock implements the gRPC ReserveStock method.
func (s *InventoryGRPCServer) ReserveStock(ctx context.Context, req *inventory_client.ReserveStockRequest) (*inventory_client.StockQuantityResponse, error) {
	log.Printf("Received ReserveStock request for Product ID: %s, Quantity: %d", req.GetProductId(), req.GetQuantity())
	resp, err := s.inventoryService.ReserveStock(ctx, req)
	if err != nil {
		log.Printf("Error reserving stock: %v", err)
		return nil, err
	}
	return resp, nil
}

// ReleaseStock implements the gRPC ReleaseStock method.
func (s *InventoryGRPCServer) ReleaseStock(ctx context.Context, req *inventory_client.ReleaseStockRequest) (*inventory_client.StockQuantityResponse, error) {
	log.Printf("Received ReleaseStock request for Product ID: %s, Quantity: %d", req.GetProductId(), req.GetQuantity())
	resp, err := s.inventoryService.ReleaseStock(ctx, req)
	if err != nil {
		log.Printf("Error releasing stock: %v", err)
		return nil, err
	}
	return resp, nil
}

// SetStock implements the gRPC SetStock method.
func (s *InventoryGRPCServer) SetStock(ctx context.Context, req *inventory_client.SetStockRequest) (*inventory_client.StockQuantityResponse, error) {
	log.Printf("Received SetStock request for Product ID: %s, Quantity: %d", req.GetProductId(), req.GetQuantity())
	resp, err := s.inventoryService.SetStock(ctx, req)
	if err != nil {
		log.Printf("Error setting stock: %v", err)
		return nil, err
	}
	return resp, nil
}

