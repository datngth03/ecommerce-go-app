# Phase 2: End-to-End Testing Guide

## Prerequisites

### 1. Start Docker Desktop
Ensure Docker Desktop is running before proceeding with tests.

### 2. Start Infrastructure Services
```powershell
# Navigate to project root
cd d:\WorkSpace\Personal\Go\ecommerce-go-app

# Start infrastructure (PostgreSQL, Redis, RabbitMQ)
docker-compose up -d postgres redis rabbitmq

# Verify services are running
docker-compose ps
```

Expected output:
```
NAME                    STATUS
ecommerce-postgres      Up (healthy)
ecommerce-redis         Up (healthy)
ecommerce-rabbitmq      Up (healthy)
```

### 3. Run Database Migrations
```powershell
# Run migrations for all services
make migrate-up

# Or manually for each service:
cd services/user-service
go run cmd/migrate/main.go up

cd ../product-service
go run cmd/migrate/main.go up

cd ../order-service
go run cmd/migrate/main.go up

cd ../payment-service
go run cmd/migrate/main.go up

cd ../inventory-service
go run cmd/migrate/main.go up

cd ../notification-service
go run cmd/migrate/main.go up
```

### 4. Start All Microservices
```powershell
# Option 1: Start all services with docker-compose
docker-compose up -d

# Option 2: Start services individually for development
# Terminal 1 - User Service
cd services/user-service
go run cmd/main.go

# Terminal 2 - Product Service
cd services/product-service
go run cmd/main.go

# Terminal 3 - Order Service
cd services/order-service
go run cmd/main.go

# Terminal 4 - Payment Service
cd services/payment-service
go run cmd/main.go

# Terminal 5 - Inventory Service
cd services/inventory-service
go run cmd/main.go

# Terminal 6 - Notification Service
cd services/notification-service
go run cmd/main.go

# Terminal 7 - API Gateway
cd services/api-gateway
go run cmd/main.go
```

### 5. Verify Services Health
```powershell
# Check API Gateway health
curl http://localhost:8000/health

# Expected response:
# {"service":"api-gateway","status":"healthy"}
```

---

## Test Suite

### Test 1: User Service Flow

#### 1.1 Register New User
```powershell
curl -X POST http://localhost:8000/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "username": "testuser",
    "full_name": "Test User"
  }'
```

**Expected Response (201 Created):**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "username": "testuser",
    "full_name": "Test User"
  }
}
```

#### 1.2 Login
```powershell
curl -X POST http://localhost:8000/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Expected Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "username": "testuser"
  }
}
```

**Save the access_token for subsequent requests:**
```powershell
$TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 1.3 Get Profile
```powershell
curl http://localhost:8000/api/v1/users/me `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "email": "test@example.com",
    "username": "testuser",
    "full_name": "Test User",
    "created_at": "2025-10-15T14:30:00Z"
  }
}
```

---

### Test 2: Product Service Flow

#### 2.1 Create Product
```powershell
curl -X POST http://localhost:8000/api/v1/products `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop with RTX 4090",
    "price": 2499.99,
    "sku": "LAPTOP-001",
    "category_id": 1,
    "stock": 50
  }'
```

**Expected Response (201 Created):**
```json
{
  "message": "Product created successfully",
  "product": {
    "id": 1,
    "name": "Gaming Laptop",
    "price": 2499.99,
    "sku": "LAPTOP-001",
    "stock": 50
  }
}
```

**Save product_id:**
```powershell
$PRODUCT_ID = 1
```

#### 2.2 List Products
```powershell
curl "http://localhost:8000/api/v1/products?page=1&page_size=10"
```

**Expected Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "Gaming Laptop",
      "price": 2499.99,
      "stock": 50
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

#### 2.3 Get Product Details
```powershell
curl http://localhost:8000/api/v1/products/$PRODUCT_ID
```

**Expected Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop with RTX 4090",
    "price": 2499.99,
    "sku": "LAPTOP-001",
    "stock": 50
  }
}
```

---

### Test 3: Inventory Service Flow

#### 3.1 Check Stock
```powershell
curl http://localhost:8000/api/v1/inventory/$PRODUCT_ID `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "stock retrieved successfully",
  "data": {
    "product_id": 1,
    "available": 50,
    "reserved": 0,
    "total": 50
  }
}
```

#### 3.2 Check Availability
```powershell
curl -X POST http://localhost:8000/api/v1/inventory/check-availability `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "items": [
      {
        "product_id": 1,
        "quantity": 2
      }
    ]
  }'
```

**Expected Response (200 OK):**
```json
{
  "message": "availability checked successfully",
  "data": {
    "available": true,
    "unavailable_items": []
  }
}
```

---

### Test 4: Order Service Flow

#### 4.1 Add to Cart
```powershell
curl -X POST http://localhost:8000/api/v1/cart `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "product_id": 1,
    "quantity": 2,
    "price": 2499.99
  }'
