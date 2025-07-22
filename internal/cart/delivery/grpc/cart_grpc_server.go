// internal/cart/delivery/grpc/cart_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/cart/application"
	cart_client "github.com/datngth03/ecommerce-go-app/pkg/client/cart" // Generated Cart gRPC client
)

// CartGRPCServer implements the cart_client.CartServiceServer interface.
type CartGRPCServer struct {
	cart_client.UnimplementedCartServiceServer // Embedded to satisfy all methods
	cartService                                application.CartService
}

// NewCartGRPCServer creates a new instance of CartGRPCServer.
func NewCartGRPCServer(svc application.CartService) *CartGRPCServer {
	return &CartGRPCServer{
		cartService: svc,
	}
}

// AddItemToCart implements the gRPC AddItemToCart method.
func (s *CartGRPCServer) AddItemToCart(ctx context.Context, req *cart_client.AddItemToCartRequest) (*cart_client.CartResponse, error) {
	log.Printf("Received AddItemToCart request for User ID: %s, Product ID: %s, Quantity: %d", req.GetUserId(), req.GetProductId(), req.GetQuantity())
	resp, err := s.cartService.AddItemToCart(ctx, req)
	if err != nil {
		log.Printf("Error adding item to cart: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateCartItemQuantity implements the gRPC UpdateCartItemQuantity method.
func (s *CartGRPCServer) UpdateCartItemQuantity(ctx context.Context, req *cart_client.UpdateCartItemQuantityRequest) (*cart_client.CartResponse, error) {
	log.Printf("Received UpdateCartItemQuantity request for User ID: %s, Product ID: %s, New Quantity: %d", req.GetUserId(), req.GetProductId(), req.GetQuantity())
	resp, err := s.cartService.UpdateCartItemQuantity(ctx, req)
	if err != nil {
		log.Printf("Error updating cart item quantity: %v", err)
		return nil, err
	}
	return resp, nil
}

// RemoveItemFromCart implements the gRPC RemoveItemFromCart method.
func (s *CartGRPCServer) RemoveItemFromCart(ctx context.Context, req *cart_client.RemoveItemFromCartRequest) (*cart_client.CartResponse, error) {
	log.Printf("Received RemoveItemFromCart request for User ID: %s, Product ID: %s", req.GetUserId(), req.GetProductId())
	resp, err := s.cartService.RemoveItemFromCart(ctx, req)
	if err != nil {
		log.Printf("Error removing item from cart: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetCart implements the gRPC GetCart method.
func (s *CartGRPCServer) GetCart(ctx context.Context, req *cart_client.GetCartRequest) (*cart_client.CartResponse, error) {
	log.Printf("Received GetCart request for User ID: %s", req.GetUserId())
	resp, err := s.cartService.GetCart(ctx, req)
	if err != nil {
		log.Printf("Error getting cart: %v", err)
		return nil, err
	}
	return resp, nil
}

// ClearCart implements the gRPC ClearCart method.
func (s *CartGRPCServer) ClearCart(ctx context.Context, req *cart_client.ClearCartRequest) (*cart_client.CartResponse, error) {
	log.Printf("Received ClearCart request for User ID: %s", req.GetUserId())
	resp, err := s.cartService.ClearCart(ctx, req)
	if err != nil {
		log.Printf("Error clearing cart: %v", err)
		return nil, err
	}
	return resp, nil
}
