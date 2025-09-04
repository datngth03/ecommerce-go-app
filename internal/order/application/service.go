// internal/order/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/order/domain"
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"     // Generated Order gRPC client
	product_client "github.com/datngth03/ecommerce-go-app/pkg/client/product" // For calling Product Service
	"github.com/google/uuid"                                                  // For generating UUIDs
)

// OrderService defines the application service interface for order-related operations.
// OrderService định nghĩa interface dịch vụ ứng dụng cho các thao tác liên quan đến đơn hàng.
type OrderService interface {
	CreateOrder(ctx context.Context, req *order_client.CreateOrderRequest) (*order_client.OrderResponse, error)
	GetOrderById(ctx context.Context, req *order_client.GetOrderByIdRequest) (*order_client.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, req *order_client.UpdateOrderStatusRequest) (*order_client.OrderResponse, error)
	CancelOrder(ctx context.Context, req *order_client.CancelOrderRequest) (*order_client.OrderResponse, error)
	ListOrders(ctx context.Context, req *order_client.ListOrdersRequest) (*order_client.ListOrdersResponse, error)
}

// orderService implements the OrderService interface.
// orderService triển khai interface OrderService.
type orderService struct {
	orderRepo     domain.OrderRepository
	productClient product_client.ProductServiceClient // Client to call Product Service
	// TODO: Add other dependencies like user client, payment client, shipping client, event publisher
	// Thêm các dependency khác như client người dùng, client thanh toán, client vận chuyển, trình phát sự kiện
}

// NewOrderService creates a new instance of OrderService.
// NewOrderService tạo một thể hiện mới của OrderService.
func NewOrderService(orderRepo domain.OrderRepository, productClient product_client.ProductServiceClient) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		productClient: productClient,
	}
}

