# E-commerce Microservices Deployment Guide

## Prerequisites

- Docker & Docker Compose
- Git
- Go 1.21+ (for local development)
- 2GB+ RAM, 20GB+ storage

## Quick Start (Local Development)

### 1. Clone & Setup
```bash
git clone https://github.com/your-username/ecommerce-microservices.git
cd ecommerce-microservices
cp .env.example .env
```

### 2. Start Infrastructure
```bash
# Start databases and message queue
docker-compose up -d postgres redis rabbitmq

# Check services are running
docker-compose ps
```

### 3. Run Database Migrations
```bash
make migrate-up
```

### 4. Start All Services
```bash
# Option 1: Using Docker (Recommended)
docker-compose up

# Option 2: Local development
make dev
```

### 5. Verify Deployment
```bash
# Check API Gateway
curl http://localhost:8080/health

# Check individual services
curl http://localhost:8081/health  # User Service
curl http://localhost:8082/health  # Product Service
```

## Production Deployment

### Docker Compose Production

#### 1. Environment Configuration
```bash
# Copy and edit production environment
cp .env.example .env.production

# Update critical values:
# - Database passwords
# - JWT secrets
# - API keys (Stripe, SMTP)
# - Production URLs
```

#### 2. Build & Deploy
```bash
# Build production images
make docker-build-prod

# Deploy with production config
docker-compose -f docker-compose.prod.yml up -d
```

### Cloud Deployment (AWS/GCP/Azure)

#### 1. Container Registry
```bash
# Tag images
docker tag ecommerce-api-gateway:latest your-registry/api-gateway:v1.0.0

# Push to registry
docker push your-registry/api-gateway:v1.0.0
```

#### 2. Database Setup
- **Managed PostgreSQL**: AWS RDS, Google Cloud SQL
- **Managed Redis**: AWS ElastiCache, Google MemoryStore
- **Message Queue**: AWS SQS/SNS, Google Pub/Sub

#### 3. Load Balancer
```nginx
# nginx.conf example
upstream api_gateway {
    server api-gateway-1:8080;
    server api-gateway-2:8080;
}

server {
    listen 80;
    location / {
        proxy_pass http://api_gateway;
    }
}
```

## Kubernetes Deployment

### 1. Namespace
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ecommerce
```

### 2. ConfigMap
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ecommerce-config
  namespace: ecommerce
data:
  DB_HOST: "postgres-service"
  REDIS_HOST: "redis-service"
  RABBITMQ_HOST: "rabbitmq-service"
```

### 3. Deployment Example
```yaml
# k8s/api-gateway-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: ecommerce
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: api-gateway
        image: your-registry/api-gateway:v1.0.0
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: ecommerce-config
```

### 4. Deploy to Kubernetes
```bash
kubectl apply -f k8s/
kubectl get pods -n ecommerce
```

## Monitoring Setup

### 1. Prometheus + Grafana
```bash
# Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# Access dashboards
# Grafana: http://localhost:3000 (admin/admin)
# Prometheus: http://localhost:9090
```

### 2. Health Checks
```bash
# API Gateway health
curl http://localhost:8080/health

# Service discovery
curl http://localhost:8080/services/status
```

## Security Checklist

### Production Security
- [ ] Change default passwords
- [ ] Use HTTPS/TLS certificates
- [ ] Configure firewall rules
- [ ] Set up VPN/private networks
- [ ] Enable database encryption
- [ ] Configure JWT with strong secrets
- [ ] Set up rate limiting
- [ ] Enable CORS properly

### Environment Variables
```bash
# Critical secrets to change:
JWT_SECRET=your-256-bit-secret
DB_PASSWORD=strong-database-password
STRIPE_SECRET_KEY=sk_live_your_live_key
SMTP_PASSWORD=your-smtp-app-password
```

## Backup & Recovery

### Database Backup
```bash
# Automated daily backup
docker exec postgres pg_dump -U postgres ecommerce > backup-$(date +%Y%m%d).sql

# Restore from backup
docker exec -i postgres psql -U postgres -d ecommerce < backup-20241201.sql
```

### Application Backup
```bash
# Backup configuration
tar -czf config-backup.tar.gz .env docker-compose.yml

# Backup uploaded files
tar -czf uploads-backup.tar.gz uploads/
```

## Troubleshooting

### Common Issues

#### Services Not Starting
```bash
# Check logs
docker-compose logs api-gateway

# Check resource usage
docker stats

# Restart specific service
docker-compose restart user-service
```

#### Database Connection Issues
```bash
# Check database is running
docker-compose ps postgres

# Test connection
docker exec postgres psql -U postgres -c "SELECT 1;"

# Check service can connect
docker-compose exec user-service nc -zv postgres 5432
```

#### Memory Issues
```bash
# Check memory usage
free -h
docker stats --no-stream

# Optimize Docker
docker system prune -f
```

### Performance Optimization

#### Database
- Enable connection pooling
- Add database indexes
- Configure query timeout
- Monitor slow queries

#### Services
- Increase replica count
- Configure resource limits
- Enable HTTP/2
- Use CDN for static assets

## Scaling Strategies

### Horizontal Scaling
```bash
# Scale specific service
docker-compose up -d --scale user-service=3

# Kubernetes scaling
kubectl scale deployment user-service --replicas=5 -n ecommerce
```

### Load Testing
```bash
# Install k6
# Run load test
k6 run tests/load/api-test.js
```

## CI/CD Pipeline

### GitHub Actions Example
```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Build & Deploy
      run: |
        make docker-build-prod
        make deploy-production
```

### Deployment Commands
```bash
# Build all services
make build

# Run tests
make test

# Deploy to staging
make deploy-staging

# Deploy to production (with approval)
make deploy-production
```

## Maintenance

### Regular Tasks
- **Daily**: Check service health, review logs
- **Weekly**: Update dependencies, backup verification
- **Monthly**: Security patches, performance review
- **Quarterly**: Capacity planning, cost optimization

### Updates
```bash
# Rolling update
docker-compose pull
docker-compose up -d --no-deps service-name

# Database migration
make migrate-up

# Zero-downtime deployment
kubectl rolling-update api-gateway --image=new-image:v2.0.0
```

---

**This deployment guide provides a complete path from development to production for the e-commerce microservices platform.**