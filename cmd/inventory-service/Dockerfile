# cmd/inventory-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /inventory-service ./cmd/inventory-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /inventory-service .
COPY .env . 
COPY migrations/inventory_service ./migrations/inventory_service/
COPY internal/shared/logger ./internal/shared/logger/
COPY internal/shared/tracing ./internal/shared/tracing/ 
COPY internal/shared/events ./internal/shared/events/ 

EXPOSE 50059
EXPOSE 9109

ENTRYPOINT ["/app/inventory-service"]
