// internal/repository/interfaces.go
package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/user-service/internal/models"
)

// UserRepositoryInterface defines the contract for user repository
type UserRepositoryInterface interface {
	// CRUD operations
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, updateData *models.UserUpdateData) (*models.User, error)
	Delete(ctx context.Context, id int64) error

	// Password operations
	UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error

	// Additional utility methods
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
