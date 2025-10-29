package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor that adds tracing
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata from incoming context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Extract trace context from metadata
		ctx = otel.GetTextMapPropagator().Extract(ctx, &metadataCarrier{md: md})

		// Start new span
		tracer := GetTracer()
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Add gRPC attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
			attribute.String("rpc.method", info.FullMethod),
		)

		// Call the handler
		resp, err := handler(ctx, req)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			// Add gRPC status code
			if st, ok := status.FromError(err); ok {
				span.SetAttributes(
					attribute.Int("rpc.grpc.status_code", int(st.Code())),
				)
			}
		} else {
			span.SetStatus(codes.Ok, "success")
			span.SetAttributes(
				attribute.Int("rpc.grpc.status_code", 0),
			)
		}

		return resp, err
	}
}

// UnaryClientInterceptor returns a grpc.UnaryClientInterceptor that adds tracing
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Start new span
		tracer := GetTracer()
		ctx, span := tracer.Start(ctx, method,
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()

		// Add gRPC attributes
		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", method),
			attribute.String("rpc.method", method),
			attribute.String("net.peer.name", cc.Target()),
		)

		// Inject trace context into metadata
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		carrier := &metadataCarrier{md: md}
		otel.GetTextMapPropagator().Inject(ctx, carrier)
		ctx = metadata.NewOutgoingContext(ctx, carrier.md)

		// Call the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			// Add gRPC status code
			if st, ok := status.FromError(err); ok {
				span.SetAttributes(
					attribute.Int("rpc.grpc.status_code", int(st.Code())),
				)
			}
		} else {
			span.SetStatus(codes.Ok, "success")
			span.SetAttributes(
				attribute.Int("rpc.grpc.status_code", 0),
			)
		}

		return err
	}
}

// metadataCarrier implements TextMapCarrier for gRPC metadata
type metadataCarrier struct {
	md metadata.MD
}

// Get retrieves a value from the metadata
func (mc *metadataCarrier) Get(key string) string {
	values := mc.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Set sets a value in the metadata
func (mc *metadataCarrier) Set(key, value string) {
	mc.md.Set(key, value)
}

// Keys returns all keys in the metadata
func (mc *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(mc.md))
	for k := range mc.md {
		keys = append(keys, k)
	}
	return keys
}
