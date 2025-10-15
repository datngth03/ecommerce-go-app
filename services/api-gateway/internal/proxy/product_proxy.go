package proxy

import (
	"context"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
)

// ProductProxy adapts product client for HTTP handlers
type ProductProxy struct {
	client *clients.ProductClient
}

// NewProductProxy creates a new product proxy
func NewProductProxy(client *clients.ProductClient) *ProductProxy {
	return &ProductProxy{client: client}
}

// GetProduct retrieves a product by ID
func (p *ProductProxy) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	return p.client.GetProduct(ctx, id)
}

// ListProducts retrieves products with pagination
func (p *ProductProxy) ListProducts(ctx context.Context, page, pageSize int32, categoryID string) ([]*pb.Product, int64, error) {
	return p.client.ListProducts(ctx, page, pageSize, categoryID)
}

// CreateProduct creates a new product
func (p *ProductProxy) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	return p.client.CreateProduct(ctx, req)
}

// UpdateProduct updates an existing product
func (p *ProductProxy) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	return p.client.UpdateProduct(ctx, req)
}

// DeleteProduct deletes a product
func (p *ProductProxy) DeleteProduct(ctx context.Context, id string) error {
	return p.client.DeleteProduct(ctx, id)
}

// GetCategory retrieves a category by ID
func (p *ProductProxy) GetCategory(ctx context.Context, id string) (*pb.Category, error) {
	return p.client.GetCategory(ctx, id)
}

// ListCategories retrieves all categories
func (p *ProductProxy) ListCategories(ctx context.Context) ([]*pb.Category, error) {
	return p.client.ListCategories(ctx)
}

// CreateCategory creates a new category
func (p *ProductProxy) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	return p.client.CreateCategory(ctx, req)
}

// UpdateCategory updates an existing category
func (p *ProductProxy) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	return p.client.UpdateCategory(ctx, req)
}

// DeleteCategory deletes a category
func (p *ProductProxy) DeleteCategory(ctx context.Context, id string) error {
	return p.client.DeleteCategory(ctx, id)
}
