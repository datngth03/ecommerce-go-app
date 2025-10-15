package repository

import (
	"context"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	GetByID(ctx context.Context, id string) (*models.Order, error)
	List(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*models.Order, int64, error)
	UpdateStatus(ctx context.Context, id, status string) (*models.Order, error)
	Cancel(ctx context.Context, id string, userID int64) error
}

type CartRepository interface {
	Get(ctx context.Context, userID int64) (*models.Cart, error)
	AddItem(ctx context.Context, userID int64, item *models.CartItem) (*models.Cart, error)
	UpdateItem(ctx context.Context, userID int64, productID string, quantity int32) (*models.Cart, error)
	RemoveItem(ctx context.Context, userID int64, productID string) (*models.Cart, error)
	Clear(ctx context.Context, userID int64) error
}
