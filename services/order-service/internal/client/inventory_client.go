package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/inventory_service"
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient struct {
	conn   *grpc.ClientConn
	client pb.InventoryServiceClient
	pool   *grpcpool.ConnectionPool // Connection pool support
}

func NewInventoryClient(endpoint sharedConfig.ServiceEndpoint) (*InventoryClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(endpoint.GRPCAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	return &InventoryClient{
		conn:   conn,
		client: pb.NewInventoryServiceClient(conn),
	}, nil
}

// NewInventoryClientWithPool creates a new inventory client with connection pooling support
func NewInventoryClientWithPool(endpoint sharedConfig.ServiceEndpoint, poolManager *grpcpool.Manager) (*InventoryClient, error) {
	pool, exists := poolManager.Get("inventory")
	if !exists {
		poolConfig := grpcpool.DefaultPoolConfig(endpoint.GRPCAddr)
		var err error
		pool, err = poolManager.GetOrCreate("inventory", poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create inventory service pool: %w", err)
		}
	}

	return &InventoryClient{
		pool: pool,
	}, nil
}

func (c *InventoryClient) Close() error {
	// If using pool, don't close individual connections
	if c.pool != nil {
		return nil
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// getClient returns a gRPC client, using pool if available
func (c *InventoryClient) getClient() (pb.InventoryServiceClient, error) {
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewInventoryServiceClient(conn), nil
	}
	return c.client, nil
}

// ReserveStock reserves stock for an order
func (c *InventoryClient) ReserveStock(ctx context.Context, orderID string, items []*pb.StockItem) (string, error) {
	client, err := c.getClient()
	if err != nil {
		return "", err
	}

	resp, err := client.ReserveStock(ctx, &pb.ReserveStockRequest{
		OrderId: orderID,
		Items:   items,
	})
	if err != nil {
		return "", fmt.Errorf("failed to reserve stock: %w", err)
	}

	if !resp.Success {
		return "", fmt.Errorf("failed to reserve stock: %s", resp.Message)
	}

	return resp.ReservationId, nil
}

// CommitStock commits reserved stock (after payment)
func (c *InventoryClient) CommitStock(ctx context.Context, reservationID string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	resp, err := client.CommitStock(ctx, &pb.CommitStockRequest{
		ReservationId: reservationID,
	})
	if err != nil {
		return fmt.Errorf("failed to commit stock: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to commit stock: %s", resp.Message)
	}

	return nil
}

// ReleaseStock releases reserved stock (on order cancel)
func (c *InventoryClient) ReleaseStock(ctx context.Context, reservationID string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	resp, err := client.ReleaseStock(ctx, &pb.ReleaseStockRequest{
		ReservationId: reservationID,
	})
	if err != nil {
		return fmt.Errorf("failed to release stock: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("failed to release stock: %s", resp.Message)
	}

	return nil
}

// CheckAvailability checks if sufficient stock is available
func (c *InventoryClient) CheckAvailability(ctx context.Context, items []*pb.StockItem) (bool, []*pb.UnavailableItem, error) {
	client, err := c.getClient()
	if err != nil {
		return false, nil, err
	}

	resp, err := client.CheckAvailability(ctx, &pb.CheckAvailabilityRequest{
		Items: items,
	})
	if err != nil {
		return false, nil, fmt.Errorf("failed to check availability: %w", err)
	}

	return resp.Available, resp.UnavailableItems, nil
}

// GetStock retrieves current stock level
func (c *InventoryClient) GetStock(ctx context.Context, productID string) (*pb.Stock, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetStock(ctx, &pb.GetStockRequest{
		ProductId: productID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get stock: %w", err)
	}

	return resp.Stock, nil
}
