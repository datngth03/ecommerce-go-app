# Docker & Makefile Configuration Guide

## 📋 Tổng Quan

Dự án E-commerce Go App đã được cấu hình lại với **architecture nhất quán** cho tất cả microservices:

### 🎯 Port Mapping (Đã Chuẩn Hóa)

| Service | HTTP Port | gRPC Port | Container |
|---------|-----------|-----------|-----------|
| **API Gateway** | 8000 | - | ecommerce-api-gateway |
| **User Service** | 8001 | 9001 | ecommerce-user-service |
| **Product Service** | 8002 | 9002 | ecommerce-product-service |
| **Order Service** | 8003 | 9003 | ecommerce-order-service |
| **Notification Service** | 8004 | 9004 | ecommerce-notification-service |
| **Inventory Service** | 8005 | 9005 | ecommerce-inventory-service |
| **Payment Service** | 8006 | 9006 | ecommerce-payment-service |

### 🗄️ Infrastructure Services

| Service | Port | UI/Management |
|---------|------|---------------|
| **PostgreSQL** | 5432 | - |
| **Redis** | 6379 | - |
| **RabbitMQ** | 5672 | Management: 15672 |
| **Prometheus** | 9090 | Web UI: 9090 |
| **Grafana** | 3000 | Web UI: 3000 (admin/admin123) |
| **Jaeger** | 16686 | Web UI: 16686 |
| **Nginx** | 80, 443 | - |

### 📊 Database Configuration

Mỗi service có database riêng:
- `users_db` - User Service
- `products_db` - Product Service
- `orders_db` - Order Service
- `payments_db` - Payment Service
- `inventory_db` - Inventory Service
- `notifications_db` - Notification Service

PostgreSQL init script tự động tạo tất cả databases khi khởi động.

---

## 🚀 Makefile Commands

### 📦 Setup & Installation

```bash
# Cài đặt dependencies và tools
make setup

# Build tất cả services thành .exe
make build-all
```

### 🔧 Development

```bash
# Start tất cả (infrastructure + services) với Docker
make dev

# Xem logs từ tất cả containers
make dev-logs

# Start chỉ microservices (giả sử infrastructure đã chạy)
make dev-services

# Run services locally (không dùng Docker)
make run-local

# Restart tất cả services
make restart
```

### 🗄️ Database Management

```bash
# Chạy migrations cho tất cả services
make migrate-up

# Rollback migrations (1 version)
make migrate-down

# Force migration version cho 1 service cụ thể
make migrate-force SERVICE=user VERSION=1

# Reset tất cả databases (XÓA TOÀN BỘ DỮ LIỆU!)
make db-reset
```

### 🔨 Protobuf Generation

```bash
# Generate proto files cho tất cả services
make proto-gen

# Xóa generated proto files
make proto-clean
```

### 🐳 Docker Operations

```bash
# Build Docker images cho tất cả services
make docker-build

# Build image cho 1 service cụ thể
make docker-build-service SERVICE=user-service

# Push images to registry
make docker-push

# Pull images from registry
make docker-pull

# Xem logs
make docker-logs

# Xem running containers
make docker-ps

# Stop containers
make stop

# Stop + remove containers, networks, volumes
make stop-all
```

### 🧹 Cleanup

```bash
# Xóa containers, volumes, binaries, coverage files
make clean
```

### 🧪 Testing

```bash
# Run tests
make test

# Run tests với coverage report
make test-cover

# Format code
make fmt

# Run linter
make lint

# Run tất cả checks (fmt + lint + test)
make check
```

### ℹ️ Help

```bash
# Xem tất cả available commands
make help
```

---

## 🔧 Environment Variables

### ✅ Đã Chuẩn Hóa (Tất Cả Services)

Tất cả services sử dụng **cùng convention** cho environment variables:

#### Service Info
```bash
SERVICE_NAME=user-service
SERVICE_VERSION=1.0.0
ENVIRONMENT=production
```

#### Server Config
```bash
HTTP_PORT=8001
GRPC_PORT=9001
```

#### Database
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=users_db
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

#### Redis (Tùy Service)
```bash
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0  # Mỗi service dùng DB number khác nhau (0-4)
REDIS_CACHE_TTL=300  # Inventory Service only
```

#### RabbitMQ (Order, Payment, Inventory, Notification)
```bash
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=admin123
RABBITMQ_VHOST=/
```

#### Logging
```bash
LOG_LEVEL=info
LOG_FORMAT=json
```

### 🔐 Service-Specific Config

#### API Gateway
```bash
JWT_SECRET=your-super-secret-jwt-key-change-in-production-2025
JWT_EXPIRATION_HOURS=24
REFRESH_TOKEN_EXPIRATION_DAYS=7
ENABLE_AUTH=true
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MIN=100
RATE_LIMIT_BURST_SIZE=50
CORS_ALLOWED_ORIGINS=http://localhost:3000
```

#### Notification Service
```bash
# Email/SMTP
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_ADDRESS=noreply@ecommerce.com
SMTP_FROM_NAME=E-Commerce Platform

# SMS/Twilio (Optional)
TWILIO_ACCOUNT_SID=your-twilio-account-sid
TWILIO_AUTH_TOKEN=your-twilio-auth-token
TWILIO_FROM_NUMBER=+1234567890
```

