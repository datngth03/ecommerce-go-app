package clients

import (
	"context"
	"fmt"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/shared/pkg/grpcpool"
	sharedTracing "github.com/datngth03/ecommerce-go-app/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ProductClient wraps the gRPC client for product-service with connection pooling
type ProductClient struct {
	conn           *grpc.ClientConn         // Legacy: single connection
	pool           *grpcpool.ConnectionPool // New: connection pool
	ProductClient  pb.ProductServiceClient
	CategoryClient pb.CategoryServiceClient
	timeout        time.Duration
}

// NewProductClient creates a new product service gRPC client (legacy method)
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
		grpc.WithUnaryInterceptor(sharedTracing.UnaryClientInterceptor()),
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

// NewProductClientWithPool creates a new product service gRPC client with connection pooling
func NewProductClientWithPool(pool *grpcpool.ConnectionPool, timeout time.Duration) (*ProductClient, error) {
	// Get a connection from the pool to create the clients
	conn := pool.Get()

	return &ProductClient{
		pool:           pool,
		ProductClient:  pb.NewProductServiceClient(conn),
		CategoryClient: pb.NewCategoryServiceClient(conn),
		timeout:        timeout,
	}, nil
}

// Close closes the gRPC connection (no-op for pooled connections)
func (c *ProductClient) Close() error {
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

// getProductClient returns a product client using either pooled or direct connection
func (c *ProductClient) getProductClient() pb.ProductServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewProductServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.ProductClient
}

// getCategoryClient returns a category client using either pooled or direct connection
func (c *ProductClient) getCategoryClient() pb.CategoryServiceClient {
	// If using connection pool, recreate client with a connection from the pool
	if c.pool != nil {
		conn := c.pool.Get()
		return pb.NewCategoryServiceClient(conn)
	}

	// Legacy: use direct connection client
	return c.CategoryClient
}

// GetProduct retrieves a product by ID
func (c *ProductClient) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getProductClient()
	resp, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}

// ListProducts retrieves a list of products with pagination
func (c *ProductClient) ListProducts(ctx context.Context, page, pageSize int32, categoryID string) ([]*pb.Product, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getProductClient()
	resp, err := client.ListProducts(ctx, &pb.ListProductsRequest{
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

	client := c.getProductClient()
	resp, err := client.CreateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}

// UpdateProduct updates an existing product
func (c *ProductClient) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getProductClient()
	resp, err := client.UpdateProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Product, nil
}

// DeleteProduct deletes a product
func (c *ProductClient) DeleteProduct(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getProductClient()
	_, err := client.DeleteProduct(ctx, &pb.DeleteProductRequest{Id: id})
	return err
}

// ListCategories retrieves all categories
func (c *ProductClient) ListCategories(ctx context.Context) ([]*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getCategoryClient()
	resp, err := client.ListCategories(ctx, &pb.ListCategoriesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Categories, nil
}

// GetCategory retrieves a category by ID
func (c *ProductClient) GetCategory(ctx context.Context, id string) (*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getCategoryClient()
	resp, err := client.GetCategory(ctx, &pb.GetCategoryRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.Category, nil
}

// CreateCategory creates a new category
func (c *ProductClient) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getCategoryClient()
	resp, err := client.CreateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Category, nil
}

// UpdateCategory updates an existing category
func (c *ProductClient) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getCategoryClient()
	resp, err := client.UpdateCategory(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Category, nil
}

// DeleteCategory deletes a category
func (c *ProductClient) DeleteCategory(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	client := c.getCategoryClient()
	_, err := client.DeleteCategory(ctx, &pb.DeleteCategoryRequest{Id: id})
	return err
}
