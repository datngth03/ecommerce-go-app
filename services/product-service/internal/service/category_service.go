package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/repository"
)

type CategoryService struct {
	repo *repository.Repository
}

func NewCategoryService(repo *repository.Repository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, req *models.CreateCategoryRequest) (*models.CategoryResponse, error) {
	// Validate input
	if err := s.validateCreateCategoryRequest(req); err != nil {
		return nil, err
	}

	// Check if category name already exists
	nameExists, err := s.repo.Category.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check category name: %w", err)
	}
	if nameExists {
		return nil, fmt.Errorf("category with name '%s' already exists", req.Name)
	}

	// Create category
	category := &models.Category{
		Name: strings.TrimSpace(req.Name),
	}

	if err := s.repo.Category.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Get created category
	createdCategory, err := s.repo.Category.GetByID(ctx, category.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created category: %w", err)
	}

	response := createdCategory.ToResponse()
	return &response, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, id string) (*models.CategoryResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("category ID is required")
	}

	category, err := s.repo.Category.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *CategoryService) GetCategoryBySlug(ctx context.Context, slug string) (*models.CategoryResponse, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, fmt.Errorf("category slug is required")
	}

	category, err := s.repo.Category.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id string, req *models.UpdateCategoryRequest) (*models.CategoryResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("category ID is required")
	}

	// Validate input
	if err := s.validateUpdateCategoryRequest(req); err != nil {
		return nil, err
	}

	// Get existing category
	existingCategory, err := s.repo.Category.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if category name already exists (excluding current category)
	if req.Name != existingCategory.Name {
		nameExists, err := s.repo.Category.ExistsByName(ctx, req.Name, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check category name: %w", err)
		}
		if nameExists {
			return nil, fmt.Errorf("category with name '%s' already exists", req.Name)
		}
	}

	// Update category
	existingCategory.Name = strings.TrimSpace(req.Name)

	if err := s.repo.Category.Update(ctx, existingCategory); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	// Get updated category
	updatedCategory, err := s.repo.Category.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated category: %w", err)
	}

	response := updatedCategory.ToResponse()
	return &response, nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("category ID is required")
	}

	// Check if category exists
	_, err := s.repo.Category.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if category has products
	productCount, err := s.repo.Product.CountByCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if productCount > 0 {
		return fmt.Errorf("cannot delete category: it contains %d products", productCount)
	}

	if err := s.repo.Category.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

func (s *CategoryService) ListCategories(ctx context.Context) (*models.ListCategoriesResponse, error) {
	categories, err := s.repo.Category.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// Convert to response
	categoryResponses := make([]models.CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = category.ToResponse()
	}

	return &models.ListCategoriesResponse{
		Categories: categoryResponses,
		Total:      int64(len(categoryResponses)),
	}, nil
}

func (s *CategoryService) GetCategoryWithProductCount(ctx context.Context, id string) (*models.CategoryResponse, int64, error) {
	if strings.TrimSpace(id) == "" {
		return nil, 0, fmt.Errorf("category ID is required")
	}

	// Get category
	category, err := s.repo.Category.GetByID(ctx, id)
	if err != nil {
		return nil, 0, err
	}

	// Get product count
	productCount, err := s.repo.Product.CountByCategory(ctx, id)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	response := category.ToResponse()
	return &response, productCount, nil
}

func (s *CategoryService) GetCategoriesWithProductCount(ctx context.Context) ([]models.CategoryResponse, map[string]int64, error) {
	// Get all categories
	categories, err := s.repo.Category.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list categories: %w", err)
	}

	// Get product count for each category
	productCounts := make(map[string]int64)
	categoryResponses := make([]models.CategoryResponse, len(categories))

	for i, category := range categories {
		categoryResponses[i] = category.ToResponse()

		count, err := s.repo.Product.CountByCategory(ctx, category.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to count products for category %s: %w", category.ID, err)
		}
		productCounts[category.ID] = count
	}

	return categoryResponses, productCounts, nil
}

func (s *CategoryService) ValidateCategoryExists(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("category ID is required")
	}

	exists, err := s.repo.Category.ExistsByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check category existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (s *CategoryService) CheckCategoryNameAvailability(ctx context.Context, name string, excludeID ...string) (bool, error) {
	if strings.TrimSpace(name) == "" {
		return false, fmt.Errorf("category name is required")
	}

	exists, err := s.repo.Category.ExistsByName(ctx, name, excludeID...)
	if err != nil {
		return false, fmt.Errorf("failed to check category name availability: %w", err)
	}

	return !exists, nil
}

// Validation methods
func (s *CategoryService) validateCreateCategoryRequest(req *models.CreateCategoryRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("category name is required")
	}

	if len(req.Name) < 1 {
		return fmt.Errorf("category name must be at least 1 character long")
	}

	if len(req.Name) > 100 {
		return fmt.Errorf("category name must be less than 100 characters")
	}

	// Check for invalid characters
	if strings.ContainsAny(req.Name, "<>&\"'") {
		return fmt.Errorf("category name contains invalid characters")
	}

	return nil
}

func (s *CategoryService) validateUpdateCategoryRequest(req *models.UpdateCategoryRequest) error {
	if req == nil {
		return fmt.Errorf("request is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("category name is required")
	}

	if len(req.Name) < 1 {
		return fmt.Errorf("category name must be at least 1 character long")
	}

	if len(req.Name) > 100 {
		return fmt.Errorf("category name must be less than 100 characters")
	}

	// Check for invalid characters
	if strings.ContainsAny(req.Name, "<>&\"'") {
		return fmt.Errorf("category name contains invalid characters")
	}

	return nil
}
