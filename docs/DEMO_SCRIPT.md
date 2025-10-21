# E-Commerce Microservices - Demo Script

## Overview

This demo script showcases the complete end-to-end functionality of the E-Commerce microservices platform, demonstrating key features from user registration through payment processing.

**Duration**: 15-20 minutes  
**Audience**: Technical stakeholders, investors, development teams  
**Prerequisites**: Docker services running on localhost:8000

---

## Table of Contents

1. [Setup & Preparation](#setup--preparation)
2. [Demo Flow](#demo-flow)
3. [Talking Points](#talking-points)
4. [Sample Data](#sample-data)
5. [Common Questions & Answers](#common-questions--answers)
6. [Troubleshooting](#troubleshooting)

---

## Setup & Preparation

### Pre-Demo Checklist

- [ ] Start all services: `docker-compose up -d`
- [ ] Verify health: `curl http://localhost:8000/health`
- [ ] Clear previous demo data (optional)
- [ ] Open Postman collection
- [ ] Open architecture diagram
- [ ] Open Grafana dashboard (http://localhost:3000)
- [ ] Prepare terminal windows (logs, metrics)
- [ ] Test network connectivity

### Environment Setup

```bash
# 1. Start services
cd ecommerce-go-app
docker-compose up -d

# 2. Wait for services to be healthy (30-60 seconds)
docker-compose ps

# 3. Verify API Gateway
curl http://localhost:8000/health

# 4. Open monitoring dashboards
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
# RabbitMQ: http://localhost:15672 (guest/guest)
```

### Demo Variables

Set these for quick reference during demo:

```bash
export BASE_URL="http://localhost:8000/api/v1"
export DEMO_EMAIL="demo@ecommerce.com"
export DEMO_PASSWORD="DemoPass123!"
export DEMO_NAME="Demo User"
```

---

## Demo Flow

### Act 1: User Management (3 minutes)

#### Scene 1: User Registration

**Talking Points**:
> "Let's start by creating a new user account. Our user service handles authentication and profile management with JWT tokens."

**API Call**:
```bash
curl -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@ecommerce.com",
    "password": "DemoPass123!",
    "name": "Demo User",
    "phone": "+1234567890"
  }'
```

**Expected Response**:
```json
{
  "data": {
    "success": true,
    "message": "User created successfully"
  }
}
```

**Show**:
- User service logs: `docker-compose logs -f --tail=20 user-service`
- Database entry (optional): PostgreSQL users_db

#### Scene 2: User Login

**Talking Points**:
> "Now we'll authenticate. The system returns JWT tokens for secure API access with 24-hour expiration."

**API Call**:
```bash
curl -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo@ecommerce.com",
    "password": "DemoPass123!"
  }'
```

**Expected Response**:
```json
{
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-10-22T12:00:00Z"
  }
}
```

**Save Token**:
```bash
export ACCESS_TOKEN="<access_token_from_response>"
```

#### Scene 3: Get User Profile

**Talking Points**:
> "With our token, we can access protected endpoints like the user profile."

**API Call**:
```bash
curl -X GET $BASE_URL/users/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response**:
```json
{
  "data": {
    "id": 1,
    "email": "demo@ecommerce.com",
    "name": "Demo User",
    "phone": "+1234567890",
    "is_active": true,
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

---

### Act 2: Product Catalog (4 minutes)

#### Scene 1: Create Product

**Talking Points**:
> "Let's add a product to our catalog. The product service manages our inventory catalog with categories."

**API Call**:
```bash
curl -X POST $BASE_URL/products \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Headphones Pro",
    "description": "Premium noise-canceling headphones with 30-hour battery life",
    "price": 299.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "image_url": "https://example.com/headphones.jpg"
  }'
```

**Expected Response**:
```json
{
  "data": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "name": "Wireless Headphones Pro",
    "slug": "wireless-headphones-pro",
    "description": "Premium noise-canceling headphones...",
    "price": 299.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "is_active": true,
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

**Save Product ID**:
```bash
export PRODUCT_ID="<product_id_from_response>"
```

**Show**:
- Product service logs
- gRPC communication in logs

#### Scene 2: Add Stock to Inventory

**Talking Points**:
> "Before customers can purchase, we need to add stock. The inventory service tracks available quantities separately."

**API Call**:
```bash
curl -X POST $BASE_URL/inventory/$PRODUCT_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "WHP-001",
    "stock": 100
  }'
```

**Expected Response**:
```json
{
  "message": "inventory created successfully"
}
```

#### Scene 3: Check Product Availability

**Talking Points**:
> "Customers can check if products are in stock before adding to cart."

**API Call**:
```bash
curl -X POST $BASE_URL/inventory/check-availability \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "product_id": "'$PRODUCT_ID'",
        "quantity": 2
      }
    ]
  }'
```

**Expected Response**:
```json
{
  "available": true,
  "message": "availability checked successfully",
  "unavailable_items": []
}
```

#### Scene 4: Browse Products

**Talking Points**:
> "Customers can browse our catalog with pagination and filtering."

**API Call**:
```bash
curl -X GET "$BASE_URL/products?page=1&page_size=10" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Act 3: Shopping Cart (3 minutes)

#### Scene 1: Add Item to Cart

**Talking Points**:
> "Now let's shop! The order service manages shopping carts with real-time price calculations."

**API Call**:
```bash
curl -X POST $BASE_URL/cart \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'$PRODUCT_ID'",
    "quantity": 2
  }'
```

**Expected Response**:
```json
{
  "message": "item added to cart successfully",
  "data": {
    "user_id": 1,
    "items": [
      {
        "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "quantity": 2,
        "price": 299.99
      }
    ],
    "total": 599.98
  }
}
```

**Show**:
- Inter-service communication (order → product → inventory)
- Real-time inventory reservation

#### Scene 2: View Cart

**Talking Points**:
> "Customers can view their cart with calculated totals."

**API Call**:
```bash
curl -X GET $BASE_URL/cart \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response**:
```json
{
  "data": {
    "user_id": 1,
    "items": [
      {
        "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "product_name": "Wireless Headphones Pro",
        "quantity": 2,
        "price": 299.99,
        "subtotal": 599.98
      }
    ],
    "total": 599.98,
    "item_count": 2
  }
}
```

---

### Act 4: Order Processing (4 minutes)

#### Scene 1: Create Order

**Talking Points**:
> "When ready to checkout, the order service creates an order and reserves inventory."

**API Call**:
```bash
curl -X POST $BASE_URL/orders \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "shipping_address": "123 Tech Street, San Francisco, CA 94105",
    "payment_method": "stripe"
  }'
```

**Expected Response**:
```json
{
  "data": {
    "id": "order-uuid-1234",
    "user_id": 1,
    "status": "PENDING",
    "total_amount": 599.98,
    "shipping_address": "123 Tech Street, San Francisco, CA 94105",
    "payment_method": "stripe",
    "items": [
      {
        "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
        "quantity": 2,
        "price": 299.99,
        "subtotal": 599.98
      }
    ],
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

**Save Order ID**:
```bash
export ORDER_ID="<order_id_from_response>"
```

**Show**:
- Order service logs
- Inventory reservation in database
- Cart cleared automatically

#### Scene 2: View Order Details

**Talking Points**:
> "Customers can track their order status and details."

**API Call**:
```bash
curl -X GET $BASE_URL/orders/$ORDER_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

---

### Act 5: Payment Processing (4 minutes)

#### Scene 1: Save Payment Method

**Talking Points**:
> "Customers can save payment methods for future purchases. We integrate with Stripe for secure payment processing."

**API Call**:
```bash
curl -X POST $BASE_URL/payment-methods \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "method_type": "card",
    "gateway_method_id": "pm_stripe_visa_4242",
    "is_default": true
  }'
```

**Expected Response**:
```json
{
  "success": true,
  "message": "payment method saved successfully"
}
```

#### Scene 2: Process Payment

**Talking Points**:
> "Now let's process the payment. The payment service handles the transaction and updates the order status."

**API Call**:
```bash
curl -X POST $BASE_URL/payments \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "'$ORDER_ID'",
    "amount": 599.98,
    "method": "stripe",
    "currency": "USD"
  }'
```

**Expected Response**:
```json
{
  "data": {
    "id": "payment-uuid-5678",
    "order_id": "order-uuid-1234",
    "amount": 599.98,
    "currency": "USD",
    "method": "stripe",
    "status": "COMPLETED",
    "created_at": "2025-10-21T10:00:00Z"
  }
}
```

**Show**:
- Payment service logs
- Order status updated to "CONFIRMED"
- Inventory deducted from available stock
- Notification service triggered (email confirmation)

#### Scene 3: Verify Inventory Deduction

**Talking Points**:
> "Notice how the inventory is automatically updated after successful payment."

**API Call**:
```bash
curl -X GET $BASE_URL/inventory/$PRODUCT_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response**:
```json
{
  "data": {
    "product_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "available": 98,
    "reserved": 0,
    "total": 98
  }
}
```

---

### Act 6: Monitoring & Observability (2 minutes)

**Talking Points**:
> "Our platform includes comprehensive monitoring and observability."

**Show**:

1. **Grafana Dashboard** (http://localhost:3000):
   - Request rate across all services
   - Average response time
   - Error rate
   - Database connection pool usage

2. **Prometheus Metrics** (http://localhost:9090):
   - `http_requests_total{service="api-gateway"}`
   - `http_request_duration_seconds{service="order-service"}`
   - `grpc_server_handled_total{service="payment-service"}`

3. **Service Logs**:
   ```bash
   # Show distributed tracing correlation IDs
   docker-compose logs --tail=50 | grep "request_id"
   ```

4. **RabbitMQ Dashboard** (http://localhost:15672):
   - Event publishing (order created, payment completed)
   - Message queues for notifications
   - Consumer statistics

---

## Talking Points

### Opening (1 minute)

> "Today I'll demonstrate our E-Commerce microservices platform built with Go. This is a production-ready system featuring six independent microservices communicating via gRPC, with an API Gateway for external clients."

**Key Points**:
- **Technology Stack**: Go 1.21, gRPC, PostgreSQL, Redis, RabbitMQ
- **Architecture**: Microservices with API Gateway pattern
- **Scalability**: Containerized with Docker, ready for Kubernetes
- **Security**: JWT authentication, password hashing, secure secrets management

### Architecture Overview (2 minutes)

> "Our architecture follows microservices best practices with clear separation of concerns."

**Services**:
1. **API Gateway** (Port 8000): Single entry point, REST to gRPC translation
2. **User Service** (Port 9001): Authentication, user profiles
3. **Product Service** (Port 9002): Product catalog management
4. **Inventory Service** (Port 9003): Stock tracking, availability
5. **Order Service** (Port 9004): Cart, orders, order processing
6. **Payment Service** (Port 9005): Payment processing, saved methods
7. **Notification Service** (Port 9006): Email notifications (async)

**Infrastructure**:
- **Database**: PostgreSQL 15 (separate DB per service)
- **Cache**: Redis 7 for session management
- **Message Queue**: RabbitMQ for async events
- **Monitoring**: Prometheus + Grafana

### Key Features

#### 1. Microservices Architecture
> "Each service is independently deployable, scalable, and maintainable. Services communicate via gRPC for high performance."

#### 2. Event-Driven Design
> "We use RabbitMQ for asynchronous communication. For example, when an order is created, we publish an event that triggers email notifications without blocking the order creation."

#### 3. Data Consistency
> "We implement eventual consistency with event sourcing. Each service owns its data, and we use saga patterns for distributed transactions."

#### 4. Security
> "All endpoints are secured with JWT tokens. Passwords are hashed with bcrypt. We use TLS for production deployments."

#### 5. Observability
> "Every request is traced with correlation IDs. We collect metrics in Prometheus and visualize in Grafana. Centralized logging with structured JSON logs."

#### 6. Testing
> "We have comprehensive test coverage:
> - Unit tests for business logic
> - Integration tests for database operations
> - E2E tests for complete user flows
> - Load tests for performance validation"

---

## Sample Data

### Demo Products

```json
[
  {
    "name": "Wireless Headphones Pro",
    "description": "Premium noise-canceling headphones",
    "price": 299.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "stock": 100
  },
  {
    "name": "Smart Watch Ultra",
    "description": "Fitness tracking with GPS and heart rate monitor",
    "price": 499.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "stock": 50
  },
  {
    "name": "4K Webcam",
    "description": "Professional webcam for streaming",
    "price": 199.99,
    "category_id": "63b957bf-0f16-4f32-8c34-8215ccc5bc46",
    "stock": 75
  }
]
```

### Demo Users

```json
[
  {
    "email": "demo@ecommerce.com",
    "password": "DemoPass123!",
    "name": "Demo User"
  },
  {
    "email": "admin@ecommerce.com",
    "password": "AdminPass123!",
    "name": "Admin User",
    "role": "admin"
  }
]
```

### Seed Script

Create `scripts/seed-demo-data.ps1`:

```powershell
# Seed demo data for presentation

$BASE_URL = "http://localhost:8000/api/v1"

# Register demo user
$registerResponse = Invoke-RestMethod -Uri "$BASE_URL/auth/register" -Method Post -ContentType "application/json" -Body @"
{
  "email": "demo@ecommerce.com",
  "password": "DemoPass123!",
  "name": "Demo User",
  "phone": "+1234567890"
}
"@

Write-Host "Demo user created" -ForegroundColor Green

# Login to get token
$loginResponse = Invoke-RestMethod -Uri "$BASE_URL/auth/login" -Method Post -ContentType "application/json" -Body @"
{
  "email": "demo@ecommerce.com",
  "password": "DemoPass123!"
}
"@

$token = $loginResponse.data.access_token
Write-Host "Logged in successfully" -ForegroundColor Green

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Create products
$products = @(
    @{
        name = "Wireless Headphones Pro"
        description = "Premium noise-canceling headphones"
        price = 299.99
        category_id = "63b957bf-0f16-4f32-8c34-8215ccc5bc46"
    },
    @{
        name = "Smart Watch Ultra"
        description = "Fitness tracking with GPS"
        price = 499.99
        category_id = "63b957bf-0f16-4f32-8c34-8215ccc5bc46"
    },
    @{
        name = "4K Webcam"
        description = "Professional webcam"
        price = 199.99
        category_id = "63b957bf-0f16-4f32-8c34-8215ccc5bc46"
    }
)

foreach ($product in $products) {
    $productResponse = Invoke-RestMethod -Uri "$BASE_URL/products" -Method Post -Headers $headers -Body ($product | ConvertTo-Json)
    $productId = $productResponse.data.id
    
    # Add inventory
    Invoke-RestMethod -Uri "$BASE_URL/inventory/$productId" -Method Post -Headers $headers -Body @"
{
  "sku": "SKU-$(Get-Random -Maximum 9999)",
  "stock": 100
}
"@
    
    Write-Host "Created product: $($product.name)" -ForegroundColor Green
}

Write-Host "`nDemo data seeded successfully!" -ForegroundColor Green
```

---

## Common Questions & Answers

### Q: How do you handle service failures?

**A**: "We implement retry logic with exponential backoff, circuit breakers, and graceful degradation. For example, if the payment service is down, we queue payment requests in RabbitMQ for later processing."

### Q: How do you ensure data consistency across services?

**A**: "We use the Saga pattern for distributed transactions. Each service publishes domain events, and we have compensating transactions for rollback scenarios. For example, if payment fails, we automatically release the inventory reservation."

### Q: What about database scaling?

**A**: "Each service has its own database, allowing independent scaling. We use connection pooling, read replicas for read-heavy services, and caching with Redis. PostgreSQL can handle millions of rows efficiently with proper indexing."

### Q: How do you deploy this to production?

**A**: "We have multiple deployment options:
1. Docker Compose for smaller deployments
2. Kubernetes for large-scale production with auto-scaling
3. CI/CD pipeline with GitHub Actions for automated testing and deployment
4. Blue-green deployments for zero-downtime updates"

### Q: What's your test coverage?

**A**: "We maintain over 80% code coverage with:
- Unit tests for all business logic
- Integration tests for database operations
- E2E tests covering critical user flows
- Load tests simulating 10,000 concurrent users"

### Q: How do you monitor production?

**A**: "We use Prometheus for metrics collection, Grafana for visualization, and structured logging with correlation IDs for distributed tracing. We have alerts configured for error rates, response times, and resource usage."

### Q: How do you handle API versioning?

**A**: "We use URL versioning (v1, v2) in the API Gateway. Old versions are maintained for backward compatibility with deprecated warnings. We follow semantic versioning for internal service contracts."

### Q: What about security?

**A**: "Multiple layers:
- JWT tokens with short expiration (24 hours)
- Password hashing with bcrypt
- Input validation and sanitization
- Rate limiting (1000 requests/hour)
- TLS encryption in production
- Database encryption at rest
- Regular security audits and dependency updates"

---

## Troubleshooting

### Service Not Responding

```bash
# Check service status
docker-compose ps

# Restart specific service
docker-compose restart user-service

# View logs
docker-compose logs -f user-service
```

### Database Connection Error

```bash
# Check PostgreSQL status
docker-compose exec postgres pg_isready

# View database logs
docker-compose logs postgres
```

### Port Already in Use

```bash
# Find process using port
netstat -ano | findstr :8000

# Kill process (Windows)
taskkill /PID <process_id> /F

# Or change port in docker-compose.yaml
```

### Clear Demo Data

```bash
# Reset databases
docker-compose down -v
docker-compose up -d postgres
make migrate-up
docker-compose up -d
```

---

## Demo Variations

### Quick Demo (5 minutes)
Focus on: Registration → Login → Create Product → Add to Cart → Create Order

### Technical Deep Dive (30 minutes)
Include: Architecture walkthrough → Code review → Database schema → Deployment process → Monitoring setup

### Business Demo (10 minutes)
Focus on: User experience → Feature highlights → Business metrics → Scalability story

---

## Visual Aids

### Architecture Diagram
Location: `docs/architecture/system_design.md`

### Database Schema
Location: `docs/architecture/database_schema.md`

### API Flow Diagrams
Create sequence diagrams for:
- User registration and login
- Product purchase flow
- Payment processing flow

### Performance Charts
- Response time distribution
- Throughput over time
- Error rate trends

---

## Post-Demo Resources

Provide attendees with:
1. **GitHub Repository**: Link to source code
2. **API Documentation**: `docs/API_REFERENCE.md`
3. **Postman Collection**: `docs/api/postman/ecommerce.postman_collection.json`
4. **Setup Guide**: `docs/QUICK_START.md`
5. **Contact Information**: For questions and follow-up

---

## Closing Remarks

> "This platform demonstrates modern microservices architecture with Go, showcasing scalability, reliability, and maintainability. It's production-ready and can handle enterprise-scale workloads. Thank you for your time, and I'm happy to answer any questions!"

**Key Takeaways**:
1. ✅ Scalable microservices architecture
2. ✅ High-performance gRPC communication
3. ✅ Comprehensive monitoring and observability
4. ✅ Production-ready with Docker and Kubernetes support
5. ✅ Secure with JWT authentication
6. ✅ Well-tested with automated test suites

---

**Version**: 1.0.0  
**Last Updated**: October 2025  
**Next Review**: December 2025
