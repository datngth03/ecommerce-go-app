# 📊 Grafana Monitoring Setup - E-Commerce Platform

Comprehensive monitoring dashboards for the E-Commerce microservices platform.

## 🎯 Available Dashboards

### 1. **All Services Overview**
**File:** `all-services-overview.json`

Unified dashboard showing metrics across all microservices in one view.

**Panels:**
- HTTP Requests Per Second (all services)
- Success Rate gauges
- Request Latency P95 comparison
- gRPC Requests Per Second
- Database Query Duration
- Business Metrics (Orders, Payments, Notifications)

**Use Cases:**
- System-wide health monitoring
- Cross-service performance comparison
- Quick incident detection
- Executive dashboards

---

### 2. **Order Service Dashboard**
**File:** `order-service.json`

Deep-dive into order processing and cart operations.

**Key Metrics:**
- `order_service_orders_total{status}` - Orders by status
- `order_service_grpc_requests_total{method}` - gRPC call rates
- `order_service_cart_operations_total{operation}` - Cart activities
- `order_service_grpc_request_duration_seconds` - Latency tracking

**Panels:**
- Orders by Status (pending/processing/completed/cancelled/failed)
- gRPC Method Call Rate (CreateOrder, GetOrder, UpdateOrderStatus, etc.)
- Method Latency (P95, P99)
- Cart Operations Rate
- Overall Success Rate
- Database Query Duration by Operation

---

### 3. **Payment Service Dashboard**
**File:** `payment-service.json`

Financial transaction monitoring with multi-currency support.

**Key Metrics:**
- `payment_service_payments_total{status,method,currency}` - Payment transactions
- `payment_service_payment_amount_total{currency}` - Payment volumes
- `payment_service_payment_duration_seconds` - Processing time
- `payment_service_refunds_total` - Refund tracking

**Panels:**
- Payment Transactions by Status
- Payment Amount by Currency (USD, EUR, GBP, VND)
- Payments by Method (credit card, PayPal, bank transfer, wallet)
- Payment Processing Duration (P95, P99, Average)
- Payment Success Rate Gauge
- Refunds Statistics (24h)
- gRPC Method Call Rates

---

### 4. **Notification Service Dashboard**
**File:** `notification-service.json`

Multi-channel notification delivery tracking.

**Key Metrics:**
- `notification_service_emails_sent_total{status}` - Email delivery
- `notification_service_sms_sent_total{status}` - SMS delivery
- `notification_service_push_notifications_sent_total{status}` - Push notifications
- `notification_service_queue_size` - Queue backlog
- `notification_service_notification_duration_seconds` - Delivery latency

**Panels:**
- Notifications Sent by Channel (Email/SMS/Push)
- Notification Queue Size Gauge
- Email Success Rate
- Notification Delivery Latency (P95, P99)
- gRPC Method Call Rates
- Total Notifications Sent (24h)
- Failed Notifications Counter

---

### 5. **Inventory Service Dashboard**
**File:** `inventory-service.json`

Stock management and reservation monitoring.

---

### 6. **Services Overview**
**File:** `services-overview.json`

High-level service health overview.

---

## 🚀 Quick Start

### Using Docker Compose

```bash
# Start monitoring stack (Prometheus + Grafana)
docker-compose -f docker-compose-monitoring.yaml up -d

# Access Grafana
open http://localhost:3000
# Username: admin
# Password: admin (change on first login)
```

### Verify Setup

```bash
# Check Grafana is running
curl http://localhost:3000/api/health

# Check Prometheus data source
curl http://localhost:9090/api/v1/targets

# Verify service metrics
curl http://localhost:8001/metrics  # user-service
curl http://localhost:8002/metrics  # product-service
curl http://localhost:8003/metrics  # order-service
curl http://localhost:8005/metrics  # payment-service
curl http://localhost:8004/metrics  # notification-service
```

---

## 📁 Directory Structure

```
infrastructure/monitoring/grafana/
├── dashboards/                          # Dashboard JSON definitions
│   ├── all-services-overview.json       # All services in one view
│   ├── order-service.json               # Order & cart metrics
│   ├── payment-service.json             # Payment transactions
│   ├── notification-service.json        # Notification delivery
│   ├── inventory-service.json           # Inventory & stock
│   └── services-overview.json           # Service health overview
├── provisioning/                        # Auto-provisioning configs
│   ├── dashboards/
│   │   └── dashboard.yml               # Dashboard provider
│   └── datasources/
│       └── prometheus.yml               # Prometheus datasource
└── README.md                            # This file
```

---

## 📈 Key Performance Indicators

### Service Health Metrics

