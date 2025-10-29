package tracing

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// GinMiddleware returns a Gin middleware that adds tracing to HTTP requests
func GinMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from HTTP headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start new span
		tracer := GetTracer()
		ctx, span := tracer.Start(ctx, c.Request.URL.Path,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Store context in Gin context
		c.Request = c.Request.WithContext(ctx)

		// Add HTTP attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.target", c.Request.URL.Path),
			attribute.String("http.host", c.Request.Host),
			attribute.String("http.scheme", c.Request.URL.Scheme),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
			attribute.String("service.name", serviceName),
		)

		// Process request
		c.Next()

		// Add response attributes
		statusCode := c.Writer.Status()
		span.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, "HTTP error")
			if len(c.Errors) > 0 {
				span.RecordError(c.Errors.Last())
			}
		} else {
			span.SetStatus(codes.Ok, "success")
		}

		// Add route if available
		if c.FullPath() != "" {
			span.SetAttributes(attribute.String("http.route", c.FullPath()))
		}
	}
}

// HTTPHeaderCarrier implements TextMapCarrier for HTTP headers
type HTTPHeaderCarrier map[string][]string

// Get retrieves a value from HTTP headers
func (hc HTTPHeaderCarrier) Get(key string) string {
	values := hc[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets a value in HTTP headers
func (hc HTTPHeaderCarrier) Set(key, value string) {
	hc[key] = []string{value}
}

// Keys returns all keys in HTTP headers
func (hc HTTPHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}
	return keys
}
