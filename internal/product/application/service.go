// internal/product/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Generated gRPC client
	"github.com/google/uuid"                                                  // For generating UUIDs
)

// ProductService defines the application service interface for product-related operations.
// ProductService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến sản phẩm.
type ProductService interface {
	CreateProduct(ctx context.Context, req *product_client.CreateProductRequest) (*product_client.ProductResponse, error)
	GetProductById(ctx context.Context, req *product_client.GetProductByIdRequest) (*product_client.ProductResponse, error)
	UpdateProduct(ctx context.Context, req *product_client.UpdateProductRequest) (*product_client.ProductResponse, error)
	DeleteProduct(ctx context.Context, req *product_client.DeleteProductRequest) (*product_client.DeleteProductResponse, error)
	ListProducts(ctx context.Context, req *product_client.ListProductsRequest) (*product_client.ListProductsResponse, error)

	CreateCategory(ctx context.Context, req *product_client.CreateCategoryRequest) (*product_client.CategoryResponse, error)
	GetCategoryById(ctx context.Context, req *product_client.GetCategoryByIdRequest) (*product_client.CategoryResponse, error)
	ListCategories(ctx context.Context, req *product_client.ListCategoriesRequest) (*product_client.ListCategoriesResponse, error)
}

// productService implements the ProductService interface.
// productService triển khai interface ProductService.
type productService struct {
	productRepo  domain.ProductRepository
	categoryRepo domain.CategoryRepository
	// TODO: Add other dependencies like event publisher, inventory client
	// Thêm các dependency khác như trình phát sự kiện, client Inventory
}

// NewProductService creates a new instance of ProductService.
// NewProductService tạo một thể hiện mới của ProductService.
func NewProductService(productRepo domain.ProductRepository, categoryRepo domain.CategoryRepository) ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// --- Product Use Cases ---

