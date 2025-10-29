# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy proto and shared modules
COPY proto/ /app/proto/
COPY shared/ /app/shared/

# Copy service directory
COPY services/inventory-service/ /app/services/inventory-service/

# Set working directory to service
WORKDIR /app/services/inventory-service

# Download dependencies
RUN go mod download
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Create app directory
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/services/inventory-service/main .

# Copy migration files
COPY --from=builder /app/services/inventory-service/migrations ./migrations/

# Expose ports (HTTP and gRPC)
EXPOSE 8005 9005

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8005/health || exit 1

# Run the application
CMD ["./main"]
