package rpc

import (
	"context"
	"time"

	pb "github.com/datngth03/ecommerce-go-app/proto/order_service"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/metrics"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	orderService *service.OrderService
	cartService  *service.CartService
}

func NewOrderServer(orderService *service.OrderService, cartService *service.CartService) *OrderServer {
	return &OrderServer{
		orderService: orderService,
		cartService:  cartService,
	}
}

// CreateOrder creates a new order
func (s *OrderServer) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	start := time.Now()

	order, err := s.orderService.CreateOrder(ctx, req.UserId, req.ShippingAddress, req.PaymentMethod)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("CreateOrder", grpcStatus, time.Since(start))
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	metrics.RecordGRPCRequest("CreateOrder", grpcStatus, time.Since(start))
	metrics.RecordOrderCreated(order.Status)

	return &pb.CreateOrderResponse{
		Order: orderToProto(order),
	}, nil
}

// GetOrder retrieves an order by ID
func (s *OrderServer) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	start := time.Now()

	// Extract user ID from context (from auth middleware)
	userID := getUserIDFromContext(ctx)

	order, err := s.orderService.GetOrder(ctx, req.Id, userID)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("GetOrder", grpcStatus, time.Since(start))
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	metrics.RecordGRPCRequest("GetOrder", grpcStatus, time.Since(start))

	return &pb.GetOrderResponse{
		Order: orderToProto(order),
	}, nil
}

// ListOrders retrieves user's orders
func (s *OrderServer) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	start := time.Now()

	orders, total, err := s.orderService.ListOrders(ctx, req.UserId, req.Page, req.PageSize, req.Status)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("ListOrders", grpcStatus, time.Since(start))
		return nil, status.Errorf(codes.Internal, "failed to list orders: %v", err)
	}

	metrics.RecordGRPCRequest("ListOrders", grpcStatus, time.Since(start))

	pbOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = orderToProto(order)
	}

	return &pb.ListOrdersResponse{
		Orders:     pbOrders,
		TotalCount: total,
	}, nil
}

// UpdateOrderStatus updates order status
func (s *OrderServer) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (*pb.UpdateOrderStatusResponse, error) {
	start := time.Now()

	userID := getUserIDFromContext(ctx)

	order, err := s.orderService.UpdateOrderStatus(ctx, req.Id, req.Status, userID)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("UpdateOrderStatus", grpcStatus, time.Since(start))
		return nil, status.Errorf(codes.Internal, "failed to update order status: %v", err)
	}

	metrics.RecordGRPCRequest("UpdateOrderStatus", grpcStatus, time.Since(start))

	return &pb.UpdateOrderStatusResponse{
		Order: orderToProto(order),
	}, nil
}

// CancelOrder cancels an order
func (s *OrderServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*emptypb.Empty, error) {
	start := time.Now()

	err := s.orderService.CancelOrder(ctx, req.Id, req.UserId)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("CancelOrder", grpcStatus, time.Since(start))
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}

	metrics.RecordGRPCRequest("CancelOrder", grpcStatus, time.Since(start))

	return &emptypb.Empty{}, nil
}

// AddToCart adds item to cart
func (s *OrderServer) AddToCart(ctx context.Context, req *pb.AddToCartRequest) (*pb.CartResponse, error) {
	start := time.Now()

	cart, err := s.cartService.AddToCart(ctx, req.UserId, req.ProductId, req.Quantity)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("AddToCart", grpcStatus, time.Since(start))
		metrics.RecordCartOperation("add", grpcStatus)
		return nil, status.Errorf(codes.Internal, "failed to add to cart: %v", err)
	}

	metrics.RecordGRPCRequest("AddToCart", grpcStatus, time.Since(start))
	metrics.RecordCartOperation("add", grpcStatus)

	return &pb.CartResponse{
		Cart: cartToProto(cart),
	}, nil
}

// GetCart retrieves user's cart
func (s *OrderServer) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.CartResponse, error) {
	start := time.Now()

	cart, err := s.cartService.GetCart(ctx, req.UserId)

	grpcStatus := "success"
	if err != nil {
		grpcStatus = "error"
		metrics.RecordGRPCRequest("GetCart", grpcStatus, time.Since(start))
		return nil, status.Errorf(codes.Internal, "failed to get cart: %v", err)
	}

	metrics.RecordGRPCRequest("GetCart", grpcStatus, time.Since(start))

	return &pb.CartResponse{
		Cart: cartToProto(cart),
	}, nil
}

// UpdateCartItem updates cart item quantity
func (s *OrderServer) UpdateCartItem(ctx context.Context, req *pb.UpdateCartItemRequest) (*pb.CartResponse, error) {
	cart, err := s.cartService.UpdateCartItem(ctx, req.UserId, req.ProductId, req.Quantity)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update cart item: %v", err)
	}

	return &pb.CartResponse{
		Cart: cartToProto(cart),
	}, nil
}

// RemoveFromCart removes item from cart
func (s *OrderServer) RemoveFromCart(ctx context.Context, req *pb.RemoveFromCartRequest) (*pb.CartResponse, error) {
	cart, err := s.cartService.RemoveFromCart(ctx, req.UserId, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove from cart: %v", err)
	}

	return &pb.CartResponse{
		Cart: cartToProto(cart),
	}, nil
}

// ClearCart clears all items from cart
func (s *OrderServer) ClearCart(ctx context.Context, req *pb.ClearCartRequest) (*emptypb.Empty, error) {
	err := s.cartService.ClearCart(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to clear cart: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// Helper functions

func orderToProto(order *models.Order) *pb.Order {
	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &pb.OrderItem{
			Id:          item.ID,
			OrderId:     item.OrderID,
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    item.Subtotal,
		}
	}

	return &pb.Order{
		Id:              order.ID,
		UserId:          order.UserID,
		Status:          order.Status,
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		PaymentMethod:   order.PaymentMethod,
		Items:           items,
		CreatedAt:       timestamppb.New(order.CreatedAt),
		UpdatedAt:       timestamppb.New(order.UpdatedAt),
	}
}

func cartToProto(cart *models.Cart) *pb.Cart {
	items := make([]*pb.CartItem, len(cart.Items))
	var totalAmount float64

	for i, item := range cart.Items {
		subtotal := float64(item.Quantity) * item.Price
		totalAmount += subtotal

		items[i] = &pb.CartItem{
			ProductId:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    subtotal,
		}
	}

	return &pb.Cart{
		UserId:      cart.UserID,
		Items:       items,
		TotalAmount: totalAmount,
		UpdatedAt:   timestamppb.New(cart.UpdatedAt),
	}
}

func getUserIDFromContext(ctx context.Context) int64 {
	// Extract user ID from context metadata (set by auth middleware)
	// For now, return 0 - should be implemented based on your auth strategy
	return 0
}
