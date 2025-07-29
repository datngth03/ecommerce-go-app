// internal/product/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"log" // Thêm import log
	"time"

	"github.com/google/uuid"

	"github.com/datngth03/ecommerce-go-app/internal/product/domain"
	"github.com/datngth03/ecommerce-go-app/internal/product/infrastructure/messaging" // THÊM: Import messaging package
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product"         // Generated gRPC client
)

// ProductService defines the application service interface for product-related operations.
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
type productService struct {
	productRepo    domain.ProductRepository
	categoryRepo   domain.CategoryRepository
	eventPublisher messaging.ProductEventPublisher // THÊM: Event Publisher
}

// NewProductService creates a new instance of ProductService.
func NewProductService(
	productRepo domain.ProductRepository,
	categoryRepo domain.CategoryRepository,
	eventPublisher messaging.ProductEventPublisher, // THÊM: eventPublisher
) ProductService {
	return &productService{
		productRepo:    productRepo,
		categoryRepo:   categoryRepo,
		eventPublisher: eventPublisher, // THÊM:
	}
}

// --- Product Use Cases ---

// CreateProduct handles the creation of a new product.
func (s *productService) CreateProduct(ctx context.Context, req *product_client.CreateProductRequest) (*product_client.ProductResponse, error) {
	// Basic validation
	if req.GetName() == "" || req.GetPrice() <= 0 || req.GetCategoryId() == "" {
		return nil, errors.New("product name, price, and category ID are required and price must be positive")
	}

	// Check if category exists
	category, err := s.categoryRepo.FindByID(ctx, req.GetCategoryId())
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) { // Assuming "category not found" error from repo
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

	// THÊM: Publish ProductCreated event to message queue (e.g., for Search Service to index, Inventory Service to initialize stock)
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishProductCreated(ctx, product); err != nil {
			log.Printf("WARNING: Failed to publish ProductCreated event for product %s: %v", product.ID, err)
			// Không trả về lỗi, vì việc phát sự kiện có thể không quan trọng bằng việc lưu sản phẩm vào DB
		}
	}

	return &product_client.ProductResponse{
		Product: &product_client.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			CategoryId:  product.CategoryID,
			ImageUrls:   product.ImageURLs,
			// StockQuantity: product.StockQuantity, // StockQuantity is from Inventory Service, not directly from Product
			CreatedAt: product.CreatedAt.Format(time.RFC3339),
			UpdatedAt: product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetProductById handles retrieving product details.
func (s *productService) GetProductById(ctx context.Context, req *product_client.GetProductByIdRequest) (*product_client.ProductResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("product ID is required")
	}

	product, err := s.productRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) { // Assuming "product not found" error from repo
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}

	return &product_client.ProductResponse{
		Product: &product_client.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			CategoryId:  product.CategoryID,
			ImageUrls:   product.ImageURLs,
			// StockQuantity: product.StockQuantity,
			CreatedAt: product.CreatedAt.Format(time.RFC3339),
			UpdatedAt: product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateProduct handles updating product information.
func (s *productService) UpdateProduct(ctx context.Context, req *product_client.UpdateProductRequest) (*product_client.ProductResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("product ID is required for update")
	}
	if req.GetName() == "" || req.GetPrice() <= 0 || req.GetCategoryId() == "" {
		return nil, errors.New("product name, price, and category ID are required and price must be positive")
	}

	product, err := s.productRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to find product for update: %w", err)
	}

	// Check if new category exists if category ID is changed
	if product.CategoryID != req.GetCategoryId() {
		category, err := s.categoryRepo.FindByID(ctx, req.GetCategoryId())
		if err != nil {
			if errors.Is(err, domain.ErrCategoryNotFound) {
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

	// THÊM: Publish ProductUpdated event
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishProductUpdated(ctx, product); err != nil {
			log.Printf("WARNING: Failed to publish ProductUpdated event for product %s: %v", product.ID, err)
		}
	}

	return &product_client.ProductResponse{
		Product: &product_client.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			CategoryId:  product.CategoryID,
			ImageUrls:   product.ImageURLs,
			// StockQuantity: product.StockQuantity,
			CreatedAt: product.CreatedAt.Format(time.RFC3339),
			UpdatedAt: product.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// DeleteProduct handles deleting a product.
func (s *productService) DeleteProduct(ctx context.Context, req *product_client.DeleteProductRequest) (*product_client.DeleteProductResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("product ID is required for deletion")
	}

	if err := s.productRepo.Delete(ctx, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			return &product_client.DeleteProductResponse{Success: false, Message: "Product not found"}, nil
		}
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	// THÊM: Publish ProductDeleted event
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishProductDeleted(ctx, req.GetId()); err != nil {
			log.Printf("WARNING: Failed to publish ProductDeleted event for product %s: %v", req.GetId(), err)
		}
	}

	return &product_client.DeleteProductResponse{Success: true, Message: "Product deleted successfully"}, nil
}

// ListProducts handles listing products with pagination and filters.
func (s *productService) ListProducts(ctx context.Context, req *product_client.ListProductsRequest) (*product_client.ListProductsResponse, error) {
	products, totalCount, err := s.productRepo.FindAll(ctx, req.GetCategoryId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	productResponses := make([]*product_client.Product, len(products))
	for i, p := range products {
		productResponses[i] = &product_client.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			CategoryId:  p.CategoryID,
			ImageUrls:   p.ImageURLs,
			// StockQuantity: p.StockQuantity, // Stock quantity should come from Inventory Service
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
			UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &product_client.ListProductsResponse{
		Products:   productResponses,
		TotalCount: totalCount,
	}, nil
}

// --- Category Use Cases ---

// CreateCategory handles the creation of a new category.
func (s *productService) CreateCategory(ctx context.Context, req *product_client.CreateCategoryRequest) (*product_client.CategoryResponse, error) {
	if req.GetName() == "" {
		return nil, errors.New("category name is required")
	}

	// Check if category with same name already exists
	existingCategory, err := s.categoryRepo.FindByName(ctx, req.GetName())
	if err != nil && !errors.Is(err, domain.ErrCategoryNotFound) { // Only return error if it's not a "not found" error
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
func (s *productService) GetCategoryById(ctx context.Context, req *product_client.GetCategoryByIdRequest) (*product_client.CategoryResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("category ID is required")
	}

	category, err := s.categoryRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
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
