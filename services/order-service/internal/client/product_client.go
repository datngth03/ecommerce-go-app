package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	sharedConfig "github.com/datngth03/ecommerce-go-app/shared/pkg/config"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	conn   *grpc.ClientConn
	client pb.ProductServiceClient
	pool   *grpcpool.ConnectionPool // Connection pool support
}

func NewProductClient(endpoint sharedConfig.ServiceEndpoint) (*ProductClient, error) {
	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(endpoint.GRPCAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	return &ProductClient{
		conn:   conn,
		client: pb.NewProductServiceClient(conn),
	}, nil
}

// NewProductClientWithPool creates a new product client with connection pooling support
func NewProductClientWithPool(endpoint sharedConfig.ServiceEndpoint, poolManager *grpcpool.Manager) (*ProductClient, error) {
	pool, exists := poolManager.Get("product")
	if !exists {
		poolConfig := grpcpool.DefaultPoolConfig(endpoint.GRPCAddr)
		var err error
		pool, err = poolManager.GetOrCreate("product", poolConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create product service pool: %w", err)
		}
	}

	return &ProductClient{
		pool: pool,
	}, nil
}

func (c *ProductClient) Close() error {
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
func (c *ProductClient) getClient() (pb.ProductServiceClient, error) {
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewProductServiceClient(conn), nil
	}
	return c.client, nil
}

// GetProduct retrieves product details by ID
func (c *ProductClient) GetProduct(ctx context.Context, productID string) (*pb.Product, error) {
	client, err := c.getClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.GetProduct(ctx, &pb.GetProductRequest{
		Id: productID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return resp.Product, nil
}

// CheckStock validates if product exists (stock management should be in Inventory Service)
func (c *ProductClient) CheckStock(ctx context.Context, productID string, quantity int32) (bool, error) {
	// For now, just check if product exists
	_, err := c.GetProduct(ctx, productID)
	if err != nil {
		return false, err
	}
	// TODO: Call Inventory Service for actual stock check
	return true, nil
}

// GetProducts retrieves multiple products by IDs
func (c *ProductClient) GetProducts(ctx context.Context, productIDs []string) ([]*pb.Product, error) {
	products := make([]*pb.Product, 0, len(productIDs))

	for _, id := range productIDs {
		product, err := c.GetProduct(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get product %s: %w", id, err)
		}
		products = append(products, product)
	}

	return products, nil
}
