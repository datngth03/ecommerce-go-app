package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient struct {
	conn   *grpc.ClientConn
	client pb.InventoryServiceClient
}

func NewInventoryClient(address string, timeout time.Duration) (*InventoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	return &InventoryClient{
		conn:   conn,
		client: pb.NewInventoryServiceClient(conn),
	}, nil
}

func (c *InventoryClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetStock retrieves stock information
func (c *InventoryClient) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	return c.client.GetStock(ctx, req)
}

// UpdateStock updates stock quantity
func (c *InventoryClient) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	return c.client.UpdateStock(ctx, req)
}

// ReserveStock reserves stock for an order
func (c *InventoryClient) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	return c.client.ReserveStock(ctx, req)
}

// ReleaseStock releases a reservation
func (c *InventoryClient) ReleaseStock(ctx context.Context, req *pb.ReleaseStockRequest) (*pb.ReleaseStockResponse, error) {
	return c.client.ReleaseStock(ctx, req)
}

// CommitStock commits a reservation
func (c *InventoryClient) CommitStock(ctx context.Context, req *pb.CommitStockRequest) (*pb.CommitStockResponse, error) {
	return c.client.CommitStock(ctx, req)
}

// CheckAvailability checks if products are available
func (c *InventoryClient) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	return c.client.CheckAvailability(ctx, req)
}

// GetStockHistory retrieves stock movement history
func (c *InventoryClient) GetStockHistory(ctx context.Context, req *pb.GetStockHistoryRequest) (*pb.GetStockHistoryResponse, error) {
	return c.client.GetStockHistory(ctx, req)
}