```promql
# Service Uptime
up{job=~".*-service"}

# Overall Success Rate
sum(rate(http_requests_total{status=~"2.."}[5m])) 
/ sum(rate(http_requests_total[5m]))

# Error Rate
sum(rate(http_requests_total{status=~"5.."}[5m])) 
/ sum(rate(http_requests_total[5m]))
```

### Performance Metrics

```promql
# P95 Latency
histogram_quantile(0.95, 
  rate(http_request_duration_seconds_bucket[5m]))

# P99 Latency
histogram_quantile(0.99, 
  rate(http_request_duration_seconds_bucket[5m]))

# Requests Per Second
rate(http_requests_total[5m])
```

### Business Metrics

```promql
# Order Creation Rate
rate(order_service_orders_total[5m])

# Payment Success Rate
sum(rate(payment_service_payments_total{status="success"}[5m])) 
/ sum(rate(payment_service_payments_total[5m]))

# Email Delivery Rate
rate(notification_service_emails_sent_total[5m])

# Total Payment Volume (USD)
rate(payment_service_payment_amount_total{currency="USD"}[5m])
```

---

## 🔧 Configuration

### Auto-Provisioning Dashboards

Grafana automatically loads dashboards from `dashboards/` directory:

```yaml
# provisioning/dashboards/dashboard.yml
apiVersion: 1
providers:
  - name: 'E-Commerce'
    folder: 'E-Commerce'
    type: file
    options:
      path: /etc/grafana/provisioning/dashboards
      foldersFromFilesStructure: true
```

### Prometheus Data Source

```yaml
# provisioning/datasources/prometheus.yml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    jsonData:
      timeInterval: "5s"
```

---

## 🎨 Creating Custom Dashboards

### Method 1: Grafana UI

1. Navigate to Dashboards → New Dashboard
2. Add panels with PromQL queries
3. Configure visualizations
4. Save dashboard
5. Export JSON: Settings → JSON Model
6. Save to `dashboards/` directory

### Method 2: JSON Template

```json
{
  "title": "My Custom Dashboard",
  "uid": "custom-dashboard-uid",
  "tags": ["ecommerce", "custom"],
  "refresh": "10s",
  "panels": [
    {
      "title": "Request Rate",
      "targets": [{
        "expr": "rate(my_service_requests_total[5m])"
      }]
    }
  ]
}
```

---

## 🔔 Alerting Integration

While Grafana supports alerting, we recommend Prometheus Alertmanager for production:

**Alert Rules Location:** `infrastructure/monitoring/prometheus/alert_rules.yml`

**Key Alerts:**
- Service Down
- High Error Rate (>5%)
- High Latency (P95 >1s)
- Payment Failure Spike (>10%)
- Email Delivery Failure (>10%)
- Notification Queue Backlog (>1000)
- High CPU/Memory Usage
- Low Disk Space

---

## 🛠️ Troubleshooting

### No Data in Dashboards

**Check Prometheus Targets:**
```bash
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

**Verify Service Metrics:**
```bash
# Test each service endpoint
for port in 8001 8002 8003 8004 8005 8007; do
  echo "Testing port $port..."
  curl -s http://localhost:$port/metrics | grep -c "^[a-z]" || echo "FAILED"
done
```

**Check Metric Names:**
```bash
# List all available metrics
curl -s http://localhost:9090/api/v1/label/__name__/values | jq -r '.data[]' | grep service
```

### Dashboard Performance Issues

1. **Reduce Time Range:** Use shorter time windows (1h instead of 24h)
2. **Increase Interval:** Use `$__rate_interval` instead of fixed `[5m]`
3. **Use Recording Rules:** Pre-calculate complex queries in Prometheus
4. **Limit Series:** Add more specific label filters

Example optimization:
```promql
# Before (slow)
sum(rate(http_requests_total[5m])) by (service, method, status)

# After (faster)
sum(rate(http_requests_total[5m])) by (service)
```

### Grafana Not Starting

```bash
# Check logs
docker-compose -f docker-compose-monitoring.yaml logs grafana

# Verify permissions
ls -la infrastructure/monitoring/grafana/provisioning/

# Restart Grafana
docker-compose -f docker-compose-monitoring.yaml restart grafana
```

---

## 📚 Additional Resources

- **Grafana Docs:** https://grafana.com/docs/
- **Prometheus Query Guide:** https://prometheus.io/docs/prometheus/latest/querying/basics/
- **PromQL Examples:** https://prometheus.io/docs/prometheus/latest/querying/examples/
- **Dashboard Best Practices:** https://grafana.com/docs/grafana/latest/best-practices/

---


