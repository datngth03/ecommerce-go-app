package repository

import (
	"context"

	"github.com/ecommerce/services/user-service/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, id int64, user *models.UpdateUserRequest) (*models.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, limit int, role string) ([]models.User, int, error)
	EmailExists(ctx context.Context, email string, excludeID int64) (bool, error)
}
