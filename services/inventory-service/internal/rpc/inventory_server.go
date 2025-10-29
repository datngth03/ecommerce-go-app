package rpc

import (
	"context"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/middleware"
	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InventoryServer implements the gRPC inventory service
type InventoryServer struct {
	pb.UnimplementedInventoryServiceServer
	service *service.InventoryService
}

// NewInventoryServer creates a new gRPC inventory server
func NewInventoryServer(svc *service.InventoryService) *InventoryServer {
	return &InventoryServer{
		service: svc,
	}
}

// GetStock retrieves stock information
func (s *InventoryServer) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("GetStock", statusCode, time.Since(start))
	}()

	stock, err := s.service.GetStock(ctx, req.ProductId)
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	return &pb.GetStockResponse{
		Stock: &pb.Stock{
			ProductId:   stock.ProductID,
			Available:   stock.Available,
			Reserved:    stock.Reserved,
			Total:       stock.Total,
			WarehouseId: stock.WarehouseID,
		},
	}, nil
}

// UpdateStock updates stock quantity
func (s *InventoryServer) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("UpdateStock", statusCode, time.Since(start))
	}()

	stock, err := s.service.UpdateStock(ctx, req.ProductId, req.Quantity, req.Reason)
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	return &pb.UpdateStockResponse{
		Stock: &pb.Stock{
			ProductId:   stock.ProductID,
			Available:   stock.Available,
			Reserved:    stock.Reserved,
			Total:       stock.Total,
			WarehouseId: stock.WarehouseID,
		},
	}, nil
}

// ReserveStock reserves stock for an order
func (s *InventoryServer) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("ReserveStock", statusCode, time.Since(start))
	}()

	// Convert proto items to service format
	items := make([]struct {
		ProductID string
		Quantity  int32
	}, len(req.Items))

	for i, item := range req.Items {
		items[i] = struct {
			ProductID string
			Quantity  int32
		}{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	orderID, err := s.service.ReserveStock(ctx, req.OrderId, items)
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	return &pb.ReserveStockResponse{
		ReservationId: orderID,
		Success:       true,
		Message:       "Stock reserved successfully",
	}, nil
}

// ReleaseStock releases reserved stock
func (s *InventoryServer) ReleaseStock(ctx context.Context, req *pb.ReleaseStockRequest) (*pb.ReleaseStockResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("ReleaseStock", statusCode, time.Since(start))
	}()

	err := s.service.ReleaseStock(ctx, req.OrderId, req.Reason)
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	return &pb.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

// CommitStock commits reserved stock
func (s *InventoryServer) CommitStock(ctx context.Context, req *pb.CommitStockRequest) (*pb.CommitStockResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("CommitStock", statusCode, time.Since(start))
	}()

	err := s.service.CommitStock(ctx, req.OrderId)
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	return &pb.CommitStockResponse{
		Success: true,
		Message: "Stock committed successfully",
	}, nil
}

// CheckAvailability checks if products are available
func (s *InventoryServer) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("CheckAvailability", statusCode, time.Since(start))
	}()

	// Convert proto items to service format
	items := make([]struct {
		ProductID string
		Quantity  int32
	}, len(req.Items))

	for i, item := range req.Items {
		items[i] = struct {
			ProductID string
			Quantity  int32
		}{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	available, unavailable, err := s.service.CheckAvailability(ctx, items)
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	// Convert unavailable items to proto format
	unavailableItems := make([]*pb.UnavailableItem, len(unavailable))
	for i, item := range unavailable {
		unavailableItems[i] = &pb.UnavailableItem{
			ProductId: item["product_id"].(string),
			Requested: item["requested"].(int32),
			Available: item["available"].(int32),
		}
	}

	return &pb.CheckAvailabilityResponse{
		Available:        available,
		UnavailableItems: unavailableItems,
	}, nil
}

// GetStockHistory retrieves stock movement history
func (s *InventoryServer) GetStockHistory(ctx context.Context, req *pb.GetStockHistoryRequest) (*pb.GetStockHistoryResponse, error) {
	start := time.Now()
	var statusCode string
	defer func() {
		middleware.RecordGRPCRequest("GetStockHistory", statusCode, time.Since(start))
	}()

	movements, total, err := s.service.GetStockHistory(ctx, req.ProductId, int(req.Limit), int(req.Offset))
	if err != nil {
		statusCode = "error"
		return nil, status.Error(codes.Internal, err.Error())
	}

	statusCode = "success"
	// Convert to proto format
	pbMovements := make([]*pb.StockMovement, len(movements))
	for i, m := range movements {
		pbMovements[i] = &pb.StockMovement{
			Id:             m.ID,
			ProductId:      m.ProductID,
			MovementType:   m.MovementType,
			Quantity:       m.Quantity,
			BeforeQuantity: m.BeforeQuantity,
			AfterQuantity:  m.AfterQuantity,
			ReferenceType:  m.ReferenceType,
			ReferenceId:    m.ReferenceID,
			Reason:         m.Reason,
			CreatedAt:      m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &pb.GetStockHistoryResponse{
		Movements: pbMovements,
		Total:     int32(total),
	}, nil
}
