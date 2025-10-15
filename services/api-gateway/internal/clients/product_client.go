package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ProductClient wraps the gRPC client for product-service
type ProductClient struct {
	conn           *grpc.ClientConn
	ProductClient  pb.ProductServiceClient
	CategoryClient pb.CategoryServiceClient
	timeout        time.Duration
}

// NewProductClient creates a new product service gRPC client
func NewProductClient(addr string, timeout time.Duration) (*ProductClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // TODO: Use TLS in production
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product service at %s: %w", addr, err)
	}

	return &ProductClient{
		conn:           conn,
		ProductClient:  pb.NewProductServiceClient(conn),
		CategoryClient: pb.NewCategoryServiceClient(conn),
		timeout:        timeout,
	}, nil
}

// Close closes the gRPC connection
func (c *ProductClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetProduct retrieves a product by ID
func (c *ProductClient) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.ProductClient.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}

// ListProducts retrieves a list of products with pagination
func (c *ProductClient) ListProducts(ctx context.Context, page, pageSize int32, categoryID string) ([]*pb.Product, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.ProductClient.ListProducts(ctx, &pb.ListProductsRequest{
		Page:       page,
		PageSize:   pageSize,
		CategoryId: categoryID,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.Products, resp.TotalCount, nil
}

// CreateProduct creates a new product
func (c *ProductClient) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.ProductClient.CreateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}

// UpdateProduct updates an existing product
func (c *ProductClient) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.ProductClient.UpdateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}

// DeleteProduct deletes a product
func (c *ProductClient) DeleteProduct(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.ProductClient.DeleteProduct(ctx, &pb.DeleteProductRequest{Id: id})
	return err
}

// ListCategories retrieves all categories
func (c *ProductClient) ListCategories(ctx context.Context) ([]*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.CategoryClient.ListCategories(ctx, &pb.ListCategoriesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Categories, nil
}

// GetCategory retrieves a category by ID
func (c *ProductClient) GetCategory(ctx context.Context, id string) (*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.CategoryClient.GetCategory(ctx, &pb.GetCategoryRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Category, nil
}

// CreateCategory creates a new category
func (c *ProductClient) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.CategoryClient.CreateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Category, nil
}

// UpdateCategory updates an existing category
func (c *ProductClient) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.CategoryClient.UpdateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Category, nil
}

// DeleteCategory deletes a category
func (c *ProductClient) DeleteCategory(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.CategoryClient.DeleteCategory(ctx, &pb.DeleteCategoryRequest{Id: id})
	return err
}
