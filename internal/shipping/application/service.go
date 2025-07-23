// internal/shipping/application/service.go
package application

import (
	"context"
	"errors"
	"fmt"
	"log" // Temporarily using log for errors
	"time"

	"github.com/datngth03/ecommerce-go-app/internal/shipping/domain"
	order_client "github.com/datngth03/ecommerce-go-app/pkg/client/order"       // For calling Order Service
	shipping_client "github.com/datngth03/ecommerce-go-app/pkg/client/shipping" // Generated Shipping gRPC client
	"github.com/google/uuid"                                                    // For generating UUIDs
)

// ShippingService defines the application service interface for shipping-related operations.
type ShippingService interface {
	CalculateShippingCost(ctx context.Context, req *shipping_client.CalculateShippingCostRequest) (*shipping_client.CalculateShippingCostResponse, error)
	CreateShipment(ctx context.Context, req *shipping_client.CreateShipmentRequest) (*shipping_client.ShipmentResponse, error)
	GetShipmentById(ctx context.Context, req *shipping_client.GetShipmentByIdRequest) (*shipping_client.ShipmentResponse, error)
	UpdateShipmentStatus(ctx context.Context, req *shipping_client.UpdateShipmentStatusRequest) (*shipping_client.ShipmentResponse, error)
	TrackShipment(ctx context.Context, req *shipping_client.TrackShipmentRequest) (*shipping_client.ShipmentResponse, error)
	ListShipments(ctx context.Context, req *shipping_client.ListShipmentsRequest) (*shipping_client.ListShipmentsResponse, error)
}

// shippingService implements the ShippingService interface.
type shippingService struct {
	shipmentRepo domain.ShipmentRepository
	orderClient  order_client.OrderServiceClient // Client to call Order Service
	// TODO: Add other dependencies like third-party carrier API clients, event publisher
}

// NewShippingService creates a new instance of ShippingService.
func NewShippingService(shipmentRepo domain.ShipmentRepository, orderClient order_client.OrderServiceClient) ShippingService {
	return &shippingService{
		shipmentRepo: shipmentRepo,
		orderClient:  orderClient,
	}
}

// CalculateShippingCost calculates the estimated shipping cost for an order.
func (s *shippingService) CalculateShippingCost(ctx context.Context, req *shipping_client.CalculateShippingCostRequest) (*shipping_client.CalculateShippingCostResponse, error) {
	if req.GetOrderId() == "" || req.GetShippingAddress() == "" || req.GetOriginAddress() == "" {
		return nil, errors.New("order ID, shipping address, and origin address are required")
	}

	// Fetch order details from Order Service to get items, total amount, etc.
	orderResp, err := s.orderClient.GetOrderById(ctx, &order_client.GetOrderByIdRequest{Id: req.GetOrderId()})
	if err != nil {
		return nil, fmt.Errorf("failed to get order details for shipping cost calculation: %w", err)
	}
	if orderResp.GetOrder() == nil {
		return nil, fmt.Errorf("order with ID %s not found for shipping cost calculation", req.GetOrderId())
	}

	// TODO: Implement actual shipping cost calculation logic.
	// This would typically involve:
	// 1. Calculating total weight/dimensions from order items (need product details from Product Service, possibly via Order Service).
	// 2. Calling a third-party shipping carrier API (e.g., FedEx, UPS, USPS) with origin, destination, weight, dimensions.
	// 3. Applying business rules (e.g., free shipping over a certain amount, flat rates).

	// For now, a dummy calculation based on order total amount.
	var calculatedCost float64
	if orderResp.GetOrder().GetTotalAmount() > 100.0 {
		calculatedCost = 0.0 // Free shipping for orders over $100
	} else {
		calculatedCost = 5.99 // Flat rate
	}

	return &shipping_client.CalculateShippingCostResponse{
		Cost:                  calculatedCost,
		Currency:              "USD",
		EstimatedDeliveryTime: "3-5 business days",
	}, nil
}

