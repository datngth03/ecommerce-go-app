package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/repository"
)

type ProductService struct {
	repo *repository.Repository
}

func NewProductService(repo *repository.Repository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.ProductResponse, error) {
	// Validate input
	if err := s.validateCreateProductRequest(req); err != nil {
		return nil, err
	}

	// Check if category exists
	exists, err := s.repo.Category.ExistsByID(ctx, req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("category not found")
	}

	// Check if product name already exists
	nameExists, err := s.repo.Product.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check product name: %w", err)
	}
	if nameExists {
		return nil, fmt.Errorf("product with name '%s' already exists", req.Name)
	}

	// Create product
	product := &models.Product{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		ImageURL:    strings.TrimSpace(req.ImageURL),
		IsActive:    true,
	}

	if err := s.repo.Product.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Get created product with category
	createdProduct, err := s.repo.Product.GetByID(ctx, product.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created product: %w", err)
	}

	response := createdProduct.ToResponse()
	return &response, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*models.ProductResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	product, err := s.repo.Product.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*models.ProductResponse, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, fmt.Errorf("product slug is required")
	}

	product, err := s.repo.Product.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id string, req *models.UpdateProductRequest) (*models.ProductResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("product ID is required")
	}

	// Validate input
	if err := s.validateUpdateProductRequest(req); err != nil {
		return nil, err
	}

	// Get existing product
	existingProduct, err := s.repo.Product.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if category exists
	if req.CategoryID != existingProduct.CategoryID {
		exists, err := s.repo.Category.ExistsByID(ctx, req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check category existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("category not found")
		}
	}

	// Check if product name already exists (excluding current product)
	if req.Name != existingProduct.Name {
		nameExists, err := s.repo.Product.ExistsByName(ctx, req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check product name: %w", err)
		}
		if nameExists {
			return nil, fmt.Errorf("product with name '%s' already exists", req.Name)
		}
	}

	// Update product
	existingProduct.Name = strings.TrimSpace(req.Name)
	existingProduct.Description = strings.TrimSpace(req.Description)
	existingProduct.Price = req.Price
	existingProduct.CategoryID = req.CategoryID
	existingProduct.ImageURL = strings.TrimSpace(req.ImageURL)
	existingProduct.IsActive = req.IsActive

	if err := s.repo.Product.Update(ctx, existingProduct); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	// Get updated product with category
	updatedProduct, err := s.repo.Product.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated product: %w", err)
	}

	response := updatedProduct.ToResponse()
	return &response, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product ID is required")
	}

	// Check if product exists
	_, err := s.repo.Product.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Product.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *models.ListProductsRequest) (*models.ListProductsResponse, error) {
	// Validate and set defaults
	if err := s.validateListProductsRequest(req); err != nil {
		return nil, err
	}

	// If category_id is provided, check if it exists
	if req.CategoryID != "" {
		exists, err := s.repo.Category.ExistsByID(ctx, req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check category existence: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("category not found")
		}
	}

	// Get products
	products, total, err := s.repo.Product.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to response
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &models.ListProductsResponse{
		Products:   productResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ProductService) ListProductsByCategory(ctx context.Context, categoryID string, req *models.ListProductsRequest) (*models.ListProductsResponse, error) {
	if strings.TrimSpace(categoryID) == "" {
		return nil, fmt.Errorf("category ID is required")
	}

	// Check if category exists
	exists, err := s.repo.Category.ExistsByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to check category existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("category not found")
	}

	// Validate and set defaults
	if err := s.validateListProductsRequest(req); err != nil {
		return nil, err
	}

	// Set category ID
	req.CategoryID = categoryID

	// Get products
	products, total, err := s.repo.Product.ListByCategoryID(ctx, categoryID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by category: %w", err)
	}

	// Convert to response
	productResponses := make([]models.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = product.ToResponse()
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &models.ListProductsResponse{
		Products:   productResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *ProductService) ActivateProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product ID is required")
	}

	product, err := s.repo.Product.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if product.IsActive {
		return fmt.Errorf("product is already active")
	}

	product.IsActive = true
	if err := s.repo.Product.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to activate product: %w", err)
	}

	return nil
}

func (s *ProductService) DeactivateProduct(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("product ID is required")
	}

	product, err := s.repo.Product.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !product.IsActive {
		return fmt.Errorf("product is already inactive")
	}

	product.IsActive = false
	if err := s.repo.Product.Update(ctx, product); err != nil {
		return fmt.Errorf("failed to deactivate product: %w", err)
	}

	return nil
}

// Validation methods
func (s *ProductService) validateCreateProductRequest(req *models.CreateProductRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("product name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("product name must be less than 255 characters")
	}

	if req.Price <= 0 {
		return fmt.Errorf("product price must be greater than 0")
	}

	if req.Price > 999999.99 {
		return fmt.Errorf("product price is too high")
	}

	if strings.TrimSpace(req.CategoryID) == "" {
		return fmt.Errorf("category ID is required")
	}

	if len(req.Description) > 5000 {
		return fmt.Errorf("product description must be less than 5000 characters")
	}

	if req.ImageURL != "" && len(req.ImageURL) > 500 {
		return fmt.Errorf("image URL must be less than 500 characters")
	}

	return nil
}

func (s *ProductService) validateUpdateProductRequest(req *models.UpdateProductRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("product name is required")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("product name must be less than 255 characters")
	}

	if req.Price <= 0 {
		return fmt.Errorf("product price must be greater than 0")
	}

	if req.Price > 999999.99 {
		return fmt.Errorf("product price is too high")
	}

	if strings.TrimSpace(req.CategoryID) == "" {
		return fmt.Errorf("category ID is required")
	}

	if len(req.Description) > 5000 {
		return fmt.Errorf("product description must be less than 5000 characters")
	}

	if req.ImageURL != "" && len(req.ImageURL) > 500 {
		return fmt.Errorf("image URL must be less than 500 characters")
	}

	return nil
}

func (s *ProductService) validateListProductsRequest(req *models.ListProductsRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}

	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	if req.PageSize > 100 {
		req.PageSize = 100
	}

	return nil
}
