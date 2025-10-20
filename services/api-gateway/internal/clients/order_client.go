package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient struct {
	conn   *grpc.ClientConn
	client pb.OrderServiceClient
}

func NewOrderClient(address string, timeout time.Duration) (*OrderClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &OrderClient{
		conn:   conn,
		client: pb.NewOrderServiceClient(conn),
	}, nil
}

func (c *OrderClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// CreateOrder creates a new order
func (c *OrderClient) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	return c.client.CreateOrder(ctx, req)
}

// GetOrder retrieves an order by ID
func (c *OrderClient) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	return c.client.GetOrder(ctx, req)
}

// ListOrders lists orders for a user
func (c *OrderClient) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	return c.client.ListOrders(ctx, req)
}

// UpdateOrderStatus updates order status
func (c *OrderClient) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	return c.client.UpdateOrderStatus(ctx, req)
}

// CancelOrder cancels an order
func (c *OrderClient) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) error {
	_, err := c.client.CancelOrder(ctx, req)
	return err
}

// Cart operations
func (c *OrderClient) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.CartResponse, error) {
	return c.client.AddToCart(ctx, req)
}

func (c *OrderClient) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
	return c.client.GetCart(ctx, req)
}

func (c *OrderClient) UpdateCartItem(ctx context.Context, req *pb.UpdateCartItemRequest) (*pb.CartResponse, error) {
	return c.client.UpdateCartItem(ctx, req)
}

func (c *OrderClient) RemoveFromCart(ctx context.Context, req *pb.RemoveFromCartRequest) (*pb.CartResponse, error) {
	return c.client.RemoveFromCart(ctx, req)
}

func (c *OrderClient) ClearCart(ctx context.Context, req *pb.ClearCartRequest) error {
	_, err := c.client.ClearCart(ctx, req)
	return err
}
