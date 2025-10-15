package service

import (
	"context"
	"fmt"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/repository"
)

type CartService struct {
	cartRepo      repository.CartRepository
	productClient *client.ProductClient
}

func NewCartService(
	cartRepo repository.CartRepository,
	productClient *client.ProductClient,
) *CartService {
	return &CartService{
		cartRepo:      cartRepo,
		productClient: productClient,
	}
}

// GetCart retrieves user's cart
func (s *CartService) GetCart(ctx context.Context, userID int64) (*models.Cart, error) {
	return s.cartRepo.Get(ctx, userID)
}

// AddToCart adds item to cart
func (s *CartService) AddToCart(ctx context.Context, userID int64, productID string, quantity int32) (*models.Cart, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}

	// Validate product exists
	product, err := s.productClient.GetProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check stock
	hasStock, err := s.productClient.CheckStock(ctx, productID, quantity)
	if err != nil || !hasStock {
		return nil, fmt.Errorf("insufficient stock for product: %s", product.Name)
	}

	// Add to cart
	item := &models.CartItem{
		ProductID:   productID,
		ProductName: product.Name,
		Quantity:    quantity,
		Price:       product.Price,
	}

	return s.cartRepo.AddItem(ctx, userID, item)
}

// UpdateCartItem updates item quantity in cart
func (s *CartService) UpdateCartItem(ctx context.Context, userID int64, productID string, quantity int32) (*models.Cart, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}

	// Check stock
	hasStock, err := s.productClient.CheckStock(ctx, productID, quantity)
	if err != nil || !hasStock {
		return nil, fmt.Errorf("insufficient stock")
	}

	return s.cartRepo.UpdateItem(ctx, userID, productID, quantity)
}

// RemoveFromCart removes item from cart
func (s *CartService) RemoveFromCart(ctx context.Context, userID int64, productID string) (*models.Cart, error) {
	return s.cartRepo.RemoveItem(ctx, userID, productID)
}

// ClearCart removes all items from cart
func (s *CartService) ClearCart(ctx context.Context, userID int64) error {
	return s.cartRepo.Clear(ctx, userID)
}
