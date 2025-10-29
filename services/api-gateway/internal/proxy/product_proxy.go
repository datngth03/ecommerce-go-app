package proxy

import (
	"context"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/product_service"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/clients"
	"github.com/datngth03/ecommerce-go-app/services/api-gateway/internal/metrics"
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
	start := time.Now()
	resp, err := p.client.GetProduct(ctx, id)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "GetProduct", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// ListProducts retrieves products with pagination
func (p *ProductProxy) ListProducts(ctx context.Context, page, pageSize int32, categoryID string) ([]*pb.Product, int64, error) {
	start := time.Now()
	products, total, err := p.client.ListProducts(ctx, page, pageSize, categoryID)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "ListProducts", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return products, total, err
}

// CreateProduct creates a new product
func (p *ProductProxy) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.Product, error) {
	start := time.Now()
	resp, err := p.client.CreateProduct(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "CreateProduct", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// UpdateProduct updates an existing product
func (p *ProductProxy) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.Product, error) {
	start := time.Now()
	resp, err := p.client.UpdateProduct(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "UpdateProduct", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// DeleteProduct deletes a product
func (p *ProductProxy) DeleteProduct(ctx context.Context, id string) error {
	start := time.Now()
	err := p.client.DeleteProduct(ctx, id)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "DeleteProduct", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return err
}

// GetCategory retrieves a category by ID
func (p *ProductProxy) GetCategory(ctx context.Context, id string) (*pb.Category, error) {
	start := time.Now()
	resp, err := p.client.GetCategory(ctx, id)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "GetCategory", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// ListCategories retrieves all categories
func (p *ProductProxy) ListCategories(ctx context.Context) ([]*pb.Category, error) {
	start := time.Now()
	resp, err := p.client.ListCategories(ctx)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "ListCategories", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// CreateCategory creates a new category
func (p *ProductProxy) CreateCategory(ctx context.Context, req *pb.CreateCategoryRequest) (*pb.Category, error) {
	start := time.Now()
	resp, err := p.client.CreateCategory(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "CreateCategory", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// UpdateCategory updates an existing category
func (p *ProductProxy) UpdateCategory(ctx context.Context, req *pb.UpdateCategoryRequest) (*pb.Category, error) {
	start := time.Now()
	resp, err := p.client.UpdateCategory(ctx, req)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "UpdateCategory", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return resp, err
}

// DeleteCategory deletes a category
func (p *ProductProxy) DeleteCategory(ctx context.Context, id string) error {
	start := time.Now()
	err := p.client.DeleteCategory(ctx, id)

	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordGRPCClientRequest("product-service", "DeleteCategory", status, time.Since(start))
	metrics.RecordProxyRequest("product-service", status, time.Since(start))

	return err
}
