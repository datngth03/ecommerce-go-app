package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	conn   *grpc.ClientConn
	client pb.OrderServiceClient
	pool   *grpcpool.ConnectionPool // Connection pool support
}

func NewOrderClient(endpoint sharedConfig.ServiceEndpoint) (*OrderClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(endpoint.GRPCAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &OrderClient{
		conn:   conn,
		client: pb.NewOrderServiceClient(conn),
	}, nil
}

// NewOrderClientWithPool creates a new order client with connection pooling support
func NewOrderClientWithPool(endpoint sharedConfig.ServiceEndpoint, poolManager *grpcpool.Manager) (*OrderClient, error) {
	pool, exists := poolManager.Get("order")
	if !exists {
		poolConfig := grpcpool.DefaultPoolConfig(endpoint.GRPCAddr)
		var err error
		pool, err = poolManager.GetOrCreate("order", poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create order service pool: %w", err)
		}
	}

	return &OrderClient{
		pool: pool,
	}, nil
}

func (c *OrderClient) Close() error {
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
func (c *OrderClient) getClient() (pb.OrderServiceClient, error) {
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewOrderServiceClient(conn), nil
	}
	return c.client, nil
}

// GetOrder retrieves order details by ID
func (c *OrderClient) GetOrder(ctx context.Context, orderID string) (*pb.Order, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetOrder(ctx, &pb.GetOrderRequest{
		Id: orderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return resp.Order, nil
}

// UpdateOrderStatus updates the status of an order
func (c *OrderClient) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	client, err := c.getClient()
	if err != nil {
		return err
	}

	_, err = client.UpdateOrderStatus(ctx, &pb.UpdateOrderStatusRequest{
		Id:     orderID,
		Status: status,
	})
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// ListOrders retrieves orders for a user
func (c *OrderClient) ListOrders(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*pb.Order, int64, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.ListOrders(ctx, &pb.ListOrdersRequest{
		UserId:   userID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	return resp.Orders, resp.TotalCount, nil
}

// GetOrderByID retrieves order for validation
func (c *OrderClient) GetOrderByID(ctx context.Context, orderID string) (*pb.Order, error) {
	return c.GetOrder(ctx, orderID)
}

// ValidateOrder checks if order exists and can be paid
func (c *OrderClient) ValidateOrder(ctx context.Context, orderID string) (bool, error) {
	order, err := c.GetOrder(ctx, orderID)
	if err != nil {
		return false, err
	}

	// Check if order status allows payment
	validStatuses := map[string]bool{
		"PENDING":   true,
		"CONFIRMED": true,
	}

	return validStatuses[order.Status], nil
}
