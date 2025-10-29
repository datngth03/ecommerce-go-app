package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InventoryClient wraps the gRPC client for inventory-service with connection pooling
type InventoryClient struct {
	conn   *grpc.ClientConn         // Legacy: single connection
	pool   *grpcpool.ConnectionPool // New: connection pool
	client pb.InventoryServiceClient
}

// NewInventoryClient creates a new inventory service gRPC client (legacy method)
func NewInventoryClient(address string, timeout time.Duration) (*InventoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	return &InventoryClient{
		conn:   conn,
		client: pb.NewInventoryServiceClient(conn),
	}, nil
}

// NewInventoryClientWithPool creates a new inventory service gRPC client with connection pooling
func NewInventoryClientWithPool(pool *grpcpool.ConnectionPool, timeout time.Duration) (*InventoryClient, error) {
	// Get a connection from the pool to create the client
	conn := pool.Get()

	return &InventoryClient{
		pool:   pool,
		client: pb.NewInventoryServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection (no-op for pooled connections)
func (c *InventoryClient) Close() error {
	// If using connection pool, connections are managed by the pool
	if c.pool != nil {
		return nil
	}

	// Legacy: close single connection
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getClient returns a client using either pooled or direct connection
func (c *InventoryClient) getClient() pb.InventoryServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewInventoryServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.client
}

// GetStock retrieves stock information
func (c *InventoryClient) GetStock(ctx context.Context, req *pb.GetStockRequest) (*pb.GetStockResponse, error) {
	client := c.getClient()
	return client.GetStock(ctx, req)
}

// UpdateStock updates stock quantity
func (c *InventoryClient) UpdateStock(ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	client := c.getClient()
	return client.UpdateStock(ctx, req)
}

// ReserveStock reserves stock for an order
func (c *InventoryClient) ReserveStock(ctx context.Context, req *pb.ReserveStockRequest) (*pb.ReserveStockResponse, error) {
	client := c.getClient()
	return client.ReserveStock(ctx, req)
}

// ReleaseStock releases a reservation
func (c *InventoryClient) ReleaseStock(ctx context.Context, req *pb.ReleaseStockRequest) (*pb.ReleaseStockResponse, error) {
	client := c.getClient()
	return client.ReleaseStock(ctx, req)
}

// CommitStock commits a reservation
func (c *InventoryClient) CommitStock(ctx context.Context, req *pb.CommitStockRequest) (*pb.CommitStockResponse, error) {
	client := c.getClient()
	return client.CommitStock(ctx, req)
}

// CheckAvailability checks if products are available
func (c *InventoryClient) CheckAvailability(ctx context.Context, req *pb.CheckAvailabilityRequest) (*pb.CheckAvailabilityResponse, error) {
	client := c.getClient()
	return client.CheckAvailability(ctx, req)
}

// GetStockHistory retrieves stock movement history
func (c *InventoryClient) GetStockHistory(ctx context.Context, req *pb.GetStockHistoryRequest) (*pb.GetStockHistoryResponse, error) {
	client := c.getClient()
	return client.GetStockHistory(ctx, req)
}
