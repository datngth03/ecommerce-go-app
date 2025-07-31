# cmd/notification-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /notification-service ./cmd/notification-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /notification-service .
COPY .env . 
COPY migrations/notification_service ./migrations/notification_service/
COPY internal/shared/logger ./internal/shared/logger/ 
COPY internal/shared/tracing ./internal/shared/tracing/ 

EXPOSE 50058
EXPOSE 9108

ENTRYPOINT ["/app/notification-service"]