// CreateShipment creates a new shipment record and potentially interacts with a carrier.
func (s *shippingService) CreateShipment(ctx context.Context, req *shipping_client.CreateShipmentRequest) (*shipping_client.ShipmentResponse, error) {
	if req.GetOrderId() == "" || req.GetUserId() == "" || req.GetShippingAddress() == "" || req.GetCarrier() == "" {
		return nil, errors.New("order ID, user ID, shipping address, and carrier are required")
	}

	// Validate order existence with Order Service
	orderResp, err := s.orderClient.GetOrderById(ctx, &order_client.GetOrderByIdRequest{Id: req.GetOrderId()})
	if err != nil {
		return nil, fmt.Errorf("failed to get order details for shipment creation: %w", err)
	}
	if orderResp.GetOrder() == nil {
		return nil, fmt.Errorf("order with ID %s not found for shipment creation", req.GetOrderId())
	}
	if orderResp.GetOrder().GetStatus() != "paid" && orderResp.GetOrder().GetStatus() != "confirmed" {
		return nil, fmt.Errorf("order %s is not in a shippable status", req.GetOrderId())
	}

	shipmentID := uuid.New().String()
	shipment := domain.NewShipment(
		shipmentID,
		req.GetOrderId(),
		req.GetUserId(),
		req.GetShippingCost(),
		req.GetShippingAddress(),
		req.GetCarrier(),
	)

	// TODO: Integrate with a real shipping carrier API here (e.g., FedEx API to generate tracking number)
	// For now, generate a dummy tracking number.
	shipment.SetTrackingNumber(fmt.Sprintf("TRK-%s-%d", shipmentID[:8], time.Now().UnixNano()))
	shipment.UpdateStatus("in_transit") // Set initial status to in_transit

	if err := s.shipmentRepo.Save(ctx, shipment); err != nil {
		return nil, fmt.Errorf("failed to save shipment: %w", err)
	}

	// Update order status in Order Service
	_, err = s.orderClient.UpdateOrderStatus(ctx, &order_client.UpdateOrderStatusRequest{
		OrderId:   shipment.OrderID,
		NewStatus: "shipped",
	})
	if err != nil {
		log.Printf("Warning: Failed to update order status to 'shipped' for order %s after shipment %s: %v", shipment.OrderID, shipment.ID, err)
		// Decide how to handle this: retry, manual intervention, dead letter queue
	}
	// TODO: Publish ShipmentCreated event

	return &shipping_client.ShipmentResponse{
		Shipment: &shipping_client.Shipment{
			Id:              shipment.ID,
			OrderId:         shipment.OrderID,
			UserId:          shipment.UserID,
			ShippingCost:    shipment.ShippingCost,
			TrackingNumber:  shipment.TrackingNumber,
			Carrier:         shipment.Carrier,
			Status:          shipment.Status,
			ShippingAddress: shipment.ShippingAddress,
			CreatedAt:       shipment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       shipment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetShipmentById retrieves shipment details by ID.
func (s *shippingService) GetShipmentById(ctx context.Context, req *shipping_client.GetShipmentByIdRequest) (*shipping_client.ShipmentResponse, error) {
	if req.GetId() == "" {
		return nil, errors.New("shipment ID is required")
	}

	shipment, err := s.shipmentRepo.FindByID(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, errors.New("shipment not found")) {
			return nil, errors.New("shipment not found")
		}
		return nil, fmt.Errorf("failed to retrieve shipment: %w", err)
	}

	return &shipping_client.ShipmentResponse{
		Shipment: &shipping_client.Shipment{
			Id:              shipment.ID,
			OrderId:         shipment.OrderID,
			UserId:          shipment.UserID,
			ShippingCost:    shipment.ShippingCost,
			TrackingNumber:  shipment.TrackingNumber,
			Carrier:         shipment.Carrier,
			Status:          shipment.Status,
			ShippingAddress: shipment.ShippingAddress,
			CreatedAt:       shipment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       shipment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateShipmentStatus handles updating a shipment's status.
func (s *shippingService) UpdateShipmentStatus(ctx context.Context, req *shipping_client.UpdateShipmentStatusRequest) (*shipping_client.ShipmentResponse, error) {
	if req.GetShipmentId() == "" || req.GetNewStatus() == "" {
		return nil, errors.New("shipment ID and new status are required")
	}

	shipment, err := s.shipmentRepo.FindByID(ctx, req.GetShipmentId())
	if err != nil {
		if errors.Is(err, errors.New("shipment not found")) {
			return nil, errors.New("shipment not found")
		}
		return nil, fmt.Errorf("failed to find shipment for status update: %w", err)
	}

	// Basic status transition validation
	// TODO: Add more robust status validation logic (e.g., cannot go from delivered to pending)

	shipment.UpdateStatus(req.GetNewStatus())
	if req.GetTrackingNumber() != "" {
		shipment.SetTrackingNumber(req.GetTrackingNumber())
	}

	if err := s.shipmentRepo.Save(ctx, shipment); err != nil {
		return nil, fmt.Errorf("failed to update shipment status: %w", err)
	}

	// TODO: Publish ShipmentStatusUpdated event (for Order Service to update its status if needed)

	return &shipping_client.ShipmentResponse{
		Shipment: &shipping_client.Shipment{
			Id:              shipment.ID,
			OrderId:         shipment.OrderID,
			UserId:          shipment.UserID,
			ShippingCost:    shipment.ShippingCost,
			TrackingNumber:  shipment.TrackingNumber,
			Carrier:         shipment.Carrier,
			Status:          shipment.Status,
			ShippingAddress: shipment.ShippingAddress,
			CreatedAt:       shipment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       shipment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// TrackShipment tracks a shipment's status (currently just retrieves details).
func (s *shippingService) TrackShipment(ctx context.Context, req *shipping_client.TrackShipmentRequest) (*shipping_client.ShipmentResponse, error) {
	if req.GetShipmentId() == "" {
		return nil, errors.New("shipment ID is required for tracking")
	}

	shipment, err := s.shipmentRepo.FindByID(ctx, req.GetShipmentId())
	if err != nil {
		if errors.Is(err, errors.New("shipment not found")) {
			return nil, errors.New("shipment not found")
		}
		return nil, fmt.Errorf("failed to retrieve shipment for tracking: %w", err)
	}

	// TODO: In a real scenario, this would involve calling a third-party carrier's tracking API
	// using the shipment.TrackingNumber and shipment.Carrier.
	// The response would then update the shipment status or provide detailed tracking info.

	return &shipping_client.ShipmentResponse{
		Shipment: &shipping_client.Shipment{
			Id:              shipment.ID,
			OrderId:         shipment.OrderID,
			UserId:          shipment.UserID,
			ShippingCost:    shipment.ShippingCost,
			TrackingNumber:  shipment.TrackingNumber,
			Carrier:         shipment.Carrier,
			Status:          shipment.Status,
			ShippingAddress: shipment.ShippingAddress,
			CreatedAt:       shipment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       shipment.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// ListShipments handles listing shipments with pagination and filters.
func (s *shippingService) ListShipments(ctx context.Context, req *shipping_client.ListShipmentsRequest) (*shipping_client.ListShipmentsResponse, error) {
	shipments, totalCount, err := s.shipmentRepo.FindAll(ctx, req.GetUserId(), req.GetOrderId(), req.GetStatus(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list shipments: %w", err)
	}

	shipmentResponses := make([]*shipping_client.Shipment, len(shipments))
	for i, s := range shipments {
		shipmentResponses[i] = &shipping_client.Shipment{
			Id:              s.ID,
			OrderId:         s.OrderID,
			UserId:          s.UserID,
			ShippingCost:    s.ShippingCost,
			TrackingNumber:  s.TrackingNumber,
			Carrier:         s.Carrier,
			Status:          s.Status,
			ShippingAddress: s.ShippingAddress,
			CreatedAt:       s.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       s.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &shipping_client.ListShipmentsResponse{
		Shipments:  shipmentResponses,
		TotalCount: totalCount,
	}, nil
}
