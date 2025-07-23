// internal/shipping/delivery/grpc/shipping_grpc_server.go
package grpc

import (
	"context"
	"log" // Temporarily using log for errors

	"github.com/datngth03/ecommerce-go-app/internal/shipping/application"
	shipping_client "github.com/datngth03/ecommerce-go-app/pkg/client/shipping" // Generated Shipping gRPC client
)

// ShippingGRPCServer implements the shipping_client.ShippingServiceServer interface.
type ShippingGRPCServer struct {
	shipping_client.UnimplementedShippingServiceServer // Embedded to satisfy all methods
	shippingService                                    application.ShippingService
}

// NewShippingGRPCServer creates a new instance of ShippingGRPCServer.
func NewShippingGRPCServer(svc application.ShippingService) *ShippingGRPCServer {
	return &ShippingGRPCServer{
		shippingService: svc,
	}
}

// CalculateShippingCost implements the gRPC CalculateShippingCost method.
func (s *ShippingGRPCServer) CalculateShippingCost(ctx context.Context, req *shipping_client.CalculateShippingCostRequest) (*shipping_client.CalculateShippingCostResponse, error) {
	log.Printf("Received CalculateShippingCost request for Order ID: %s", req.GetOrderId())
	resp, err := s.shippingService.CalculateShippingCost(ctx, req)
	if err != nil {
		log.Printf("Error calculating shipping cost: %v", err)
		return nil, err
	}
	return resp, nil
}

// CreateShipment implements the gRPC CreateShipment method.
func (s *ShippingGRPCServer) CreateShipment(ctx context.Context, req *shipping_client.CreateShipmentRequest) (*shipping_client.ShipmentResponse, error) {
	log.Printf("Received CreateShipment request for Order ID: %s, Carrier: %s", req.GetOrderId(), req.GetCarrier())
	resp, err := s.shippingService.CreateShipment(ctx, req)
	if err != nil {
		log.Printf("Error creating shipment: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetShipmentById implements the gRPC GetShipmentById method.
func (s *ShippingGRPCServer) GetShipmentById(ctx context.Context, req *shipping_client.GetShipmentByIdRequest) (*shipping_client.ShipmentResponse, error) {
	log.Printf("Received GetShipmentById request for ID: %s", req.GetId())
	resp, err := s.shippingService.GetShipmentById(ctx, req)
	if err != nil {
		log.Printf("Error getting shipment by ID: %v", err)
		return nil, err
	}
	return resp, nil
}

// UpdateShipmentStatus implements the gRPC UpdateShipmentStatus method.
func (s *ShippingGRPCServer) UpdateShipmentStatus(ctx context.Context, req *shipping_client.UpdateShipmentStatusRequest) (*shipping_client.ShipmentResponse, error) {
	log.Printf("Received UpdateShipmentStatus request for Shipment ID: %s, New Status: %s", req.GetShipmentId(), req.GetNewStatus())
	resp, err := s.shippingService.UpdateShipmentStatus(ctx, req)
	if err != nil {
		log.Printf("Error updating shipment status: %v", err)
		return nil, err
	}
	return resp, nil
}

// TrackShipment implements the gRPC TrackShipment method.
func (s *ShippingGRPCServer) TrackShipment(ctx context.Context, req *shipping_client.TrackShipmentRequest) (*shipping_client.ShipmentResponse, error) {
	log.Printf("Received TrackShipment request for Shipment ID: %s", req.GetShipmentId())
	resp, err := s.shippingService.TrackShipment(ctx, req)
	if err != nil {
		log.Printf("Error tracking shipment: %v", err)
		return nil, err
	}
	return resp, nil
}

// ListShipments implements the gRPC ListShipments method.
func (s *ShippingGRPCServer) ListShipments(ctx context.Context, req *shipping_client.ListShipmentsRequest) (*shipping_client.ListShipmentsResponse, error) {
	log.Printf("Received ListShipments request (User ID: %s, Order ID: %s, Status: %s)", req.GetUserId(), req.GetOrderId(), req.GetStatus())
	resp, err := s.shippingService.ListShipments(ctx, req)
	if err != nil {
		log.Printf("Error listing shipments: %v", err)
		return nil, err
	}
	return resp, nil
}
