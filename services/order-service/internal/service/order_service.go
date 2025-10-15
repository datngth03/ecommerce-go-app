package service

import (
	"context"
	"fmt"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/client"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/events"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/repository"
)

type OrderService struct {
	orderRepo      repository.OrderRepository
	cartRepo       repository.CartRepository
	productClient  *client.ProductClient
	userClient     *client.UserClient
	eventPublisher *events.Publisher
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	cartRepo repository.CartRepository,
	productClient *client.ProductClient,
	userClient *client.UserClient,
	eventPublisher *events.Publisher,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		cartRepo:       cartRepo,
		productClient:  productClient,
		userClient:     userClient,
		eventPublisher: eventPublisher,
	}
}

// CreateOrder creates a new order from cart or direct items
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, shippingAddress, paymentMethod string) (*models.Order, error) {
	// Validate user
	if _, err := s.userClient.ValidateUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("invalid user: %w", err)
	}

	// Get cart items
	cart, err := s.cartRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if len(cart.Items) == 0 {
		return nil, fmt.Errorf("cart is empty")
	}

	// Validate products and stock
	var totalAmount float64
	orderItems := make([]models.OrderItem, 0, len(cart.Items))

	for _, cartItem := range cart.Items {
		// Get product details
		product, err := s.productClient.GetProduct(ctx, cartItem.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %s not found: %w", cartItem.ProductID, err)
		}

		// Check stock
		hasStock, err := s.productClient.CheckStock(ctx, cartItem.ProductID, cartItem.Quantity)
		if err != nil || !hasStock {
			return nil, fmt.Errorf("insufficient stock for product %s", product.Name)
		}

		// Create order item
		subtotal := float64(cartItem.Quantity) * cartItem.Price
		orderItems = append(orderItems, models.OrderItem{
			ProductID:   cartItem.ProductID,
			ProductName: product.Name,
			Quantity:    cartItem.Quantity,
			Price:       cartItem.Price,
			Subtotal:    subtotal,
		})

		totalAmount += subtotal
	}

	// Create order
	order := &models.Order{
		UserID:          userID,
		Status:          models.OrderStatusPending,
		TotalAmount:     totalAmount,
		ShippingAddress: shippingAddress,
		PaymentMethod:   paymentMethod,
		Items:           orderItems,
	}

	createdOrder, err := s.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Clear cart after successful order
	s.cartRepo.Clear(ctx, userID)

	// Publish order created event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishOrderCreated(ctx, createdOrder)
	}

	return createdOrder, nil
}

// GetOrder retrieves order by ID
func (s *OrderService) GetOrder(ctx context.Context, orderID string, userID int64) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Verify order belongs to user
	if order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}

// ListOrders retrieves user's orders with pagination
func (s *OrderService) ListOrders(ctx context.Context, userID int64, page, pageSize int32, status string) ([]*models.Order, int64, error) {
	return s.orderRepo.List(ctx, userID, page, pageSize, status)
}

// UpdateOrderStatus updates order status
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID, status string, userID int64) (*models.Order, error) {
	// Validate status
	validStatuses := map[string]bool{
		models.OrderStatusPending:    true,
		models.OrderStatusConfirmed:  true,
		models.OrderStatusProcessing: true,
		models.OrderStatusShipped:    true,
		models.OrderStatusDelivered:  true,
		models.OrderStatusCancelled:  true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid order status: %s", status)
	}

	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Verify order belongs to user (or allow admin to update any order)
	if order.UserID != userID {
		return nil, fmt.Errorf("order not found")
	}

	// Update status
	updatedOrder, err := s.orderRepo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		return nil, err
	}

	// Publish status change event
	if s.eventPublisher != nil {
		s.eventPublisher.PublishOrderStatusChanged(ctx, updatedOrder)
	}

	return updatedOrder, nil
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, orderID string, userID int64) error {
	// Get order
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Can only cancel pending orders
	if order.Status != models.OrderStatusPending {
		return fmt.Errorf("cannot cancel order with status: %s", order.Status)
	}

	// Cancel order
	if err := s.orderRepo.Cancel(ctx, orderID, userID); err != nil {
		return err
	}

	// Publish order cancelled event
	if s.eventPublisher != nil {
		order.Status = models.OrderStatusCancelled
		s.eventPublisher.PublishOrderCancelled(ctx, order)
	}

	return nil
}
