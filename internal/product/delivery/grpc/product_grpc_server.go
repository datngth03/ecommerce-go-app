// internal/product/delivery/grpc/product_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/product/application"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated gRPC client
)

// ProductGRPCServer implements the product_client.ProductServiceServer interface.
type ProductGRPCServer struct {
	product_client.UnimplementedProductServiceServer // Embedded to satisfy all methods
	productService                                   application.ProductService
}

// NewProductGRPCServer creates a new instance of ProductGRPCServer.
func NewProductGRPCServer(svc application.ProductService) *ProductGRPCServer {
	return &ProductGRPCServer{
		productService: svc,
	}
}

// --- Product RPC Methods ---

// CreateProduct implements the gRPC CreateProduct method.
func (s *ProductGRPCServer) CreateProduct(ctx context.Context, req *product_client.CreateProductRequest) (*product_client.ProductResponse, error) {
	log.Printf("Received CreateProduct request for product: %s", req.GetName())
	resp, err := s.productService.CreateProduct(ctx, req)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetProductById implements the gRPC GetProductById method.
func (s *ProductGRPCServer) GetProductById(ctx context.Context, req *product_client.GetProductByIdRequest) (*product_client.ProductResponse, error) {
	log.Printf("Received GetProductById request for ID: %s", req.GetId())
	resp, err := s.productService.GetProductById(ctx, req)
	if err != nil {
		log.Printf("Error getting product by ID: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateProduct implements the gRPC UpdateProduct method.
func (s *ProductGRPCServer) UpdateProduct(ctx context.Context, req *product_client.UpdateProductRequest) (*product_client.ProductResponse, error) {
	log.Printf("Received UpdateProduct request for ID: %s", req.GetId())
	resp, err := s.productService.UpdateProduct(ctx, req)
	if err != nil {
		log.Printf("Error updating product: %v", err)
		return nil, err
	}
	return resp, nil
}

// DeleteProduct implements the gRPC DeleteProduct method.
func (s *ProductGRPCServer) DeleteProduct(ctx context.Context, req *product_client.DeleteProductRequest) (*product_client.DeleteProductResponse, error) {
	log.Printf("Received DeleteProduct request for ID: %s", req.GetId())
	resp, err := s.productService.DeleteProduct(ctx, req)
	if err != nil {
		log.Printf("Error deleting product: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListProducts implements the gRPC ListProducts method.
func (s *ProductGRPCServer) ListProducts(ctx context.Context, req *product_client.ListProductsRequest) (*product_client.ListProductsResponse, error) {
	log.Printf("Received ListProducts request (Category ID: %s, Limit: %d, Offset: %d)", req.GetCategoryId(), req.GetLimit(), req.GetOffset())
	resp, err := s.productService.ListProducts(ctx, req)
	if err != nil {
		log.Printf("Error listing products: %v", err)
		return nil, err
	}
	return resp, nil
}

// --- Category RPC Methods ---

// CreateCategory implements the gRPC CreateCategory method.
func (s *ProductGRPCServer) CreateCategory(ctx context.Context, req *product_client.CreateCategoryRequest) (*product_client.CategoryResponse, error) {
	log.Printf("Received CreateCategory request for category: %s", req.GetName())
	resp, err := s.productService.CreateCategory(ctx, req)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCategoryById implements the gRPC GetCategoryById method.
func (s *ProductGRPCServer) GetCategoryById(ctx context.Context, req *product_client.GetCategoryByIdRequest) (*product_client.CategoryResponse, error) {
	log.Printf("Received GetCategoryById request for ID: %s", req.GetId())
	resp, err := s.productService.GetCategoryById(ctx, req)
	if err != nil {
		log.Printf("Error getting category by ID: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListCategories implements the gRPC ListCategories method.
func (s *ProductGRPCServer) ListCategories(ctx context.Context, req *product_client.ListCategoriesRequest) (*product_client.ListCategoriesResponse, error) {
	log.Println("Received ListCategories request")
	resp, err := s.productService.ListCategories(ctx, req)
	if err != nil {
		log.Printf("Error listing categories: %v", err)
		return nil, err
	}
	return resp, nil
}
