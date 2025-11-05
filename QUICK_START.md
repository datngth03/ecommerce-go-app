# Hướng Dẫn Khởi Động Hệ Thống - Quick Start Guide

## Yêu Cầu
- Docker Desktop đang chạy
- Go 1.24+ đã cài đặt
- Port 5432, 6379, 5672, 8000-8006, 9001-9006 không bị chiếm dụng

## Cách 1: Khởi Động Với Docker Compose (Khuyến Nghị)

### Bước 1: Khởi động tất cả services
```powershell
cd d:\WorkSpace\Personal\Go\ecommerce-go-app
docker-compose up -d
```

Lệnh này sẽ khởi động:
- **Infrastructure**: PostgreSQL, Redis, RabbitMQ, Nginx, Prometheus, Grafana, Jaeger
- **Microservices**: 7 services (User, Product, Order, Payment, Inventory, Notification, API Gateway)

### Bước 2: Kiểm tra trạng thái
```powershell
docker-compose ps
```

Tất cả services phải có status "Up" và "healthy"

### Bước 3: Xem logs (nếu có lỗi)
```powershell
# Xem logs tất cả services
docker-compose logs -f

# Xem logs của service cụ thể
docker-compose logs -f api-gateway
docker-compose logs -f user-service
```

### Bước 4: Kiểm tra health
```powershell
# API Gateway
curl http://localhost:8000/health

# User Service
curl http://localhost:8001/health

# Product Service  
curl http://localhost:8002/health
```

### Bước 5: Chạy tests
```powershell
.\tests\e2e\test-api.ps1
```

### Dừng tất cả services
```powershell
docker-compose down
```

### Xóa tất cả và reset
```powershell
docker-compose down -v  # Xóa cả volumes (databases)
```

---

## Cách 2: Khởi Động Local (Không dùng Docker cho services)

Phương pháp này chạy infrastructure (DB, Redis) bằng Docker nhưng chạy services bằng Go trực tiếp.

### Bước 1: Start infrastructure
```powershell
docker-compose up -d postgres redis rabbitmq
```

### Bước 2: Chờ infrastructure sẵn sàng
```powershell
Start-Sleep -Seconds 10
docker-compose ps postgres redis rabbitmq
```

### Bước 3: Setup environment files
```powershell
.\scripts\setup-env.ps1
```

Script này sẽ tạo file `.env` cho mỗi service với password đúng (`postgres123`)

### Bước 4: Start services với script
```powershell
.\scripts\quick-start-phase2.ps1
```

Hoặc start từng service riêng:

**Terminal 1 - User Service:**
```powershell
cd services\user-service
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

**Terminal 2 - Product Service:**
```powershell
cd services\product-service
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

**Terminal 3 - Order Service:**
```powershell
cd services\order-service
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

**Terminal 4 - Payment Service:**
```powershell
cd services\payment-service
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

**Terminal 5 - Inventory Service:**
```powershell
cd services\inventory-service
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

**Terminal 6 - Notification Service:**
```powershell
cd services\notification-service
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

**Terminal 7 - API Gateway:**
```powershell
cd services\api-gateway
$env:DB_PASSWORD="postgres123"
go run cmd/main.go
```

---

## Cách 3: Sử Dụng Makefile

### Xem tất cả commands
```powershell
make help
```

### Build tất cả services
```powershell
make build
```

### Start với Docker
```powershell
make docker-up
```

### Stop
```powershell
make docker-down
```

### Run migrations
```powershell
make migrate-up
```

---

## Troubleshooting

### Lỗi: Port already in use
```powershell
# Tìm process đang dùng port
netstat -ano | findstr :8000

# Kill process (thay PID)
taskkill /PID <PID> /F
```

### Lỗi: Database connection failed
```powershell
# Kiểm tra PostgreSQL
docker-compose ps postgres

# Restart PostgreSQL
docker-compose restart postgres

# Kiểm tra password trong docker-compose.yaml
# POSTGRES_PASSWORD phải là: postgres123
```

