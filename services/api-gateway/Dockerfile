# =============================================================================
# Stage 1: Build Stage
# =============================================================================
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy proto and shared modules
COPY proto/ /build/proto/
COPY shared/ /build/shared/

# Copy service directory
COPY services/api-gateway/ /build/services/api-gateway/

# Set working directory to service
WORKDIR /build/services/api-gateway

# Download dependencies
RUN go mod download

# Build the application
# CGO_ENABLED=0: Build static binary without C dependencies
# -ldflags="-w -s": Strip debug information to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o api-gateway \
    ./cmd/main.go

# =============================================================================
# Stage 2: Runtime Stage
# =============================================================================
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/services/api-gateway/api-gateway .

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
# 8000: HTTP API Gateway port
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Set environment variables
ENV GIN_MODE=release \
    TZ=Asia/Ho_Chi_Minh

# Run the application
CMD ["./api-gateway"]