```

**Expected Response (201 Created):**
```json
{
  "message": "item added to cart successfully"
}
```

#### 4.2 Get Cart
```powershell
curl http://localhost:8000/api/v1/cart `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "cart retrieved successfully",
  "data": {
    "items": [
      {
        "product_id": 1,
        "product_name": "Gaming Laptop",
        "quantity": 2,
        "price": 2499.99,
        "subtotal": 4999.98
      }
    ],
    "total": 4999.98
  }
}
```

#### 4.3 Create Order from Cart
```powershell
curl -X POST http://localhost:8000/api/v1/orders `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "shipping_address": "123 Main St, San Francisco, CA 94105",
    "payment_method": "stripe"
  }'
```

**Expected Response (201 Created):**
```json
{
  "message": "order created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "total_amount": 4999.98,
    "status": "pending",
    "items": [
      {
        "product_id": 1,
        "quantity": 2,
        "price": 2499.99
      }
    ]
  }
}
```

**Save order_id:**
```powershell
$ORDER_ID = 1
```

#### 4.4 Get Order Details
```powershell
curl http://localhost:8000/api/v1/orders/$ORDER_ID `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "order retrieved successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "status": "pending",
    "total_amount": 4999.98,
    "shipping_address": "123 Main St, San Francisco, CA 94105"
  }
}
```

#### 4.5 List Orders
```powershell
curl "http://localhost:8000/api/v1/orders?page=1&page_size=10" `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "orders retrieved successfully",
  "data": [
    {
      "id": 1,
      "status": "pending",
      "total_amount": 4999.98,
      "created_at": "2025-10-15T14:35:00Z"
    }
  ],
  "total": 1
}
```

---

### Test 5: Payment Service Flow

#### 5.1 Process Payment
```powershell
curl -X POST http://localhost:8000/api/v1/payments `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "order_id": "1",
    "amount": 4999.98,
    "method": "stripe",
    "currency": "USD"
  }'
```

**Expected Response (201 Created):**
```json
{
  "message": "payment processed successfully",
  "data": {
    "id": "1",
    "order_id": "1",
    "amount": 4999.98,
    "status": "pending",
    "method": "stripe"
  },
  "success": true
}
```

**Save payment_id:**
```powershell
$PAYMENT_ID = "1"
```

#### 5.2 Confirm Payment
```powershell
curl -X POST http://localhost:8000/api/v1/payments/$PAYMENT_ID/confirm `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "payment_intent_id": "pi_test_123456"
  }'
```

**Expected Response (200 OK):**
```json
{
  "message": "payment confirmed successfully",
  "data": {
    "id": "1",
    "status": "completed",
    "amount": 4999.98
  },
  "success": true
}
```

#### 5.3 Get Payment Details
```powershell
curl http://localhost:8000/api/v1/payments/$PAYMENT_ID `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "payment retrieved successfully",
  "data": {
    "id": "1",
    "order_id": "1",
    "amount": 4999.98,
    "status": "completed",
    "method": "stripe"
  }
}
```

#### 5.4 Get Payment History
```powershell
curl "http://localhost:8000/api/v1/payments?page=1&page_size=10" `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "payment history retrieved successfully",
  "data": [
    {
      "id": "1",
      "order_id": "1",
      "amount": 4999.98,
      "status": "completed",
      "created_at": "2025-10-15T14:36:00Z"
    }
  ],
  "total": 1
}
```

#### 5.5 Save Payment Method
```powershell
curl -X POST http://localhost:8000/api/v1/payment-methods `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{
    "method_type": "card",
    "gateway_method_id": "pm_test_visa4242",
    "is_default": true
  }'
```

**Expected Response (201 Created):**
```json
{
  "message": "payment method saved successfully",
  "data": {
    "id": "1",
    "method_type": "card",
    "is_default": true
  },
  "success": true
}
```

#### 5.6 Get Payment Methods
```powershell
curl http://localhost:8000/api/v1/payment-methods `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "payment methods retrieved successfully",
  "data": [
    {
      "id": "1",
      "method_type": "card",
      "is_default": true
    }
  ]
}
```

---

### Test 6: Inventory Verification

#### 6.1 Verify Stock After Order
```powershell
curl http://localhost:8000/api/v1/inventory/$PRODUCT_ID `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "stock retrieved successfully",
  "data": {
    "product_id": 1,
    "available": 48,
    "reserved": 2,
    "total": 50
  }
}
```

#### 6.2 Get Stock History (Admin)
```powershell
curl http://localhost:8000/api/v1/inventory/$PRODUCT_ID/history `
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "stock history retrieved successfully",
  "data": [
    {
      "id": 1,
      "product_id": 1,
      "type": "reserve",
      "quantity": 2,
      "reason": "order_1",
      "created_at": "2025-10-15T14:35:00Z"
    }
  ]
}
```

---

## Complete E-Commerce Flow Test

### Full User Journey Script
```powershell
# 1. Register User
$registerResponse = curl -X POST http://localhost:8000/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"customer@shop.com","password":"Shop123!","username":"customer","full_name":"John Doe"}' | ConvertFrom-Json

