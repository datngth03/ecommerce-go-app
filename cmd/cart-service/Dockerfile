# cmd/cart-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /cart-service ./cmd/cart-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /cart-service .
COPY .env . 
COPY internal/shared/logger ./internal/shared/logger/ 
COPY internal/shared/tracing ./internal/shared/tracing/ 

EXPOSE 50054
EXPOSE 9104

ENTRYPOINT ["/app/cart-service"]