#### Payment Service
```bash
STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret
PAYPAL_CLIENT_ID=your_paypal_client_id
PAYPAL_SECRET=your_paypal_secret
```

#### Order Service (gRPC Addresses)
```bash
USER_SERVICE_GRPC_ADDR=user-service:9001
PRODUCT_SERVICE_GRPC_ADDR=product-service:9002
INVENTORY_SERVICE_GRPC_ADDR=inventory-service:9005
PAYMENT_SERVICE_GRPC_ADDR=payment-service:9006
```

---

## 📝 Key Changes Summary

### ✅ Đã Sửa

1. **Port Mapping**
   - ❌ Cũ: Ports bị lộn xộn (8081, 8082, 8084, 8085, 8086)
   - ✅ Mới: Ports theo thứ tự logic (8001-8006) + gRPC ports (9001-9006)

2. **Environment Variables**
   - ❌ Cũ: Mỗi service dùng tên biến khác nhau (`PORT`, `SERVER_PORT`, v.v.)
   - ✅ Mới: Tất cả dùng `HTTP_PORT` và `GRPC_PORT`

3. **Redis DB Numbers**
   - ✅ Mỗi service dùng Redis DB number riêng (0-4) để tránh conflict

4. **Health Checks**
   - ❌ Cũ: Dùng `curl` (không có trong alpine images)
   - ✅ Mới: Dùng `wget` (có sẵn trong alpine)
   - ✅ Thêm `start_period=40s` để service có thời gian khởi động

5. **Database Init**
   - ✅ Tạo file `init.sql` để tự động tạo 6 databases + extensions

6. **Makefile**
   - ✅ Thêm commands: `build-all`, `dev-logs`, `migrate-force`, `db-reset`
   - ✅ Fix proto generation path
   - ✅ Tách biệt `SERVICES` và `GATEWAY`

7. **Docker Dependencies**
   - ✅ API Gateway depends on tất cả services (không chỉ user + product)
   - ✅ Health checks với `condition: service_healthy`

---

## 🎯 Quick Start

### 1️⃣ Start Everything (First Time)

```bash
# 1. Install tools
make setup

# 2. Start infrastructure + services
make dev

# 3. Check logs
make dev-logs
```

### 2️⃣ Development Workflow

```bash
# Start infrastructure only
docker-compose up -d postgres redis rabbitmq

# Run migrations
make migrate-up

# Build services
make build-all

# Run services locally
cd services/user-service && ./user-service.exe
cd services/product-service && ./product-service.exe
# ... repeat for other services
```

### 3️⃣ Test API

```bash
# API Gateway Health Check
curl http://localhost:8000/health

# User Service Health Check
curl http://localhost:8001/health

# Create User (via API Gateway)
curl -X POST http://localhost:8000/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123!"
  }'
```

---

## 🔍 Troubleshooting

### Issue: Service không start được

**Giải pháp:**
```bash
# Check logs
docker-compose logs <service-name>

# Example
docker-compose logs user-service
```

### Issue: Database connection failed

**Giải pháp:**
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check databases exist
docker exec -it ecommerce-postgres psql -U postgres -c "\l"

# Re-create databases
make db-reset
make migrate-up
```

### Issue: Port already in use

**Giải pháp:**
```bash
# Stop all containers
make stop-all

# Check ports
netstat -ano | findstr "8000"
netstat -ano | findstr "5432"

# Kill process on port (Windows)
taskkill /PID <PID> /F
```

### Issue: Health check failing

**Giải pháp:**
```bash
# Tăng start_period trong docker-compose.yaml
healthcheck:
  start_period: 60s  # Tăng từ 40s lên 60s

# Hoặc test manual
docker exec -it ecommerce-user-service wget --spider http://localhost:8001/health
```

---

## 📚 Additional Resources

- **Architecture Docs**: `docs/architecture/system_design.md`
- **API Documentation**: `docs/api/swagger.yaml`
- **Deployment Guide**: `docs/deployment/deployment_guide.md`
- **Database Schema**: `docs/architecture/database_schema.md`

---

## 🚨 Production Checklist

Trước khi deploy lên production:

- [ ] Đổi tất cả passwords (PostgreSQL, RabbitMQ, JWT secret)
- [ ] Config SMTP credentials cho Notification Service
- [ ] Setup Stripe/PayPal keys cho Payment Service
- [ ] Enable HTTPS với certificates
- [ ] Configure CORS với domain thật
- [ ] Setup monitoring (Prometheus + Grafana)
- [ ] Enable distributed tracing (Jaeger)
- [ ] Configure backup cho databases
- [ ] Setup log aggregation (ELK/Loki)
- [ ] Configure auto-scaling (Kubernetes)
- [ ] Setup CI/CD pipeline
- [ ] Enable rate limiting
- [ ] Configure firewall rules
- [ ] Setup health check endpoints
- [ ] Test disaster recovery procedures

---

**Last Updated**: 2025-10-15
**Version**: 2.0.0