// CreateOrder handles the creation of a new order.
// CreateOrder xử lý việc tạo một đơn hàng mới.
func (s *orderService) CreateOrder(ctx context.Context, req *order_client.CreateOrderRequest) (*order_client.OrderResponse, error) {
	if req.GetUserId() == "" || len(req.GetItems()) == 0 || req.GetShippingAddress() == "" {
		return nil, errors.New("user ID, items, and shipping address are required")
	}

	var orderItems []domain.OrderItem
	var totalAmount float64

	// Validate products and calculate total amount
	// for _, itemReq := range req.GetItems() {
	// 	// Call Product Service to get product details and validate price/existence
	// 	productResp, err := s.productClient.GetProduct(ctx, &product_client.GetProductRequest{Id: itemReq.GetProductId()})
	// 	if err != nil {
	// 		return nil, fmt.Errorf("failed to get product details for ID %s: %w", itemReq.GetProductId(), err)
	// 	}
	// 	if productResp.GetProduct() == nil {
	// 		return nil, fmt.Errorf("product with ID %s not found", itemReq.GetProductId())
	// 	}
	// 	if productResp.GetProduct().GetPrice() != itemReq.GetPrice() {
	// 		return nil, fmt.Errorf("price mismatch for product %s. Expected %.2f, got %.2f",
	// 			itemReq.GetProductId(), productResp.GetProduct().GetPrice(), itemReq.GetPrice())
	// 	}
	// 	if itemReq.GetQuantity() <= 0 {
	// 		return nil, fmt.Errorf("quantity for product %s must be positive", itemReq.GetProductId())
	// 	}

	// 	orderItems = append(orderItems, domain.OrderItem{
	// 		ProductID:   itemReq.GetProductId(),
	// 		ProductName: productResp.GetProduct().GetName(),
	// 		Price:       itemReq.GetPrice(),
	// 		Quantity:    itemReq.GetQuantity(),
	// 	})
	// 	totalAmount += itemReq.GetPrice() * float64(itemReq.GetQuantity())
	// }

	orderID := uuid.New().String()
	order := domain.NewOrder(orderID, req.GetUserId(), orderItems, totalAmount, req.GetShippingAddress())

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	// TODO: Publish OrderCreated event (for Payment, Inventory, Shipping services)

	return &order_client.OrderResponse{
		Order: &order_client.Order{
			Id:              order.ID,
			UserId:          order.UserID,
			Items:           mapOrderItemsToProto(order.Items),
			TotalAmount:     order.TotalAmount,
			Status:          order.Status,
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       order.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetOrderById handles retrieving order details.
// GetOrderById xử lý việc lấy chi tiết đơn hàng.
func (s *orderService) GetOrderById(ctx context.Context, req *order_client.GetOrderByIdRequest) (*order_client.OrderResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("order ID is required")
	}

	order, err := s.orderRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, errors.New("order not found")) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}

	return &order_client.OrderResponse{
		Order: &order_client.Order{
			Id:              order.ID,
			UserId:          order.UserID,
			Items:           mapOrderItemsToProto(order.Items),
			TotalAmount:     order.TotalAmount,
			Status:          order.Status,
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       order.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateOrderStatus handles updating an order's status.
// UpdateOrderStatus xử lý việc cập nhật trạng thái của đơn hàng.
func (s *orderService) UpdateOrderStatus(ctx context.Context, req *order_client.UpdateOrderStatusRequest) (*order_client.OrderResponse, error) {
	if req.GetOrderId() == "" || req.GetNewStatus() == "" {
		return nil, errors.New("order ID and new status are required")
	}

	order, err := s.orderRepo.FindByID(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, errors.New("order not found")) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to find order for status update: %w", err)
	}

	// Basic status transition validation (can be more complex)
	if order.Status == "cancelled" || order.Status == "shipped" {
		return nil, errors.New("cannot change status of a cancelled or already shipped order")
	}
	// TODO: Add more robust status validation logic

	order.UpdateStatus(req.GetNewStatus())

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// TODO: Publish OrderStatusUpdated event

	return &order_client.OrderResponse{
		Order: &order_client.Order{
			Id:              order.ID,
			UserId:          order.UserID,
			Items:           mapOrderItemsToProto(order.Items),
			TotalAmount:     order.TotalAmount,
			Status:          order.Status,
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       order.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// CancelOrder handles cancelling an order.
// CancelOrder xử lý việc hủy một đơn hàng.
func (s *orderService) CancelOrder(ctx context.Context, req *order_client.CancelOrderRequest) (*order_client.OrderResponse, error) {
	if req.GetOrderId() == "" {
		return nil, errors.New("order ID is required for cancellation")
	}

	order, err := s.orderRepo.FindByID(ctx, req.GetOrderId())
	if err != nil {
		if errors.Is(err, errors.New("order not found")) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to find order for cancellation: %w", err)
	}

	if order.Status == "cancelled" || order.Status == "shipped" {
		return nil, errors.New("order cannot be cancelled if already cancelled or shipped")
	}

	order.UpdateStatus("cancelled") // Set status to cancelled

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// TODO: Publish OrderCancelled event (for Inventory to restock, Payment to refund)

	return &order_client.OrderResponse{
		Order: &order_client.Order{
			Id:              order.ID,
			UserId:          order.UserID,
			Items:           mapOrderItemsToProto(order.Items),
			TotalAmount:     order.TotalAmount,
			Status:          order.Status,
			ShippingAddress: order.ShippingAddress,
			CreatedAt:       order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       order.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ListOrders handles listing orders with pagination and filters.
// ListOrders xử lý việc liệt kê đơn hàng với phân trang và bộ lọc.
func (s *orderService) ListOrders(ctx context.Context, req *order_client.ListOrdersRequest) (*order_client.ListOrdersResponse, error) {
	orders, totalCount, err := s.orderRepo.FindAll(ctx, req.GetUserId(), req.GetStatus(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}

	orderResponses := make([]*order_client.Order, len(orders))
	for i, o := range orders {
		orderResponses[i] = &order_client.Order{
			Id:              o.ID,
			UserId:          o.UserID,
			Items:           mapOrderItemsToProto(o.Items),
			TotalAmount:     o.TotalAmount,
			Status:          o.Status,
			ShippingAddress: o.ShippingAddress,
			CreatedAt:       o.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       o.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &order_client.ListOrdersResponse{
		Orders:     orderResponses,
		TotalCount: totalCount,
	}, nil
}

// Helper function to map domain OrderItem to protobuf OrderItem
// Hàm trợ giúp để ánh xạ OrderItem từ domain sang protobuf
func mapOrderItemsToProto(items []domain.OrderItem) []*order_client.OrderItem {
	protoItems := make([]*order_client.OrderItem, len(items))
	for i, item := range items {
		protoItems[i] = &order_client.OrderItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
		}
	}
	return protoItems
}
