# cmd/product-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /product-service ./cmd/product-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /product-service .
COPY .env .
COPY migrations/product_service ./migrations/product_service/
COPY internal/shared/events ./internal/shared/events/ 

EXPOSE 50052
EXPOSE 9102

ENTRYPOINT ["/app/product-service"]
