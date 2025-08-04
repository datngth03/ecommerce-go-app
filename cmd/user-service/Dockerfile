# cmd/user-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary, useful for Alpine base image
# -a for forcing rebuild of packages
# -installsuffix cgo for avoiding cgo dependencies
# -ldflags "-s -w" to strip debug information and symbol table for smaller binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /user-service ./cmd/user-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

# Set timezone (optional, but good practice)
ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /user-service .

# Copy the .env file (for local testing within Docker if needed, but usually passed via env vars)
# Be careful with sensitive data in .env in production images
COPY .env .

# Copy migrations directory
COPY migrations/user_service ./migrations/user_service/

# Expose the gRPC port and metrics port
EXPOSE 50051
EXPOSE 9101

# Command to run the executable
ENTRYPOINT ["/app/user-service"]
