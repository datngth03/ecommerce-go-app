# E-Commerce Microservices - Deployment Guide# E-commerce Microservices Deployment Guide



## Table of Contents## Prerequisites

1. [Overview](#overview)

2. [Prerequisites](#prerequisites)- Docker & Docker Compose

3. [Environment Setup](#environment-setup)- Git

4. [Development Deployment](#development-deployment)- Go 1.21+ (for local development)

5. [Production Deployment](#production-deployment)- 2GB+ RAM, 20GB+ storage

6. [Database Management](#database-management)

7. [Monitoring & Health Checks](#monitoring--health-checks)## Quick Start (Local Development)

8. [Backup & Recovery](#backup--recovery)

9. [Scaling & Load Balancing](#scaling--load-balancing)### 1. Clone & Setup

10. [Troubleshooting](#troubleshooting)```bash

11. [Rollback Procedures](#rollback-procedures)git clone https://github.com/your-username/ecommerce-microservices.git

cd ecommerce-microservices

---cp .env.example .env

```

## Overview

### 2. Start Infrastructure

This guide covers the deployment of the E-Commerce microservices platform across development, staging, and production environments using Docker, Docker Compose, and Kubernetes.```bash

# Start databases and message queue

**Architecture**:docker-compose up -d postgres redis rabbitmq

- **6 Microservices**: user-service, product-service, inventory-service, order-service, payment-service, notification-service

- **1 API Gateway**: Centralized entry point for all client requests# Check services are running

- **Database**: PostgreSQL 15 (separate database per service)docker-compose ps

- **Cache**: Redis 7```

- **Message Queue**: RabbitMQ (for asynchronous communication)

- **Monitoring**: Prometheus + Grafana### 3. Run Database Migrations

```bash

---make migrate-up

```

## Prerequisites

### 4. Start All Services

### Software Requirements```bash

# Option 1: Using Docker (Recommended)

| Component | Version | Purpose |docker-compose up

|-----------|---------|---------|

| Docker | 24.0+ | Container runtime |# Option 2: Local development

| Docker Compose | 2.20+ | Multi-container orchestration |make dev

| Go | 1.21+ | Service compilation |```

| PostgreSQL | 15+ | Database (if not using Docker) |

| Redis | 7+ | Caching layer |### 5. Verify Deployment

| Make | 4.0+ | Build automation |```bash

| Git | 2.40+ | Version control |# Check API Gateway

curl http://localhost:8080/health

### Hardware Requirements

# Check individual services

#### Development Environmentcurl http://localhost:8081/health  # User Service

- **CPU**: 4 cores minimumcurl http://localhost:8082/health  # Product Service

- **RAM**: 8GB minimum, 16GB recommended```

- **Storage**: 20GB free space

- **Network**: Stable internet connection## Production Deployment



#### Production Environment### Docker Compose Production

- **CPU**: 8 cores minimum per node

- **RAM**: 32GB minimum per node#### 1. Environment Configuration

- **Storage**: 200GB SSD minimum```bash

- **Network**: 1Gbps minimum bandwidth# Copy and edit production environment

cp .env.example .env.production

### Network Ports

# Update critical values:

Ensure the following ports are available:# - Database passwords

# - JWT secrets

| Port | Service | Protocol |# - API keys (Stripe, SMTP)

|------|---------|----------|# - Production URLs

| 8000 | API Gateway | HTTP |```

| 9001 | User Service | gRPC |

| 9002 | Product Service | gRPC |#### 2. Build & Deploy

| 9003 | Inventory Service | gRPC |```bash

| 9004 | Order Service | gRPC |# Build production images

| 9005 | Payment Service | gRPC |make docker-build-prod

| 9006 | Notification Service | gRPC |

| 5432 | PostgreSQL | TCP |# Deploy with production config

| 6379 | Redis | TCP |docker-compose -f docker-compose.prod.yml up -d

| 5672 | RabbitMQ | AMQP |```

| 15672 | RabbitMQ Management | HTTP |

| 9090 | Prometheus | HTTP |### Cloud Deployment (AWS/GCP/Azure)

| 3000 | Grafana | HTTP |

#### 1. Container Registry

---```bash

# Tag images

## Environment Setupdocker tag ecommerce-api-gateway:latest your-registry/api-gateway:v1.0.0



### 1. Clone Repository# Push to registry

docker push your-registry/api-gateway:v1.0.0

```bash```

git clone https://github.com/yourusername/ecommerce-go-app.git

cd ecommerce-go-app#### 2. Database Setup

```- **Managed PostgreSQL**: AWS RDS, Google Cloud SQL

- **Managed Redis**: AWS ElastiCache, Google MemoryStore

### 2. Environment Variables- **Message Queue**: AWS SQS/SNS, Google Pub/Sub



Create `.env` files for each service. A template is provided:#### 3. Load Balancer

```nginx

```bash# nginx.conf example

# Copy templateupstream api_gateway {

cp .env.example .env    server api-gateway-1:8080;

    server api-gateway-2:8080;

# Edit with your values}

nano .env

```server {

    listen 80;

#### Required Environment Variables    location / {

        proxy_pass http://api_gateway;

```bash    }

# Database Configuration}

POSTGRES_HOST=postgres```

POSTGRES_PORT=5432

POSTGRES_USER=ecommerce_user## Kubernetes Deployment

POSTGRES_PASSWORD=your_secure_password

POSTGRES_DB=ecommerce_db### 1. Namespace

```yaml

# Redis Configuration# k8s/namespace.yaml

REDIS_HOST=redisapiVersion: v1

REDIS_PORT=6379kind: Namespace

REDIS_PASSWORD=your_redis_passwordmetadata:

  name: ecommerce

# JWT Configuration```

JWT_SECRET=your_jwt_secret_key_at_least_32_characters

JWT_EXPIRATION=24h### 2. ConfigMap

REFRESH_TOKEN_EXPIRATION=168h```yaml

# k8s/configmap.yaml

# Service URLs (for inter-service communication)apiVersion: v1

USER_SERVICE_URL=user-service:9001kind: ConfigMap

PRODUCT_SERVICE_URL=product-service:9002metadata:

INVENTORY_SERVICE_URL=inventory-service:9003  name: ecommerce-config

ORDER_SERVICE_URL=order-service:9004  namespace: ecommerce

PAYMENT_SERVICE_URL=payment-service:9005data:

NOTIFICATION_SERVICE_URL=notification-service:9006  DB_HOST: "postgres-service"

  REDIS_HOST: "redis-service"

# External Services  RABBITMQ_HOST: "rabbitmq-service"

STRIPE_API_KEY=sk_test_your_stripe_key```

STRIPE_WEBHOOK_SECRET=whsec_your_webhook_secret

SENDGRID_API_KEY=SG.your_sendgrid_key### 3. Deployment Example

EMAIL_FROM=noreply@yourdomain.com```yaml

# k8s/api-gateway-deployment.yaml

# RabbitMQapiVersion: apps/v1

RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/kind: Deployment

RABBITMQ_EXCHANGE=ecommerce_eventsmetadata:

  name: api-gateway

# Monitoring  namespace: ecommerce

PROMETHEUS_ENABLED=truespec:

METRICS_PORT=8080  replicas: 3

  selector:

# Logging    matchLabels:

LOG_LEVEL=info      app: api-gateway

LOG_FORMAT=json  template:

    metadata:

# Environment      labels:

ENVIRONMENT=development        app: api-gateway

```    spec:

      containers:

### 3. Generate Secrets      - name: api-gateway

        image: your-registry/api-gateway:v1.0.0

Generate secure secrets for production:        ports:

        - containerPort: 8080

```bash        envFrom:

# JWT Secret (32 bytes)        - configMapRef:

openssl rand -base64 32            name: ecommerce-config

```

# Database Password

openssl rand -base64 24### 4. Deploy to Kubernetes

```bash

# Redis Passwordkubectl apply -f k8s/

openssl rand -base64 16kubectl get pods -n ecommerce

``````



---## Monitoring Setup



## Development Deployment### 1. Prometheus + Grafana

```bash

### Quick Start (Docker Compose)# Start monitoring stack

docker-compose -f docker-compose.monitoring.yml up -d

#### 1. Build Services

# Access dashboards

```bash# Grafana: http://localhost:3000 (admin/admin)

# Build all services# Prometheus: http://localhost:9090

make build```



# Or build specific service### 2. Health Checks

make build-service SERVICE=user-service```bash

```# API Gateway health

curl http://localhost:8080/health

#### 2. Start Infrastructure

# Service discovery

```bashcurl http://localhost:8080/services/status

# Start databases and dependencies first```

docker-compose up -d postgres redis rabbitmq

## Security Checklist

# Wait for services to be healthy

docker-compose ps### Production Security

```- [ ] Change default passwords

- [ ] Use HTTPS/TLS certificates

#### 3. Run Database Migrations- [ ] Configure firewall rules

- [ ] Set up VPN/private networks

```bash- [ ] Enable database encryption

# Run migrations for all services- [ ] Configure JWT with strong secrets

make migrate-up- [ ] Set up rate limiting

- [ ] Enable CORS properly

# Or migrate specific service

docker-compose run --rm migrations-user up### Environment Variables

``````bash

# Critical secrets to change:

#### 4. Start All ServicesJWT_SECRET=your-256-bit-secret

DB_PASSWORD=strong-database-password

```bashSTRIPE_SECRET_KEY=sk_live_your_live_key

# Start all microservices and API gatewaySMTP_PASSWORD=your-smtp-app-password

docker-compose up -d```



# View logs## Backup & Recovery

docker-compose logs -f

### Database Backup

# View specific service logs```bash

docker-compose logs -f api-gateway# Automated daily backup

```docker exec postgres pg_dump -U postgres ecommerce > backup-$(date +%Y%m%d).sql



#### 5. Verify Deployment# Restore from backup

docker exec -i postgres psql -U postgres -d ecommerce < backup-20241201.sql

```bash```

# Check service health

curl http://localhost:8000/health### Application Backup

```bash

# Check individual service# Backup configuration

curl http://localhost:8000/api/v1/users/healthtar -czf config-backup.tar.gz .env docker-compose.yml

```

# Backup uploaded files

### Service-by-Service Startuptar -czf uploads-backup.tar.gz uploads/

```

If you prefer granular control:

## Troubleshooting

```bash

# 1. Start infrastructure### Common Issues

docker-compose up -d postgres redis rabbitmq

#### Services Not Starting

# 2. Start core services```bash

docker-compose up -d user-service product-service inventory-service# Check logs

docker-compose logs api-gateway

# 3. Start business services

docker-compose up -d order-service payment-service notification-service# Check resource usage

docker stats

# 4. Start API gateway

docker-compose up -d api-gateway# Restart specific service

docker-compose restart user-service

# 5. Start monitoring```

docker-compose up -d prometheus grafana

```#### Database Connection Issues

```bash

### Development Workflow# Check database is running

docker-compose ps postgres

```bash

# Watch logs in development# Test connection

docker-compose logs -fdocker exec postgres psql -U postgres -c "SELECT 1;"



# Restart a service after code changes# Check service can connect

docker-compose restart user-servicedocker-compose exec user-service nc -zv postgres 5432

```

# Rebuild and restart

docker-compose up -d --build user-service#### Memory Issues

```bash

# Stop all services# Check memory usage

docker-compose downfree -h

docker stats --no-stream

# Stop and remove volumes (fresh start)

docker-compose down -v# Optimize Docker

```docker system prune -f

```

---

### Performance Optimization

## Production Deployment

#### Database

### Option 1: Docker Compose (Small-Scale)- Enable connection pooling

- Add database indexes

#### 1. Production Configuration- Configure query timeout

- Monitor slow queries

Create `docker-compose.prod.yaml`:

#### Services

```yaml- Increase replica count

version: '3.8'- Configure resource limits

- Enable HTTP/2

services:- Use CDN for static assets

  # Use production-optimized images

  api-gateway:## Scaling Strategies

    image: yourdockerhub/ecommerce-api-gateway:v1.0.0

    deploy:### Horizontal Scaling

      replicas: 3```bash

      restart_policy:# Scale specific service

        condition: on-failuredocker-compose up -d --scale user-service=3

        max_attempts: 3

    environment:# Kubernetes scaling

      ENVIRONMENT: productionkubectl scale deployment user-service --replicas=5 -n ecommerce

      LOG_LEVEL: warn```

    healthcheck:

      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]### Load Testing

      interval: 30s```bash

      timeout: 10s# Install k6

      retries: 3# Run load test

      start_period: 40sk6 run tests/load/api-test.js

```

  # Additional services...

```## CI/CD Pipeline



#### 2. Deploy to Production### GitHub Actions Example

```yaml

```bash# .github/workflows/deploy.yml

# Pull latest imagesname: Deploy

docker-compose -f docker-compose.prod.yaml pullon:

  push:

# Deploy    branches: [main]

docker-compose -f docker-compose.prod.yaml up -djobs:

  deploy:

# Verify    runs-on: ubuntu-latest

docker-compose -f docker-compose.prod.yaml ps    steps:

```    - uses: actions/checkout@v2

    - name: Build & Deploy

### Option 2: Kubernetes (Recommended for Scale)      run: |

        make docker-build-prod

#### 1. Prerequisites        make deploy-production

```

```bash

# Install kubectl### Deployment Commands

curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"```bash

sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl# Build all services

make build

# Verify connection to cluster

kubectl cluster-info# Run tests

```make test



#### 2. Create Namespace# Deploy to staging

make deploy-staging

```bash

kubectl create namespace ecommerce-prod# Deploy to production (with approval)

kubectl config set-context --current --namespace=ecommerce-prodmake deploy-production

``````



#### 3. Configure Secrets## Maintenance



```bash### Regular Tasks

# Create database secret- **Daily**: Check service health, review logs

kubectl create secret generic postgres-secret \- **Weekly**: Update dependencies, backup verification

  --from-literal=username=ecommerce_user \- **Monthly**: Security patches, performance review

  --from-literal=password=your_secure_password- **Quarterly**: Capacity planning, cost optimization



# Create JWT secret### Updates

kubectl create secret generic jwt-secret \```bash

  --from-literal=secret=your_jwt_secret_key# Rolling update

docker-compose pull

# Create Stripe secretdocker-compose up -d --no-deps service-name

kubectl create secret generic stripe-secret \

  --from-literal=api-key=sk_live_your_stripe_key \# Database migration

  --from-literal=webhook-secret=whsec_your_webhook_secretmake migrate-up

```

# Zero-downtime deployment

#### 4. Deploy Infrastructurekubectl rolling-update api-gateway --image=new-image:v2.0.0

```

```bash

# Deploy PostgreSQL StatefulSet---

kubectl apply -f infrastructure/k8s/postgres/

**This deployment guide provides a complete path from development to production for the e-commerce microservices platform.**
# Deploy Redis
kubectl apply -f infrastructure/k8s/redis/

# Deploy RabbitMQ
kubectl apply -f infrastructure/k8s/rabbitmq/
```

#### 5. Deploy Services

```bash
# Deploy all microservices
kubectl apply -f infrastructure/k8s/services/

# Deploy API Gateway
kubectl apply -f infrastructure/k8s/api-gateway/

# Deploy Ingress
kubectl apply -f infrastructure/k8s/ingress/
```

#### 6. Verify Deployment

```bash
# Check pod status
kubectl get pods

# Check services
kubectl get services

# Check ingress
kubectl get ingress

# View logs
kubectl logs -l app=api-gateway --tail=100 -f
```

### Load Balancer Configuration

#### Using NGINX Ingress Controller

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ecommerce-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.yourdomain.com
      secretName: ecommerce-tls
  rules:
    - host: api.yourdomain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: api-gateway
                port:
                  number: 8000
```

---

## Database Management

### Database Initialization

#### 1. Create Databases

Each service requires its own database:

```sql
-- Connect to PostgreSQL
psql -h localhost -U postgres

-- Create databases
CREATE DATABASE users_db;
CREATE DATABASE products_db;
CREATE DATABASE inventory_db;
CREATE DATABASE orders_db;
CREATE DATABASE payments_db;
CREATE DATABASE notifications_db;

-- Create user with permissions
CREATE USER ecommerce_user WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE users_db TO ecommerce_user;
GRANT ALL PRIVILEGES ON DATABASE products_db TO ecommerce_user;
GRANT ALL PRIVILEGES ON DATABASE inventory_db TO ecommerce_user;
GRANT ALL PRIVILEGES ON DATABASE orders_db TO ecommerce_user;
GRANT ALL PRIVILEGES ON DATABASE payments_db TO ecommerce_user;
GRANT ALL PRIVILEGES ON DATABASE notifications_db TO ecommerce_user;
```

#### 2. Run Migrations

Using the migrations tool:

```bash
# Install golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Run migrations
cd services/user-service
migrate -path migrations -database "postgresql://ecommerce_user:password@localhost:5432/users_db?sslmode=disable" up

# Repeat for each service
```

Using Docker:

```bash
# Run migrations container
docker-compose up migrations
```

### Migration Management

```bash
# Check migration status
make migrate-status SERVICE=user-service

# Rollback last migration
make migrate-down SERVICE=user-service

# Rollback to specific version
make migrate-version SERVICE=user-service VERSION=3

# Create new migration
make migrate-create SERVICE=user-service NAME=add_user_preferences
```

---

## Monitoring & Health Checks

### Health Check Endpoints

Each service exposes health check endpoints:

```bash
# API Gateway health
curl http://localhost:8000/health

# Service-specific health (through gateway)
curl http://localhost:8000/api/v1/users/health
curl http://localhost:8000/api/v1/products/health
curl http://localhost:8000/api/v1/inventory/health
curl http://localhost:8000/api/v1/orders/health
curl http://localhost:8000/api/v1/payments/health
```

### Prometheus Metrics

Access Prometheus dashboard:

```bash
# Open Prometheus UI
http://localhost:9090

# View available metrics
http://localhost:9090/metrics
```

**Key Metrics**:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `grpc_server_handled_total` - gRPC requests
- `db_connection_pool_size` - Database connections
- `redis_connected_clients` - Redis connections

### Grafana Dashboards

Access Grafana:

```bash
# Open Grafana UI
http://localhost:3000

# Default credentials
Username: admin
Password: admin
```

**Pre-configured Dashboards**:
1. **Service Overview**: Request rate, latency, error rate
2. **Database Metrics**: Connection pool, query performance
3. **Infrastructure**: CPU, memory, disk usage
4. **Business Metrics**: Orders, payments, user registrations

### Logging

Centralized logging with ELK Stack (optional):

```bash
# Start ELK stack
docker-compose -f docker-compose.monitoring.yaml up -d elasticsearch logstash kibana

# View logs in Kibana
http://localhost:5601
```

View service logs:

```bash
# Docker Compose
docker-compose logs -f --tail=100 user-service

# Kubernetes
kubectl logs -l app=user-service --tail=100 -f

# Filter by log level
kubectl logs -l app=user-service | grep ERROR
```

---

## Backup & Recovery

### Database Backup

#### Automated Backup Script

Create `scripts/backup-db.sh`:

```bash
#!/bin/bash

BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DATABASES=("users_db" "products_db" "inventory_db" "orders_db" "payments_db")

mkdir -p $BACKUP_DIR

for DB in "${DATABASES[@]}"; do
    echo "Backing up $DB..."
    docker-compose exec -T postgres pg_dump -U ecommerce_user $DB | gzip > "$BACKUP_DIR/${DB}_${TIMESTAMP}.sql.gz"
done

# Keep only last 7 days of backups
find $BACKUP_DIR -type f -mtime +7 -delete

echo "Backup completed: $TIMESTAMP"
```

#### Schedule Backups (Cron)

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /path/to/ecommerce-go-app/scripts/backup-db.sh >> /var/log/ecommerce-backup.log 2>&1
```

### Database Recovery

```bash
# Restore specific database
gunzip < /backups/postgres/users_db_20250121_020000.sql.gz | docker-compose exec -T postgres psql -U ecommerce_user users_db

# Restore all databases
for backup in /backups/postgres/*_20250121_020000.sql.gz; do
    DB=$(basename $backup | cut -d'_' -f1)
    echo "Restoring $DB..."
    gunzip < $backup | docker-compose exec -T postgres psql -U ecommerce_user $DB
done
```

### Application State Backup

```bash
# Backup Redis data
docker-compose exec redis redis-cli BGSAVE
docker cp ecommerce-redis:/data/dump.rdb /backups/redis/dump_$(date +"%Y%m%d").rdb

# Backup RabbitMQ definitions
docker-compose exec rabbitmq rabbitmqctl export_definitions /tmp/definitions.json
docker cp ecommerce-rabbitmq:/tmp/definitions.json /backups/rabbitmq/definitions_$(date +"%Y%m%d").json
```

---

## Scaling & Load Balancing

### Horizontal Scaling with Docker Compose

```bash
# Scale specific service
docker-compose up -d --scale user-service=3 --scale order-service=3

# Scale all services
docker-compose up -d --scale api-gateway=2 --scale user-service=3 --scale product-service=3
```

### Horizontal Scaling with Kubernetes

```bash
# Scale deployment
kubectl scale deployment user-service --replicas=5

# Auto-scaling based on CPU
kubectl autoscale deployment user-service --min=2 --max=10 --cpu-percent=70

# Check autoscaler status
kubectl get hpa
```

### Database Connection Pooling

Configure connection pooling in each service:

```go
// config/database.go
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, _ := db.DB()

// Set connection pool limits
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### Redis Caching Strategy

Implement caching for frequently accessed data:

```go
// Example: Cache product details
func GetProduct(id string) (*Product, error) {
    // Try cache first
    cached, err := redis.Get(ctx, "product:"+id)
    if err == nil {
        var product Product
        json.Unmarshal([]byte(cached), &product)
        return &product, nil
    }
    
    // Fetch from database
    product, err := db.GetProduct(id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache with TTL
    data, _ := json.Marshal(product)
    redis.Set(ctx, "product:"+id, data, 15*time.Minute)
    
    return product, nil
}
```

---

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

**Symptoms**: Service crashes immediately or doesn't respond

**Diagnosis**:
```bash
# Check logs
docker-compose logs user-service

# Check if port is available
netstat -tuln | grep 9001

# Check environment variables
docker-compose exec user-service env
```

**Solutions**:
- Verify environment variables are set correctly
- Ensure database is accessible
- Check port conflicts
- Verify dependencies are running

#### 2. Database Connection Failures

**Symptoms**: "connection refused" or "too many connections"

**Diagnosis**:
```bash
# Check PostgreSQL status
docker-compose exec postgres pg_isready

# Check active connections
docker-compose exec postgres psql -U ecommerce_user -c "SELECT count(*) FROM pg_stat_activity;"

# Check connection limits
docker-compose exec postgres psql -U ecommerce_user -c "SHOW max_connections;"
```

**Solutions**:
```bash
# Increase max connections (postgresql.conf)
max_connections = 200

# Adjust connection pool in services
SetMaxOpenConns(50)  # per service

# Restart PostgreSQL
docker-compose restart postgres
```

#### 3. High Memory Usage

**Diagnosis**:
```bash
# Check container memory usage
docker stats

# Check service memory usage
docker-compose exec user-service ps aux

# Kubernetes
kubectl top pods
```

**Solutions**:
- Set memory limits in docker-compose.yaml or K8s manifests
- Optimize database queries
- Implement pagination for large result sets
- Increase cache TTL to reduce database hits

#### 4. Slow API Response

**Diagnosis**:
```bash
# Check response time
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8000/api/v1/products

# Check database query performance
docker-compose exec postgres psql -U ecommerce_user -d products_db -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

**Solutions**:
- Add database indexes
- Implement caching
- Optimize N+1 queries
- Use database connection pooling

#### 5. gRPC Connection Errors

**Symptoms**: "connection refused" between services

**Diagnosis**:
```bash
# Check service discovery
docker-compose exec api-gateway ping user-service

# Test gRPC connection
grpcurl -plaintext user-service:9001 list
```

**Solutions**:
- Verify service names in docker-compose.yaml
- Check network configuration
- Ensure services are healthy before connecting
- Use retry logic with exponential backoff

### Debug Mode

Enable debug logging:

```bash
# Set environment variable
export LOG_LEVEL=debug

# Or in .env file
LOG_LEVEL=debug

# Restart service
docker-compose restart user-service

# View debug logs
docker-compose logs -f user-service
```

### Performance Profiling

Enable Go profiling:

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Access profiling data:
```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

---

## Rollback Procedures

### Docker Compose Rollback

```bash
# Tag current version before deploying
docker-compose images

# Deploy new version
docker-compose up -d

# If issues arise, rollback to previous image
docker-compose stop
docker tag yourdockerhub/ecommerce-api-gateway:v1.1.0 yourdockerhub/ecommerce-api-gateway:previous
docker-compose up -d yourdockerhub/ecommerce-api-gateway:v1.0.0
```

### Kubernetes Rollback

```bash
# View deployment history
kubectl rollout history deployment/user-service

# Rollback to previous version
kubectl rollout undo deployment/user-service

# Rollback to specific revision
kubectl rollout undo deployment/user-service --to-revision=3

# Check rollback status
kubectl rollout status deployment/user-service
```

### Database Migration Rollback

```bash
# Check current version
make migrate-version SERVICE=user-service

# Rollback one version
make migrate-down SERVICE=user-service

# Rollback to specific version
migrate -path migrations -database $DATABASE_URL goto 5
```

### Emergency Rollback Checklist

1. **Notify team** - Inform all stakeholders
2. **Check logs** - Identify the issue
3. **Stop deployment** - Prevent further rollout
4. **Rollback application** - Revert to stable version
5. **Verify functionality** - Test critical paths
6. **Rollback database** (if needed) - Restore from backup
7. **Document incident** - Create postmortem
8. **Fix and redeploy** - Address root cause

---

## Security Considerations

### SSL/TLS Configuration

```yaml
# Use TLS for production
api-gateway:
  environment:
    TLS_ENABLED: "true"
    TLS_CERT_PATH: "/certs/server.crt"
    TLS_KEY_PATH: "/certs/server.key"
  volumes:
    - ./certs:/certs:ro
```

### Secrets Management

Use external secrets management:

```bash
# HashiCorp Vault
vault kv put secret/ecommerce/jwt jwt_secret="your_secret"
vault kv put secret/ecommerce/db password="your_password"

# AWS Secrets Manager
aws secretsmanager create-secret --name ecommerce/jwt --secret-string "your_secret"
```

### Network Security

```bash
# Use private networks
docker network create --internal ecommerce-internal

# Restrict service communication
docker-compose up -d --scale api-gateway=1 --scale user-service=2 --network ecommerce-internal
```

---

## Maintenance Windows

Schedule regular maintenance:

1. **Weekly**: Log rotation, temporary file cleanup
2. **Monthly**: Database vacuum/analyze, security updates
3. **Quarterly**: Dependency updates, performance review
4. **Annually**: Major version upgrades, architecture review

---

## Support & Escalation

**Level 1 Support**: Check logs, restart services  
**Level 2 Support**: Database issues, performance tuning  
**Level 3 Support**: Architecture changes, major incidents

**Emergency Contact**: devops@yourdomain.com  
**On-Call Rotation**: PagerDuty integration

---

**Version**: 2.0.0  
**Last Updated**: October 2025  
**Next Review**: January 2026
