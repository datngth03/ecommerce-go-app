# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy proto and shared modules
COPY proto/ /app/proto/
COPY shared/ /app/shared/

# Copy service directory
COPY services/product-service/ /app/services/product-service/

# Set working directory to service
WORKDIR /app/services/product-service

# Download dependencies
RUN go mod download
# RUN go mod tidy
RUN go mod tidy

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o product-service ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/services/product-service/product-service .

# Copy migrations
COPY --from=builder /app/services/product-service/migrations ./migrations

EXPOSE 8081 9091

CMD ["./product-service"]
