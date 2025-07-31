# cmd/api-gateway/Dockerfile

# Stage 1: Build the application
# Use Go 1.22 Alpine as the build environment for a compact binary
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache for dependencies
COPY go.mod .
COPY go.sum .

# Download Go module dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the API Gateway application
# CGO_ENABLED=0 for a static binary, suitable for Alpine base image
# GOOS=linux ensures the binary runs on Linux (Docker environment)
# -a -installsuffix cgo forces rebuild and avoids cgo dependencies
# -ldflags "-s -w" strips debug information and symbol table for a smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /api-gateway ./cmd/api-gateway/main.go

# Stage 2: Create the final lean image
# Use Alpine Linux as the final base image for the smallest possible size
FROM alpine:latest

# Set timezone (optional, but good practice for consistent timestamps)
ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

# Set the working directory inside the final image
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /api-gateway .

# Copy the .env file (for local testing within Docker if needed,
# but in production, environment variables are typically passed directly)
COPY .env .

# Copy necessary shared packages for runtime
# Ensure these paths match your project structure
COPY internal/shared/logger ./internal/shared/logger/
COPY internal/shared/tracing ./internal/shared/tracing/
COPY internal/api_gateway/delivery/http ./internal/api_gateway/delivery/http/ 
COPY internal/api_gateway/middleware ./internal/api_gateway/middleware/
COPY api/protobufs ./api/protobufs/ 

# Expose the ports the service will listen on
# 8080 is the HTTP port for the API Gateway
# 9080 is the metrics port for Prometheus
EXPOSE 8080
EXPOSE 9080

# Command to run the executable when the container starts
ENTRYPOINT ["/app/api-gateway"]
