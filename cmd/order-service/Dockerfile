# cmd/order-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /order-service ./cmd/order-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /order-service .
COPY .env .
COPY migrations/order_service ./migrations/order_service/
COPY internal/shared/logger ./internal/shared/logger/ 
COPY internal/shared/tracing ./internal/shared/tracing/ 

EXPOSE 50053
EXPOSE 9103

ENTRYPOINT ["/app/order-service"]
