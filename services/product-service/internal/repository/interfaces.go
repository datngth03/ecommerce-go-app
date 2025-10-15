// services/product-service/internal/repository/interfaces.go

package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/product-service/internal/models"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	GetBySlug(ctx context.Context, slug string) (*models.Product, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *models.ListProductsRequest) ([]models.Product, int64, error)
	ListByCategoryID(ctx context.Context, categoryID string, req *models.ListProductsRequest) ([]models.Product, int64, error)
	ExistsByName(ctx context.Context, name string, excludeID ...string) (bool, error)
	CountByCategory(ctx context.Context, categoryID string) (int64, error)
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id string) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]models.Category, error)
	ExistsByName(ctx context.Context, name string, excludeID ...string) (bool, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
}

// Repository aggregates all repository interfaces
type Repository struct {
	Product  ProductRepository
	Category CategoryRepository
}

// RepositoryOptions contains options for repository initialization
type RepositoryOptions struct {
	Database interface{} // Can be *sql.DB, *gorm.DB, etc.
}

// NewRepository creates a new repository instance
type NewRepositoryFunc func(opts *RepositoryOptions) (*Repository, error)