### Lỗi: Cannot connect to Docker daemon
```
- Mở Docker Desktop
- Đợi Docker Desktop sẵn sàng (icon Docker màu xanh)
- Chạy lại docker-compose up -d
```

### Services không start
```powershell
# Xem logs chi tiết
docker-compose logs [service-name]

# Rebuild image
docker-compose build [service-name]

# Force recreate
docker-compose up -d --force-recreate
```

### Lỗi: go.sum missing
```powershell
# Tạo go.sum cho mỗi service
cd services/user-service
go mod tidy

cd ../product-service
go mod tidy

# ... làm tương tự cho tất cả services
```

---

## Kiểm Tra Hệ Thống

### Services đang chạy
```powershell
docker-compose ps
```

Expected output: Tất cả services có STATE = "Up" hoặc "Up (healthy)"

### Health Checks
```powershell
curl http://localhost:8000/health  # API Gateway
curl http://localhost:8001/health  # User Service  
curl http://localhost:8002/health  # Product Service
curl http://localhost:8003/health  # Order Service
curl http://localhost:8004/health  # Notification Service
curl http://localhost:8005/health  # Inventory Service
curl http://localhost:8006/health  # Payment Service
```

### Databases
```powershell
# Connect to PostgreSQL
docker exec -it ecommerce-postgres psql -U postgres

# List databases
\l

# Expected: users_db, products_db, orders_db, payments_db, inventory_db, notifications_db
```

### Redis
```powershell
docker exec -it ecommerce-redis redis-cli ping
# Expected: PONG
```

---

## Testing

### Automated Tests
```powershell
.\tests\e2e\test-api.ps1
```

### Manual Test với Postman
1. Import collection: `docs/api/postman/ecommerce-phase2.postman_collection.json`
2. Run requests theo thứ tự:
   - Register User
   - Login (token auto-saved)
   - Create Product
   - Add to Cart
   - Create Order
   - Process Payment

### Manual Test với curl
```powershell
# Register
curl -X POST http://localhost:8000/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"test@test.com","password":"Pass123!","username":"test","full_name":"Test User"}'

# Login
curl -X POST http://localhost:8000/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"test@test.com","password":"Pass123!"}'

# List Products
curl http://localhost:8000/api/v1/products
```

---

## Monitoring

### Prometheus
- URL: http://localhost:9090
- Metrics của tất cả services

### Grafana
- URL: http://localhost:3000
- Default login: admin/admin
- Dashboards: Service metrics, Business metrics

### Jaeger (Tracing)
- URL: http://localhost:16686
- Distributed tracing

---

## Configuration Files

### docker-compose.yaml
- Cấu hình tất cả services và infrastructure
- Environment variables
- Ports mapping
- Networks
- Volumes

### .env files (cho mỗi service)
- `services/<service-name>/.env`
- Được tạo từ `.env.example` bằng script `setup-env.ps1`
- **QUAN TRỌNG**: `DB_PASSWORD=postgres123` phải khớp với docker-compose.yaml

### Makefile
- Build commands
- Test commands  
- Docker commands
- Migration commands

---

## Scripts Có Sẵn

| Script | Mô Tả |
|--------|-------|
| `scripts/setup-env.ps1` | Tạo .env files cho tất cả services |
| `scripts/create-databases.ps1` | Tạo databases trong PostgreSQL |
| `scripts/quick-start-phase2.ps1` | Start infrastructure + migrations + services |
| `tests/e2e/test-api.ps1` | Automated API tests |

---

## Port Reference

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| API Gateway | 8000 | - |
| User Service | 8001 | 9001 |
| Product Service | 8002 | 9002 |
| Order Service | 8003 | 9003 |
| Notification Service | 8004 | 9004 |
| Inventory Service | 8005 | 9005 |
| Payment Service | 8006 | 9006 |
| PostgreSQL | 5432 | - |
| Redis | 6379 | - |
| RabbitMQ | 5672 | - |
| RabbitMQ Management | 15672 | - |
| Prometheus | 9090 | - |
| Grafana | 3000 | - |
| Jaeger | 16686 | - |

---


