# API Gateway Service

API Gateway service là điểm vào chính cho toàn bộ hệ thống ecommerce microservices. Nó xử lý routing, authentication, rate limiting và load balancing cho tất cả các requests.

## 🏗️ Architecture

```
Client Request
     ↓
API Gateway (Port 8000)
     ↓
  ┌─────────────────────────────┐
  │                             │
  ↓                             ↓
User Service          Product Service
(gRPC: 9090)         (gRPC: 9091)
(HTTP: 8081)         (HTTP: 8082)
```

## 🚀 Quick Start

### Prerequisites
- Go 1.24.0+
- Docker (optional)

### Local Development

#### Build và Run
```powershell
# Build binary
go build -o api-gateway.exe ./cmd/main.go

# Run
.\api-gateway.exe
```

Hoặc sử dụng script:
```powershell
# Build only
.\run-api-gateway.ps1 build

# Build and run
.\run-api-gateway.ps1 run

# Build Docker image
.\run-api-gateway.ps1 docker

# Clean artifacts
.\run-api-gateway.ps1 clean
```

### Docker

#### Build image
```bash
docker build -t ecommerce-api-gateway:latest .
```

#### Run container
```bash
docker run -p 8000:8000 \
  -e USER_SERVICE_GRPC_ADDR=user-service:9090 \
  -e PRODUCT_SERVICE_GRPC_ADDR=product-service:9091 \
  ecommerce-api-gateway:latest
```

### Docker Compose
```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down
```

## 📡 API Endpoints

### Health Checks
- `GET /health` - Health check endpoint
- `GET /ready` - Readiness probe

### Authentication (`/api/v1/auth`)
- `POST /register` - Register new user
- `POST /login` - User login
- `POST /refresh` - Refresh access token

### Users (`/api/v1/users`)
- `GET /:id` - Get user by ID
- `GET /me` - Get current user profile (auth required)
- `PUT /:id` - Update user (auth required)
- `DELETE /:id` - Delete user (auth required)

### Products (`/api/v1/products`)
- `GET /` - List all products
- `GET /:id` - Get product by ID
- `POST /` - Create product (auth required)
- `PUT /:id` - Update product (auth required)
- `DELETE /:id` - Delete product (auth required)

### Categories (`/api/v1/categories`)
- `GET /` - List all categories
- `GET /:id` - Get category by ID
- `POST /` - Create category (auth required)
- `PUT /:id` - Update category (auth required)
- `DELETE /:id` - Delete category (auth required)

## ⚙️ Configuration

Environment variables:

### Server
- `SERVER_PORT` - HTTP server port (default: 8000)
- `SERVER_HOST` - Server host (default: 0.0.0.0)
- `GIN_MODE` - Gin mode: debug/release (default: debug)

### Services - gRPC
- `USER_SERVICE_GRPC_ADDR` - User service gRPC address
- `PRODUCT_SERVICE_GRPC_ADDR` - Product service gRPC address
- `ORDER_SERVICE_GRPC_ADDR` - Order service gRPC address
- `PAYMENT_SERVICE_GRPC_ADDR` - Payment service gRPC address
- `INVENTORY_SERVICE_GRPC_ADDR` - Inventory service gRPC address
- `NOTIFICATION_SERVICE_GRPC_ADDR` - Notification service gRPC address

### Services - HTTP (Fallback)
- `USER_SERVICE_HTTP_ADDR` - User service HTTP address
- `PRODUCT_SERVICE_HTTP_ADDR` - Product service HTTP address
- etc.

### Timeouts
- `SERVICE_TIMEOUT` - Default timeout for service calls (default: 30s)

### Authentication
- `JWT_SECRET` - JWT signing secret
- `JWT_EXPIRATION_HOURS` - JWT expiration in hours (default: 24)
- `REFRESH_TOKEN_EXP_DAYS` - Refresh token expiration in days (default: 7)
- `ENABLE_AUTH` - Enable authentication middleware (default: true)

### Rate Limiting
- `RATE_LIMIT_ENABLED` - Enable rate limiting (default: true)
- `RATE_LIMIT_REQUESTS_PER_MIN` - Max requests per minute (default: 100)
- `RATE_LIMIT_BURST_SIZE` - Burst size (default: 50)

### CORS
- `CORS_ALLOWED_ORIGINS` - Comma-separated allowed origins
- `CORS_ALLOW_CREDENTIALS` - Allow credentials (default: true)

## 🧪 Testing

### Health Check
```bash
curl http://localhost:8000/health
```

### Register User
```bash
curl -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "name": "Test User",
    "phone": "0123456789",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get Products
```bash
curl http://localhost:8000/api/v1/products
```

### Get Product by ID
```bash
curl http://localhost:8000/api/v1/products/1
```

### Create Product (with auth)
```bash
curl -X POST http://localhost:8000/api/v1/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "New Product",
    "description": "Product description",
    "price": 99.99,
    "stock": 100,
    "category_id": 1
  }'
```

## 🏗️ Project Structure

```
api-gateway/
├── cmd/
│   └── main.go              # Application entrypoint
├── internal/
│   ├── clients/             # gRPC clients
│   │   ├── clients.go       # Client manager
│   │   ├── product_client.go
│   │   └── user_client.go
│   ├── config/              # Configuration
│   │   └── config.go
│   ├── handler/             # HTTP handlers
│   │   ├── product_handler.go
│   │   └── user_handler.go
│   ├── middleware/          # Middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logging.go
│   │   └── rate_limit.go
│   ├── models/              # Data models
│   │   └── response.go
│   └── proxy/               # Service proxies
│       ├── product_proxy.go
│       └── user_proxy.go
├── Dockerfile               # Docker build
├── .dockerignore
├── go.mod
└── README.md
```

## 🔧 Development

### Add New Service Integration

1. Create gRPC client in `internal/clients/`
2. Create proxy adapter in `internal/proxy/`
3. Create HTTP handlers in `internal/handler/`
4. Register routes in `cmd/main.go`

### Add Middleware

1. Create middleware in `internal/middleware/`
2. Register in router setup in `cmd/main.go`

## 📊 Monitoring

### Metrics
- Request count
- Response time
- Error rate
- Service health

### Logs
- Structured JSON logs
- Request/Response logging
- Error tracking

## 🚨 Troubleshooting

### Cannot connect to services
- Check service addresses in config
- Verify services are running
- Check network connectivity

### Authentication errors
- Verify JWT secret matches across services
- Check token expiration
- Verify user-service is reachable

### Rate limiting errors
- Check rate limit configuration
- Verify burst size settings
- Consider increasing limits for high-traffic endpoints

## 📝 License

Copyright © 2025 Ecommerce Go App
