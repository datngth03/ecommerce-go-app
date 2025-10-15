package client

import (
	"context"
	"fmt"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	sharedConfig "github.com/ecommerce-go-app/shared/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ProductClient struct {
	conn   *grpc.ClientConn
	client pb.ProductServiceClient
}

func NewProductClient(endpoint sharedConfig.ServiceEndpoint) (*ProductClient, error) {
	conn, err := grpc.Dial(endpoint.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service: %w", err)
	}

	return &ProductClient{
		conn:   conn,
		client: pb.NewProductServiceClient(conn),
	}, nil
}

func (c *ProductClient) Close() error {
	return c.conn.Close()
}

// GetProduct retrieves product details by ID
func (c *ProductClient) GetProduct(ctx context.Context, productID string) (*pb.Product, error) {
	resp, err := c.client.GetProduct(ctx, &pb.GetProductRequest{
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