// CreateProduct handles the creation of a new product.
// CreateProduct xử lý việc tạo một sản phẩm mới.
func (s *productService) CreateProduct(ctx context.Context, req *product_client.CreateProductRequest) (*product_client.ProductResponse, error) {
	// Basic validation
	if req.GetName() == "" || req.GetPrice() <= 0 || req.GetCategoryId() == "" {
		return nil, errors.New("product name, price, and category ID are required and price must be positive")
	}

	// Check if category exists
	category, err := s.categoryRepo.FindByID(ctx, req.GetCategoryId())
	if err != nil {
		if errors.Is(err, errors.New("category not found")) { // Assuming "category not found" error from repo
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if category == nil { // Double check in case error is nil but category is not found
		return nil, errors.New("category not found")
	}

	// Create new product domain entity
	productID := uuid.New().String()
	product := domain.NewProduct(
		productID,
		req.GetName(),
		req.GetDescription(),
		req.GetPrice(),
		req.GetCategoryId(),
		req.GetImageUrls(),
	)
	product.StockQuantity = 0 // Initial stock is 0, will be updated by Inventory Service

	// Save product to repository
	if err := s.productRepo.Save(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	// TODO: Publish ProductCreated event to message queue (e.g., for Inventory Service to initialize stock)

	return &product_client.ProductResponse{
		Product: &product_client.Product{
			Id:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			CategoryId:    product.CategoryID,
			ImageUrls:     product.ImageURLs,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetProductById handles retrieving product details.
// GetProductById xử lý việc lấy chi tiết sản phẩm.
func (s *productService) GetProductById(ctx context.Context, req *product_client.GetProductByIdRequest) (*product_client.ProductResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("product ID is required")
	}

	product, err := s.productRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, errors.New("product not found")) { // Assuming "product not found" error from repo
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}

	return &product_client.ProductResponse{
		Product: &product_client.Product{
			Id:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			CategoryId:    product.CategoryID,
			ImageUrls:     product.ImageURLs,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateProduct handles updating product information.
// UpdateProduct xử lý việc cập nhật thông tin sản phẩm.
func (s *productService) UpdateProduct(ctx context.Context, req *product_client.UpdateProductRequest) (*product_client.ProductResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("product ID is required for update")
	}
	if req.GetName() == "" || req.GetPrice() <= 0 || req.GetCategoryId() == "" {
		return nil, errors.New("product name, price, and category ID are required and price must be positive")
	}

	product, err := s.productRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, errors.New("product not found")) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to find product for update: %w", err)
	}

	// Check if new category exists if category ID is changed
	if product.CategoryID != req.GetCategoryId() {
		category, err := s.categoryRepo.FindByID(ctx, req.GetCategoryId())
		if err != nil {
			if errors.Is(err, errors.New("category not found")) {
				return nil, errors.New("new category not found")
			}
			return nil, fmt.Errorf("failed to check new category existence: %w", err)
		}
		if category == nil {
			return nil, errors.New("new category not found")
		}
	}

	product.UpdateProductInfo(
		req.GetName(),
		req.GetDescription(),
		req.GetPrice(),
		req.GetCategoryId(),
		req.GetImageUrls(),
	)

	if err := s.productRepo.Save(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// TODO: Publish ProductUpdated event

	return &product_client.ProductResponse{
		Product: &product_client.Product{
			Id:            product.ID,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			CategoryId:    product.CategoryID,
			ImageUrls:     product.ImageURLs,
			StockQuantity: product.StockQuantity,
			CreatedAt:     product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// DeleteProduct handles deleting a product.
// DeleteProduct xử lý việc xóa một sản phẩm.
func (s *productService) DeleteProduct(ctx context.Context, req *product_client.DeleteProductRequest) (*product_client.DeleteProductResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("product ID is required for deletion")
	}

	if err := s.productRepo.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, errors.New("product not found")) {
			return &product_client.DeleteProductResponse{Success: false, Message: "Product not found"}, nil
		}
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	// TODO: Publish ProductDeleted event

	return &product_client.DeleteProductResponse{Success: true, Message: "Product deleted successfully"}, nil
}

// ListProducts handles listing products with pagination and filters.
// ListProducts xử lý việc liệt kê sản phẩm với phân trang và bộ lọc.
func (s *productService) ListProducts(ctx context.Context, req *product_client.ListProductsRequest) (*product_client.ListProductsResponse, error) {
	products, totalCount, err := s.productRepo.FindAll(ctx, req.GetCategoryId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	productResponses := make([]*product_client.Product, len(products))
	for i, p := range products {
		productResponses[i] = &product_client.Product{
			Id:            p.ID,
			Name:          p.Name,
			Description:   p.Description,
			Price:         p.Price,
			CategoryId:    p.CategoryID,
			ImageUrls:     p.ImageURLs,
			StockQuantity: p.StockQuantity,
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     p.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &product_client.ListProductsResponse{
		Products:   productResponses,
		TotalCount: totalCount,
	}, nil
}

// --- Category Use Cases ---

// CreateCategory handles the creation of a new category.
// CreateCategory xử lý việc tạo một danh mục mới.
func (s *productService) CreateCategory(ctx context.Context, req *product_client.CreateCategoryRequest) (*product_client.CategoryResponse, error) {
	if req.GetName() == "" {
		return nil, errors.New("category name is required")
	}

	// Check if category with same name already exists
	existingCategory, err := s.categoryRepo.FindByName(ctx, req.GetName())
	if err != nil && err.Error() != "category not found" {
		return nil, fmt.Errorf("failed to check existing category: %w", err)
	}
	if existingCategory != nil {
		return nil, errors.New("category with this name already exists")
	}

	categoryID := uuid.New().String()
	category := domain.NewCategory(categoryID, req.GetName(), req.GetDescription())

	if err := s.categoryRepo.Save(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to save category: %w", err)
	}

	return &product_client.CategoryResponse{
		Category: &product_client.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetCategoryById handles retrieving category details.
// GetCategoryById xử lý việc lấy chi tiết danh mục.
func (s *productService) GetCategoryById(ctx context.Context, req *product_client.GetCategoryByIdRequest) (*product_client.CategoryResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("category ID is required")
	}

	category, err := s.categoryRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, errors.New("category not found")) {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("failed to retrieve category: %w", err)
	}

	return &product_client.CategoryResponse{
		Category: &product_client.Category{
			Id:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			CreatedAt:   category.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   category.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ListCategories handles listing all categories.
// ListCategories xử lý việc liệt kê tất cả các danh mục.
func (s *productService) ListCategories(ctx context.Context, req *product_client.ListCategoriesRequest) (*product_client.ListCategoriesResponse, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	categoryResponses := make([]*product_client.Category, len(categories))
	for i, c := range categories {
		categoryResponses[i] = &product_client.Category{
			Id:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &product_client.ListCategoriesResponse{
		Categories: categoryResponses,
	}, nil
}
