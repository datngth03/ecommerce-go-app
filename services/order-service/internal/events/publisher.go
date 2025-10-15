package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/datngth03/ecommerce-go-app/services/order-service/internal/models"
	"github.com/streadway/amqp"
)

const (
	ExchangeName = "ecommerce.orders"
	ExchangeType = "topic"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		ExchangeName,
		ExchangeType,
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("✅ Connected to RabbitMQ and declared exchange: %s", ExchangeName)

	return &Publisher{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// PublishOrderCreated publishes order created event
func (p *Publisher) PublishOrderCreated(ctx context.Context, order *models.Order) error {
	event := NewOrderCreatedEvent(order)
	return p.publish(ctx, EventOrderCreated, event)
}

// PublishOrderStatusChanged publishes order status changed event
func (p *Publisher) PublishOrderStatusChanged(ctx context.Context, order *models.Order) error {
	// Note: We don't have old status in current implementation
	// You might want to pass it as parameter or fetch from DB
	event := NewOrderStatusChangedEvent(order, "")
	return p.publish(ctx, EventOrderStatusChanged, event)
}

// PublishOrderCancelled publishes order cancelled event
func (p *Publisher) PublishOrderCancelled(ctx context.Context, order *models.Order) error {
	event := NewOrderCancelledEvent(order, "User cancelled")
	return p.publish(ctx, EventOrderCancelled, event)
}

// publish is the internal method to publish events
func (p *Publisher) publish(ctx context.Context, routingKey string, event interface{}) error {
	if p.channel == nil {
		return fmt.Errorf("publisher not initialized")
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.Publish(
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("📤 Published event: %s, size: %d bytes", routingKey, len(body))
	return nil
}

// HealthCheck checks if RabbitMQ connection is alive
func (p *Publisher) HealthCheck() error {
	if p.conn == nil || p.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}
	if p.channel == nil {
		return fmt.Errorf("channel is closed")
	}
	return nil
}
