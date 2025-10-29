package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderClient wraps the gRPC client for order-service with connection pooling
type OrderClient struct {
	conn   *grpc.ClientConn         // Legacy: single connection
	pool   *grpcpool.ConnectionPool // New: connection pool
	client pb.OrderServiceClient
}

// NewOrderClient creates a new order service gRPC client (legacy method)
func NewOrderClient(address string, timeout time.Duration) (*OrderClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &OrderClient{
		conn:   conn,
		client: pb.NewOrderServiceClient(conn),
	}, nil
}

// NewOrderClientWithPool creates a new order service gRPC client with connection pooling
func NewOrderClientWithPool(pool *grpcpool.ConnectionPool, timeout time.Duration) (*OrderClient, error) {
	// Get a connection from the pool to create the client
	conn := pool.Get()

	return &OrderClient{
		pool:   pool,
		client: pb.NewOrderServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection (no-op for pooled connections)
func (c *OrderClient) Close() error {
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
func (c *OrderClient) getClient() pb.OrderServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewOrderServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.client
}

// CreateOrder creates a new order
func (c *OrderClient) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	client := c.getClient()
	return client.CreateOrder(ctx, req)
}

// GetOrder retrieves an order by ID
func (c *OrderClient) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	client := c.getClient()
	return client.GetOrder(ctx, req)
}

// ListOrders lists orders for a user
func (c *OrderClient) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	client := c.getClient()
	return client.ListOrders(ctx, req)
}

// UpdateOrderStatus updates order status
func (c *OrderClient) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	client := c.getClient()
	return client.UpdateOrderStatus(ctx, req)
}

// CancelOrder cancels an order
func (c *OrderClient) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) error {
	client := c.getClient()
	_, err := client.CancelOrder(ctx, req)
	return err
}

// Cart operations
func (c *OrderClient) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.CartResponse, error) {
	client := c.getClient()
	return client.AddToCart(ctx, req)
}

func (c *OrderClient) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
	client := c.getClient()
	return client.GetCart(ctx, req)
}

func (c *OrderClient) UpdateCartItem(ctx context.Context, req *pb.UpdateCartItemRequest) (*pb.CartResponse, error) {
	client := c.getClient()
	return client.UpdateCartItem(ctx, req)
}

func (c *OrderClient) RemoveFromCart(ctx context.Context, req *pb.RemoveFromCartRequest) (*pb.CartResponse, error) {
	client := c.getClient()
	return client.RemoveFromCart(ctx, req)
}

func (c *OrderClient) ClearCart(ctx context.Context, req *pb.ClearCartRequest) error {
	client := c.getClient()
	_, err := client.ClearCart(ctx, req)
	return err
}