# 2. Login
$loginResponse = curl -X POST http://localhost:8000/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"customer@shop.com","password":"Shop123!"}' | ConvertFrom-Json

$TOKEN = $loginResponse.access_token

# 3. Browse Products
curl http://localhost:8000/api/v1/products

# 4. Check Product Stock
curl http://localhost:8000/api/v1/inventory/1

# 5. Add to Cart
curl -X POST http://localhost:8000/api/v1/cart `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{"product_id":1,"quantity":2,"price":2499.99}'

# 6. View Cart
curl http://localhost:8000/api/v1/cart `
  -H "Authorization: Bearer $TOKEN"

# 7. Create Order
$orderResponse = curl -X POST http://localhost:8000/api/v1/orders `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{"shipping_address":"123 Main St","payment_method":"stripe"}' | ConvertFrom-Json

$ORDER_ID = $orderResponse.data.id

# 8. Process Payment
$paymentResponse = curl -X POST http://localhost:8000/api/v1/payments `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d "{`"order_id`":`"$ORDER_ID`",`"amount`":4999.98,`"method`":`"stripe`"}" | ConvertFrom-Json

$PAYMENT_ID = $paymentResponse.data.id

# 9. Confirm Payment
curl -X POST http://localhost:8000/api/v1/payments/$PAYMENT_ID/confirm `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $TOKEN" `
  -d '{"payment_intent_id":"pi_test_123"}'

# 10. Verify Inventory Updated
curl http://localhost:8000/api/v1/inventory/1 `
  -H "Authorization: Bearer $TOKEN"

# 11. Check Order Status
curl http://localhost:8000/api/v1/orders/$ORDER_ID `
  -H "Authorization: Bearer $TOKEN"

Write-Host "✅ Complete E-Commerce Flow Test Completed!" -ForegroundColor Green
```

---

## Troubleshooting

### Services Not Starting
```powershell
# Check service logs
docker-compose logs -f [service-name]

# Restart specific service
docker-compose restart [service-name]

# Rebuild and restart
docker-compose up -d --build [service-name]
```

### Database Connection Issues
```powershell
# Check PostgreSQL is running
docker-compose ps postgres

# Connect to PostgreSQL
docker exec -it ecommerce-postgres psql -U postgres -d ecommerce

# List databases
\l

# List tables
\dt
```

### gRPC Connection Issues
```powershell
# Verify services are listening on gRPC ports
netstat -an | Select-String "9001|9002|9003|9004|9005|9006"

# Check service health
curl http://localhost:8001/health  # User Service
curl http://localhost:8002/health  # Product Service
curl http://localhost:8003/health  # Order Service
curl http://localhost:8006/health  # Payment Service
curl http://localhost:8005/health  # Inventory Service
curl http://localhost:8004/health  # Notification Service
```

### Authentication Errors
- Verify JWT token is valid and not expired
- Check middleware is properly configured
- Ensure `user_id` context is set correctly

---

## Performance Testing

### Load Test with Apache Bench
```powershell
# Install Apache Bench (comes with Apache)
# Test login endpoint
ab -n 1000 -c 10 -p login.json -T application/json http://localhost:8000/api/v1/auth/login

# Test product listing
ab -n 5000 -c 50 http://localhost:8000/api/v1/products
```

### Monitor Service Health
```powershell
# Watch logs in real-time
docker-compose logs -f api-gateway

# Monitor resource usage
docker stats
```

---

## Success Criteria

- ✅ All services start without errors
- ✅ User can register and login
- ✅ Products can be created and listed
- ✅ Cart operations work correctly
- ✅ Orders are created successfully
- ✅ Payments process and confirm
- ✅ Inventory updates after orders
- ✅ All API responses match expected format
- ✅ No errors in service logs

---

## Next Steps After Phase 2

Once all tests pass:
1. **Phase 3**: Add Postman collection for API documentation
2. **Phase 4**: Implement monitoring (Prometheus + Grafana)
3. **Phase 5**: Add distributed tracing (Jaeger)
4. **Phase 6**: Performance optimization and load testing
5. **Phase 7**: Security hardening and production deployment

---

**Generated:** October 15, 2025  
**Version:** 1.0  
**Status:** Ready for execution
