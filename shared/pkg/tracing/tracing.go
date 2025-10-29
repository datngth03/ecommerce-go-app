package tracing

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerConfig holds configuration for tracing
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string // e.g., "jaeger:4317" for OTLP gRPC
	Enabled        bool
}

var (
	globalTracer trace.Tracer
)

// InitTracer initializes the OpenTelemetry tracer with Jaeger exporter via OTLP
func InitTracer(cfg TracerConfig) (func(context.Context) error, error) {
	if !cfg.Enabled {
		log.Println("⚠️  Tracing is disabled")
		// Return no-op cleanup function
		return func(ctx context.Context) error { return nil }, nil
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create gRPC connection to Jaeger collector (OTLP endpoint)
	conn, err := grpc.Dial(cfg.JaegerEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Create OTLP trace exporter
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithGRPCConn(conn),
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces (use ParentBased in production)
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator to W3C Trace Context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get tracer instance
	globalTracer = tp.Tracer(cfg.ServiceName)

	log.Printf("✓ Distributed tracing initialized: service=%s, endpoint=%s", cfg.ServiceName, cfg.JaegerEndpoint)

	// Return cleanup function
	cleanup := func(ctx context.Context) error {
		log.Println("Shutting down tracer...")
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
			return fmt.Errorf("failed to shutdown tracer: %w", err)
		}
		if err := conn.Close(); err != nil {
			log.Printf("Error closing gRPC connection: %v", err)
			return fmt.Errorf("failed to close gRPC connection: %w", err)
		}
		return nil
	}

	return cleanup, nil
}

// GetTracer returns the global tracer instance
func GetTracer() trace.Tracer {
	if globalTracer == nil {
		// Return no-op tracer if not initialized
		return otel.Tracer("noop")
	}
	return globalTracer
}

// StartSpan creates a new span with the given name
func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return GetTracer().Start(ctx, spanName, opts...)
}

// AddSpanTags adds tags/attributes to the current span
func AddSpanTags(span trace.Span, tags map[string]string) {
	if span == nil {
		return
	}
	attrs := make([]attribute.KeyValue, 0, len(tags))
	for k, v := range tags {
		attrs = append(attrs, attribute.String(k, v))
	}
	span.SetAttributes(attrs...)
}

// AddSpanError adds error information to the span
func AddSpanError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetAttributes(attribute.Bool("error", true))
}

// AddSpanEvent adds an event to the span
func AddSpanEvent(span trace.Span, name string, attributes ...attribute.KeyValue) {
	if span == nil {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attributes...))
}
