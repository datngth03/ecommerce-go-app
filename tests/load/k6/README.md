# Load Testing with K6

## Overview

This directory contains K6 load testing scripts to validate the **9.75x performance improvement** achieved through gRPC Connection Pool implementation.

## 📁 Test Scripts

### 1. `order-creation-test.js`
Tests Order Service performance with connection pool to multiple services.

**What it tests:**
- API Gateway → Order Service
- Order Service → Product, Inventory, Payment, Notification (via connection pool)
- **Target**: <200ms p95 (vs ~900ms without pool)

**Load profile:**
- Warm-up: 10 users (30s)
- Load: 50 users (2m)
- Peak: 100 users (2m)


### 2. `end-to-end-test.js` 
Comprehensive test of the complete order + payment flow.

**What it tests:**
- Full e-commerce flow: Auth → Browse → Cart → Order → Payment
- Tests all 3 services with connection pools
- **Target**: <250ms p95 end-to-end (vs ~1560ms without pool)

**Load profile:**
- 9 stages, 14 minutes total
- Max 150 concurrent users
- Stress test + spike test

---

## 📦 Prerequisites

### 1. Install K6

**Windows (using Chocolatey):**
```powershell
choco install k6
```

**macOS:**
```bash
brew install k6
```

**Linux:**
```bash
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Docker:**
```bash
docker pull grafana/k6:latest
```

### 2. Start Services

Make sure all services are running:

```bash
# Start infrastructure
docker-compose up -d

# Or start services individually
cd services/api-gateway && go run cmd/main.go
cd services/order-service && go run cmd/main.go
cd services/payment-service && go run cmd/main.go
# ... etc
```

### 3. Seed Test Data (Optional)

Create test users and products:

```bash
# Create test user
curl -X POST http://localhost:8000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "loadtest@example.com",
    "password": "testpassword123",
    "full_name": "Load Test User"
  }'

# Verify products exist
curl http://localhost:8000/products
```

---

## Running Tests

### Quick Start (Recommended)

Run the comprehensive end-to-end test:

```bash
cd tests/load/k6
k6 run end-to-end-test.js
```

### With Custom Configuration

```bash
# Custom base URL
k6 run --env BASE_URL=http://localhost:8000 end-to-end-test.js

# Enable detailed logs
k6 run --env DETAILED_LOGS=true end-to-end-test.js

# Save results to JSON
k6 run --out json=results.json end-to-end-test.js

# Run with Grafana Cloud (for visualization)
k6 run --out cloud end-to-end-test.js
```

### Run Individual Tests

```bash
# Test order creation only
k6 run order-creation-test.js

# Test payment processing only
k6 run payment-processing-test.js
```

### Docker Execution

```bash
docker run --rm -i \
  -e BASE_URL=http://host.docker.internal:8000 \
  -v ${PWD}:/scripts \
  grafana/k6:latest \
  run /scripts/end-to-end-test.js
```

---


## 📚 References

- [K6 Documentation](https://k6.io/docs/)
- [K6 Best Practices](https://k6.io/docs/testing-guides/test-types/)
- [Connection Pool Guide](../../../docs/GRPC_CONNECTION_POOL_GUIDE.md)
- [Phase 2 Summary](../../../docs/PHASE2_COMPLETE_SUMMARY.md)

---

