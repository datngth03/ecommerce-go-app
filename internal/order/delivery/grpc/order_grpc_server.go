// internal/order/delivery/grpc/order_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/order/application"
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order" // Generated Order gRPC client
)

// OrderGRPCServer implements the order_client.OrderServiceServer interface.
type OrderGRPCServer struct {
	order_client.UnimplementedOrderServiceServer // Embedded to satisfy all methods
	orderService                                 application.OrderService
}

// NewOrderGRPCServer creates a new instance of OrderGRPCServer.
func NewOrderGRPCServer(svc application.OrderService) *OrderGRPCServer {
	return &OrderGRPCServer{
		orderService: svc,
	}
}

// CreateOrder implements the gRPC CreateOrder method.
func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *order_client.CreateOrderRequest) (*order_client.OrderResponse, error) {
	log.Printf("Received CreateOrder request for User ID: %s", req.GetUserId())
	resp, err := s.orderService.CreateOrder(ctx, req)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetOrderById implements the gRPC GetOrderById method.
func (s *OrderGRPCServer) GetOrderById(ctx context.Context, req *order_client.GetOrderByIdRequest) (*order_client.OrderResponse, error) {
	log.Printf("Received GetOrderById request for ID: %s", req.GetId())
	resp, err := s.orderService.GetOrderById(ctx, req)
	if err != nil {
		log.Printf("Error getting order by ID: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateOrderStatus implements the gRPC UpdateOrderStatus method.
func (s *OrderGRPCServer) UpdateOrderStatus(ctx context.Context, req *order_client.UpdateOrderStatusRequest) (*order_client.OrderResponse, error) {
	log.Printf("Received UpdateOrderStatus request for Order ID: %s, New Status: %s", req.GetOrderId(), req.GetNewStatus())
	resp, err := s.orderService.UpdateOrderStatus(ctx, req)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		return nil, err
	}
	return resp, nil
}

// CancelOrder implements the gRPC CancelOrder method.
func (s *OrderGRPCServer) CancelOrder(ctx context.Context, req *order_client.CancelOrderRequest) (*order_client.OrderResponse, error) {
	log.Printf("Received CancelOrder request for Order ID: %s", req.GetOrderId())
	resp, err := s.orderService.CancelOrder(ctx, req)
	if err != nil {
		log.Printf("Error canceling order: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListOrders implements the gRPC ListOrders method.
func (s *OrderGRPCServer) ListOrders(ctx context.Context, req *order_client.ListOrdersRequest) (*order_client.ListOrdersResponse, error) {
	log.Printf("Received ListOrders request (User ID: %s, Status: %s, Limit: %d, Offset: %d)", req.GetUserId(), req.GetStatus(), req.GetLimit(), req.GetOffset())
	resp, err := s.orderService.ListOrders(ctx, req)
	if err != nil {
		log.Printf("Error listing orders: %v", err)
		return nil, err
	}
	return resp, nil
}
