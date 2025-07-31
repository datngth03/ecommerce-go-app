# cmd/review-service/Dockerfile

# Stage 1: Build the application
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /review-service ./cmd/review-service/main.go

# Stage 2: Create the final lean image
FROM alpine:latest

ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /review-service .
COPY .env .
COPY migrations/review_service ./migrations/review_service/
COPY internal/shared/logger ./internal/shared/logger/ 
COPY internal/shared/tracing ./internal/shared/tracing/

EXPOSE 50060
EXPOSE 9110

ENTRYPOINT ["/app/review-service"]
