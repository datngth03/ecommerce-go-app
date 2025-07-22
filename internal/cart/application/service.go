// internal/cart/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/cart/domain"
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart"       // Generated Cart gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // Product gRPC client for product details
)

// CartService defines the application service interface for cart-related operations.
// CartService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến giỏ hàng.
type CartService interface {
	AddItemToCart(ctx context.Context, req *cart_client.AddItemToCartRequest) (*cart_client.CartResponse, error)
	UpdateCartItemQuantity(ctx context.Context, req *cart_client.UpdateCartItemQuantityRequest) (*cart_client.CartResponse, error)
	RemoveItemFromCart(ctx context.Context, req *cart_client.RemoveItemFromCartRequest) (*cart_client.CartResponse, error)
	GetCart(ctx context.Context, req *cart_client.GetCartRequest) (*cart_client.CartResponse, error)
	ClearCart(ctx context.Context, req *cart_client.ClearCartRequest) (*cart_client.CartResponse, error)
}

// cartService implements the CartService interface.
// cartService triển khai interface CartService.
type cartService struct {
	cartRepo      domain.CartRepository
	productClient product_client.ProductServiceClient // Client to call Product Service
	// TODO: Add other dependencies like user client (to validate user existence)
	// Thêm các dependency khác như client người dùng (để xác thực sự tồn tại của người dùng)
}

// NewCartService creates a new instance of CartService.
// NewCartService tạo một thể hiện mới của CartService.
func NewCartService(cartRepo domain.CartRepository, productClient product_client.ProductServiceClient) CartService {
	return &cartService{
		cartRepo:      cartRepo,
		productClient: productClient,
	}
}

// AddItemToCart adds a product to the user's cart or updates its quantity.
// AddItemToCart thêm một sản phẩm vào giỏ hàng của người dùng hoặc cập nhật số lượng của nó.
func (s *cartService) AddItemToCart(ctx context.Context, req *cart_client.AddItemToCartRequest) (*cart_client.CartResponse, error) {
	if req.GetUserId() == "" || req.GetProductId() == "" || req.GetQuantity() <= 0 {
		return nil, errors.New("user ID, product ID, and positive quantity are required")
	}

	// Get product details from Product Service to ensure valid product and get correct price/name
	productResp, err := s.productClient.GetProductById(ctx, &product_client.GetProductByIdRequest{Id: req.GetProductId()})
	if err != nil {
		return nil, fmt.Errorf("failed to get product details for ID %s: %w", req.GetProductId(), err)
	}
	if productResp.GetProduct() == nil {
		return nil, fmt.Errorf("product with ID %s not found", req.GetProductId())
	}

	// Use product details from Product Service, not from request directly, for consistency
	productName := productResp.GetProduct().GetName()
	price := productResp.GetProduct().GetPrice()

	cart, err := s.cartRepo.GetCart(ctx, req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}
	if cart == nil {
		cart = domain.NewCart(req.GetUserId())
	}

	cart.AddItem(req.GetProductId(), productName, price, req.GetQuantity())

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	return &cart_client.CartResponse{Cart: mapDomainCartToProto(cart)}, nil
}

// UpdateCartItemQuantity updates the quantity of a specific item in the cart.
// UpdateCartItemQuantity cập nhật số lượng của một mặt hàng cụ thể trong giỏ hàng.
func (s *cartService) UpdateCartItemQuantity(ctx context.Context, req *cart_client.UpdateCartItemQuantityRequest) (*cart_client.CartResponse, error) {
	if req.GetUserId() == "" || req.GetProductId() == "" || req.GetQuantity() < 0 {
		return nil, errors.New("user ID, product ID, and non-negative quantity are required")
	}

	cart, err := s.cartRepo.GetCart(ctx, req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}
	if cart == nil {
		return nil, errors.New("cart not found")
	}

	if updated := cart.UpdateItemQuantity(req.GetProductId(), req.GetQuantity()); !updated {
		return nil, errors.New("item not found in cart")
	}

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	return &cart_client.CartResponse{Cart: mapDomainCartToProto(cart)}, nil
}

// RemoveItemFromCart removes a product from the cart.
// RemoveItemFromCart xóa một sản phẩm khỏi giỏ hàng.
func (s *cartService) RemoveItemFromCart(ctx context.Context, req *cart_client.RemoveItemFromCartRequest) (*cart_client.CartResponse, error) {
	if req.GetUserId() == "" || req.GetProductId() == "" {
		return nil, errors.New("user ID and product ID are required")
	}

	cart, err := s.cartRepo.GetCart(ctx, req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}
	if cart == nil {
		return nil, errors.New("cart not found")
	}

	if removed := cart.RemoveItem(req.GetProductId()); !removed {
		return nil, errors.New("item not found in cart")
	}

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to save cart: %w", err)
	}

	return &cart_client.CartResponse{Cart: mapDomainCartToProto(cart)}, nil
}

// GetCart retrieves the current state of the user's cart.
// GetCart lấy trạng thái hiện tại của giỏ hàng của người dùng.
func (s *cartService) GetCart(ctx context.Context, req *cart_client.GetCartRequest) (*cart_client.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, errors.New("user ID is required")
	}

	cart, err := s.cartRepo.GetCart(ctx, req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}
	if cart == nil {
		// Return an empty cart if not found, rather than an error
		return &cart_client.CartResponse{Cart: mapDomainCartToProto(domain.NewCart(req.GetUserId()))}, nil
	}

	return &cart_client.CartResponse{Cart: mapDomainCartToProto(cart)}, nil
}

// ClearCart clears all items from the user's cart.
// ClearCart xóa tất cả các mặt hàng khỏi giỏ hàng của người dùng.
func (s *cartService) ClearCart(ctx context.Context, req *cart_client.ClearCartRequest) (*cart_client.CartResponse, error) {
	if req.GetUserId() == "" {
		return nil, errors.New("user ID is required")
	}

	cart, err := s.cartRepo.GetCart(ctx, req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cart: %w", err)
	}
	if cart == nil {
		// Cart already empty or not found, consider it cleared
		return &cart_client.CartResponse{Cart: mapDomainCartToProto(domain.NewCart(req.GetUserId()))}, nil
	}

	cart.Clear()

	if err := s.cartRepo.SaveCart(ctx, cart); err != nil {
		return nil, fmt.Errorf("failed to clear cart: %w", err)
	}

	return &cart_client.CartResponse{Cart: mapDomainCartToProto(cart)}, nil
}

// Helper function to map domain.Cart to cart_client.Cart (protobuf)
// Hàm trợ giúp để ánh xạ domain.Cart sang cart_client.Cart (protobuf)
func mapDomainCartToProto(cart *domain.Cart) *cart_client.Cart {
	protoItems := make([]*cart_client.CartItem, len(cart.Items))
	for i, item := range cart.Items {
		protoItems[i] = &cart_client.CartItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}

	return &cart_client.Cart{
		UserId:        cart.UserID,
		Items:         protoItems,
		TotalAmount:   cart.TotalAmount,
		LastUpdatedAt: cart.LastUpdatedAt.Format(time.RFC3339),
	}
}
