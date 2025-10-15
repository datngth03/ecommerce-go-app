package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/datngth03/ecommerce-go-app/services/inventory-service/internal/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

// EventSubscriber handles inventory-related events
type EventSubscriber struct {
	service *service.InventoryService
	conn    *amqp.Connection
	channel *amqp.Channel
}

// OrderCreatedEvent represents an order creation event
type OrderCreatedEvent struct {
	OrderID string `json:"order_id"`
	Items   []struct {
		ProductID string `json:"product_id"`
		Quantity  int32  `json:"quantity"`
	} `json:"items"`
}

// OrderCancelledEvent represents an order cancellation event
type OrderCancelledEvent struct {
	OrderID string `json:"order_id"`
	Reason  string `json:"reason"`
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(svc *service.InventoryService, rabbitmqURL string) (*EventSubscriber, error) {
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &EventSubscriber{
		service: svc,
		conn:    conn,
		channel: channel,
	}, nil
}

// Start starts listening to events
func (s *EventSubscriber) Start(ctx context.Context) error {
	// Declare exchange
	err := s.channel.ExchangeDeclare(
		"orders",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	queue, err := s.channel.QueueDeclare(
		"inventory.orders",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind to order.created
	err = s.channel.QueueBind(
		queue.Name,
		"order.created",
		"orders",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind order.created: %w", err)
	}

	// Bind to order.cancelled
	err = s.channel.QueueBind(
		queue.Name,
		"order.cancelled",
		"orders",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind order.cancelled: %w", err)
	}

	// Start consuming
	msgs, err := s.channel.Consume(
		queue.Name,
		"inventory-service",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Println("Inventory event subscriber started")

	// Process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping inventory event subscriber")
				return
			case msg := <-msgs:
				s.handleMessage(ctx, msg)
			}
		}
	}()

	return nil
}

// handleMessage processes incoming messages
func (s *EventSubscriber) handleMessage(ctx context.Context, msg amqp.Delivery) {
	log.Printf("Received event: %s", msg.RoutingKey)

	switch msg.RoutingKey {
	case "order.created":
		s.handleOrderCreated(ctx, msg)
	case "order.cancelled":
		s.handleOrderCancelled(ctx, msg)
	default:
		log.Printf("Unknown routing key: %s", msg.RoutingKey)
		msg.Ack(false)
	}
}

// handleOrderCreated handles order creation event
func (s *EventSubscriber) handleOrderCreated(ctx context.Context, msg amqp.Delivery) {
	var event OrderCreatedEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		log.Printf("Failed to unmarshal order.created event: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("Reserving stock for order: %s", event.OrderID)

	// Convert items format
	items := make([]struct {
		ProductID string
		Quantity  int32
	}, len(event.Items))

	for i, item := range event.Items {
		items[i] = struct {
			ProductID string
			Quantity  int32
		}{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// Reserve stock
	_, err = s.service.ReserveStock(ctx, event.OrderID, items)
	if err != nil {
		log.Printf("Failed to reserve stock for order %s: %v", event.OrderID, err)
		msg.Nack(false, true) // requeue
		return
	}

	log.Printf("Stock reserved successfully for order: %s", event.OrderID)
	msg.Ack(false)
}

// handleOrderCancelled handles order cancellation event
func (s *EventSubscriber) handleOrderCancelled(ctx context.Context, msg amqp.Delivery) {
	var event OrderCancelledEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		log.Printf("Failed to unmarshal order.cancelled event: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("Releasing stock for cancelled order: %s", event.OrderID)

	// Release reserved stock
	err = s.service.ReleaseStock(ctx, event.OrderID, event.Reason)
	if err != nil {
		log.Printf("Failed to release stock for order %s: %v", event.OrderID, err)
		msg.Nack(false, true) // requeue
		return
	}

	log.Printf("Stock released successfully for order: %s", event.OrderID)
	msg.Ack(false)
}

// Close closes the connection
func (s *EventSubscriber) Close() error {
	if s.channel != nil {
		s.channel.Close()
	}
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
