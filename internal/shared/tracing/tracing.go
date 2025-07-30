// internal/shared/tracing/tracing.go
package tracing

import (
	"context"
	"fmt"

	// "time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"      // Alias sdk/trace as sdktrace
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0" // Using latest semconv version
	otrace "go.opentelemetry.io/otel/trace"            // ADDED: Import go.opentelemetry.io/otel/trace and alias as otrace
)

// InitTracerProvider initializes the OpenTelemetry TracerProvider with Jaeger exporter.
func InitTracerProvider(ctx context.Context, serviceName, jaegerCollectorURL string) (*sdktrace.TracerProvider, error) { // Using sdktrace.TracerProvider
	// Create the Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerCollectorURL)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create a new TracerProvider with the Jaeger exporter
	tp := sdktrace.NewTracerProvider( // Using sdktrace.NewTracerProvider
		// Always be sure to batch in production.
		sdktrace.WithBatcher(exporter), // Using sdktrace.WithBatcher
		// Record information about this application in a Resource.
		sdktrace.WithResource(resource.NewWithAttributes( // Using sdktrace.WithResource
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("environment", "development"), // Or production
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

// NewSpan creates a new span for manual tracing.
// You can use this function in business logic to create child spans.
func NewSpan(ctx context.Context, spanName string, serviceName string) (context.Context, otrace.Span) {
	return otel.Tracer(serviceName).Start(ctx, spanName)
}
