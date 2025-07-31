# cmd/search-service/Dockerfile

# Stage 1: Build the application
# Sử dụng Go 1.22 Alpine làm môi trường build để tạo binary nhỏ gọn
FROM golang:1.22-alpine AS builder

# Đặt thư mục làm việc bên trong container
WORKDIR /app

# Sao chép go.mod và go.sum trước để tận dụng Docker cache cho các dependencies
COPY go.mod .
COPY go.sum .

# Tải các dependencies của Go modules
RUN go mod download

# Sao chép phần còn lại của mã nguồn ứng dụng
COPY . .

# Build ứng dụng Search Service
# CGO_ENABLED=0 để tạo binary tĩnh, hữu ích cho image nền Alpine
# GOOS=linux để đảm bảo binary chạy trên Linux (môi trường Docker)
# -a -installsuffix cgo để buộc rebuild và tránh các phụ thuộc cgo
# -ldflags "-s -w" để loại bỏ thông tin debug và bảng ký hiệu, giúp binary nhỏ hơn
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags "-s -w" -o /search-service ./cmd/search-service/main.go

# Stage 2: Create the final lean image
# Sử dụng Alpine Linux làm image nền cuối cùng để có kích thước nhỏ nhất
FROM alpine:latest

# Đặt múi giờ (tùy chọn, nhưng là thực hành tốt)
ENV TZ=Asia/Ho_Chi_Minh
RUN apk add --no-cache tzdata

# Đặt thư mục làm việc bên trong image cuối cùng
WORKDIR /app

# Sao chép binary đã build từ stage builder
COPY --from=builder /search-service .

# Sao chép file .env (cho mục đích kiểm thử cục bộ trong Docker nếu cần,
# nhưng trong production thường truyền qua biến môi trường)
COPY .env .

# Sao chép các gói dùng chung cần thiết cho runtime
# Đảm bảo các đường dẫn này khớp với cấu trúc dự án của bạn
COPY internal/shared/logger ./internal/shared/logger/
COPY internal/shared/tracing ./internal/shared/tracing/
COPY internal/shared/events ./internal/shared/events/

# Expose các cổng mà dịch vụ sẽ lắng nghe
# 50061 là cổng gRPC của Search Service
# 9101 là cổng metrics cho Prometheus
EXPOSE 50061
EXPOSE 9111

# Lệnh để chạy executable khi container khởi động
ENTRYPOINT ["/app/search-service"]
