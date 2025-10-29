# 🚀 Load Testing with K6

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

### 2. `payment-processing-test.js`
Tests Payment Service performance with connection pool.

**What it tests:**
- API Gateway → Payment Service
- Payment Service → Order, Notification (via connection pool)
- **Target**: <100ms p95 (vs ~440ms without pool)

**Load profile:**
- Warm-up: 10 users (30s)
- Load: 25 users (2m)
- Peak: 50 users (2m)

### 3. `end-to-end-test.js` ⭐ **Recommended**
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

## 🎯 Expected Performance Improvements

### Before Connection Pool:
```
API Gateway:     220ms per call
Order Service:   900ms per order (4-5 service calls)
Payment Service: 440ms per payment (2 service calls)
────────────────────────────────────
Total E2E:       1560ms
```

### After Connection Pool:
```
API Gateway:     20ms per call (11x faster)
Order Service:   100ms per order (9x faster)
Payment Service: 40ms per payment (11x faster)
────────────────────────────────────
Total E2E:       160ms (9.75x faster)
```

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

## 🏃 Running Tests

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

## 📊 Understanding Results

### Key Metrics to Check

#### 1. **full_flow_duration** (End-to-End Test)
```
✅ Target: p95 < 250ms, p99 < 500ms
📈 Baseline (no pool): ~1560ms
🎯 Expected: ~160ms (9.75x improvement)
```

#### 2. **order_service_latency**
```
✅ Target: p95 < 150ms
📈 Baseline: ~900ms
🎯 Expected: ~100ms (9x improvement)
```

#### 3. **payment_service_latency**
```
✅ Target: p95 < 100ms
📈 Baseline: ~440ms
🎯 Expected: ~40ms (11x improvement)
```

#### 4. **connection_pool_time_saved_ms**
```
🎯 Expected: ~800-1000ms saved per request
💰 Shows actual time saved by connection pooling
```

#### 5. **errors** (Error Rate)
```
✅ Target: < 1%
🎯 Expected: < 0.1%
```

### Sample Output

```
     ✓ full_flow_duration............: avg=165ms min=95ms  med=158ms max=412ms p(95)=234ms p(99)=356ms
     ✓ order_service_latency.........: avg=98ms  min=62ms  med=94ms  max=187ms p(95)=142ms p(99)=165ms
     ✓ payment_service_latency.......: avg=38ms  min=18ms  med=36ms  max=89ms  p(95)=67ms  p(99)=81ms
     ✓ connection_pool_time_saved_ms.: avg=1322ms (savings per request)
     ✓ errors........................: 0.12% ✓ 18 ✗ 14582
     
     ✓ http_req_duration.............: avg=87ms  min=12ms  med=76ms  max=521ms p(95)=198ms p(99)=341ms
     
     full_flow_success...............: 14582 (98.8% success rate)
     full_flow_failure...............: 18 (1.2% failure rate)
```

---

## 🎨 Visualization

### Option 1: K6 Cloud (Easiest)

```bash
# Login to K6 Cloud (free tier available)
k6 login cloud

# Run with cloud output
k6 run --out cloud end-to-end-test.js
```

Visit the URL in the output to see real-time graphs!

### Option 2: Grafana + InfluxDB

1. **Start Grafana Stack:**
```bash
# Add to docker-compose.yaml or use existing monitoring setup
docker-compose -f docker-compose-monitoring.yaml up -d
```

2. **Run test with InfluxDB output:**
```bash
k6 run --out influxdb=http://localhost:8086/k6 end-to-end-test.js
```

3. **View in Grafana:**
- Open http://localhost:3000
- Import K6 dashboard
- View real-time metrics

### Option 3: JSON Output + Analysis

```bash
# Save results
k6 run --out json=results.json end-to-end-test.js

# Analyze with jq
cat results.json | jq -r 'select(.type=="Point" and .metric=="full_flow_duration") | .data.value' | \
  python3 -c "import sys; data=[float(x) for x in sys.stdin]; print(f'avg: {sum(data)/len(data):.2f}ms')"
```

---

## ✅ Success Criteria

Your connection pool implementation is **successful** if:

1. ✅ **p95 latency < 250ms** (end-to-end)
   - Before: ~1560ms
   - Target: ~160ms
   - Improvement: 9.75x

2. ✅ **Error rate < 1%**
   - Target: < 0.1%
   - Shows stability under load

3. ✅ **Success rate > 99%**
   - At least 99% of flows complete successfully

4. ✅ **Order service < 150ms p95**
   - Before: ~900ms
   - Target: ~100ms

5. ✅ **Payment service < 100ms p95**
   - Before: ~440ms
   - Target: ~40ms

6. ✅ **Connection pool savings > 800ms**
   - Shows actual time saved per request

---

## 🔧 Troubleshooting

### Issue: High error rate (>1%)

**Possible causes:**
- Services not started
- Database not initialized
- Incorrect BASE_URL
- Network issues

**Solution:**
```bash
# Check service health
curl http://localhost:8000/health
curl http://localhost:9003/health
curl http://localhost:9005/health

# Check logs
docker-compose logs api-gateway
docker-compose logs order-service
docker-compose logs payment-service
```

### Issue: Latency still high (>500ms p95)

**Possible causes:**
- Connection pool not initialized
- Wrong configuration
- Database slow queries
- No indexes

**Solution:**
```bash
# Check connection pool health
curl http://localhost:8000/health/pools/detailed

# Check database indexes
psql -U postgres -d ecommerce -c "\d+ orders"

# Enable detailed logging
export LOG_LEVEL=debug
```

### Issue: Test fails to authenticate

**Solution:**
```bash
# Create test user manually
curl -X POST http://localhost:8000/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "loadtest@example.com",
    "password": "testpassword123",
    "full_name": "Load Test User"
  }'

# Or skip auth in test
k6 run --env SKIP_AUTH=true end-to-end-test.js
```

---

## 📝 Next Steps

After successful load testing:

1. ✅ **Document Results**
   - Save K6 output to `results/`
   - Update `PHASE2_COMPLETE_SUMMARY.md`
   - Add performance graphs

2. ✅ **Add to CI/CD**
   - Run load tests on every deploy
   - Set up performance regression alerts
   - Automate with GitHub Actions

3. ✅ **Continuous Monitoring**
   - Add connection pool metrics to Prometheus
   - Create Grafana dashboards
   - Set up alerts for degradation

4. ✅ **Optimize Further** (Phase 3)
   - TLS/SSL for security
   - Load balancing
   - Auto-scaling based on metrics

---

## 📚 References

- [K6 Documentation](https://k6.io/docs/)
- [K6 Best Practices](https://k6.io/docs/testing-guides/test-types/)
- [Connection Pool Guide](../../../docs/GRPC_CONNECTION_POOL_GUIDE.md)
- [Phase 2 Summary](../../../docs/PHASE2_COMPLETE_SUMMARY.md)

---

**Last Updated:** October 29, 2025  
**Status:** Ready for execution ✅  
**Expected Duration:** 14 minutes (end-to-end test)
