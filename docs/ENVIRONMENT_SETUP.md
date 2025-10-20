# Environment Configuration Guide

## Overview

This project uses **per-service environment files** for better isolation and security. Each microservice has its own `.env.example` template.

## File Structure

```
📁 project-root/
├── .env.example                          # For docker-compose (all services)
├── .gitignore                            # Ignores .env files
├── services/
│   ├── user-service/
│   │   ├── .env.example                 # Template
│   │   └── .env                         # Your config (gitignored)
│   ├── product-service/
│   │   ├── .env.example
│   │   └── .env
│   ├── order-service/
│   │   ├── .env.example
│   │   └── .env
│   ├── payment-service/
│   │   ├── .env.example
│   │   └── .env
│   ├── inventory-service/
│   │   ├── .env.example
│   │   └── .env
│   ├── notification-service/
│   │   ├── .env.example
│   │   └── .env
│   └── api-gateway/
│       ├── .env.example
│       └── .env
```

## Quick Start

### For Local Development (Run services individually)

1. **Copy template for each service:**
```bash
# User Service
cd services/user-service
cp .env.example .env
# Edit .env with your values

# Product Service
cd ../product-service
cp .env.example .env
# Edit .env with your values

# Repeat for other services...
```

2. **Start infrastructure:**
```bash
docker-compose up -d postgres redis rabbitmq
```

3. **Run services:**
```bash
# Terminal 1 - User Service
cd services/user-service
go run cmd/main.go

# Terminal 2 - Product Service
cd services/product-service
go run cmd/main.go

# Terminal 3 - Order Service
cd services/order-service
go run cmd/main.go

# Terminal 4 - API Gateway
cd services/api-gateway
go run cmd/main.go
```

### For Docker Compose (Run all services together)

1. **Copy root template:**
```bash
cp .env.example .env
# Edit .env with your values
```

2. **Start all services:**
```bash
docker-compose up -d
```

## Port Convention

| Service            | HTTP Port | gRPC Port | Database         |
|-------------------|-----------|-----------|------------------|
| API Gateway       | 8000      | -         | -                |
| User Service      | 8001      | 9001      | users_db         |
| Product Service   | 8002      | 9002      | product_db       |
| Order Service     | 8003      | 9003      | orders_db        |
| Payment Service   | 8004      | 9004      | payment_db       |
| Inventory Service | 8005      | 9005      | inventory_db     |
| Notification Service | 8006   | 9006      | notification_db  |

## Environment Variables Convention

### Service-specific format:
```bash
# Without prefix (used in service's .env file)
HTTP_PORT=8001
GRPC_PORT=9001
DB_NAME=users_db
```

### Docker-compose format:
```bash
# With prefix (used in root .env file)
USER_HTTP_PORT=8001
USER_GRPC_PORT=9001
USER_DB_NAME=users_db
```

## Configuration Sections

### Common to All Services:
- `SERVICE_NAME` - Service identifier
- `SERVICE_VERSION` - Semantic version
- `ENVIRONMENT` - development/staging/production
- `HTTP_PORT` - HTTP REST API port
- `GRPC_PORT` - gRPC server port (if applicable)
- `DB_*` - PostgreSQL configuration
- `LOG_*` - Logging configuration

### Service-specific:

**User Service:**
- `REDIS_*` - Session storage
- `JWT_*` - Authentication tokens

**Order Service:**
- `REDIS_*` - Cart caching
- `RABBITMQ_*` - Event publishing
- `*_SERVICE_GRPC` - External service endpoints

**Payment Service:**
- `STRIPE_*` - Stripe payment gateway
- `PAYPAL_*` - PayPal integration
- `PAYMENT_CURRENCY` - Default currency

**Inventory Service:**
- `REDIS_*` - Stock caching
- `RABBITMQ_*` - Stock reservation events

**Notification Service:**
- `SMTP_*` - Email configuration
- `TWILIO_*` - SMS configuration
- `EMAIL_FROM_*` - Sender information

**API Gateway:**
- `RATE_LIMIT_*` - Rate limiting
- `*_SERVICE_GRPC` - All microservice endpoints
- `STRIPE_*`, `SMTP_*`, `TWILIO_*` - External APIs

## Security Best Practices

1. **Never commit `.env` files** (already in .gitignore)
2. **Use strong secrets in production:**
   - Change `JWT_SECRET` from default
   - Use real credentials for payment gateways
   - Use environment-specific SMTP credentials

3. **Rotate secrets regularly:**
   ```bash
   # Generate new JWT secret
   openssl rand -base64 32
   ```

4. **Use different credentials per environment:**
   - Development: `users_db_dev`
   - Staging: `users_db_staging`
   - Production: `users_db_prod`

## Troubleshooting

### Service can't find .env file
```bash
# Check if .env exists
ls services/user-service/.env

# If not, copy from template
cp services/user-service/.env.example services/user-service/.env
```

### Port already in use
```bash
# Change port in .env file
HTTP_PORT=8011  # Instead of 8001

# Or kill process using the port
# Windows
netstat -ano | findstr :8001
taskkill /PID <PID> /F

# Linux/Mac
lsof -ti:8001 | xargs kill -9
```

### Database connection failed
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check credentials in .env
DB_USER=postgres
DB_PASSWORD=postgres1213
DB_HOST=localhost  # or 'postgres' if using docker-compose network
```

### Service can't connect to other services
```bash
# Check if target service is running
curl http://localhost:9001/health  # User Service

# Check service addresses in .env
USER_SERVICE_GRPC=localhost:9001   # For local development
# or
USER_SERVICE_GRPC=user-service:9001  # For docker-compose
```

## Migration Guide

If you have old `.env` file with shared variables:

```bash
# Old format (shared)
DB_HOST=localhost
USER_SERVICE_GRPC=localhost:9090

# New format (per-service)
# In services/user-service/.env
DB_HOST=localhost

# In services/order-service/.env
USER_SERVICE_GRPC=localhost:9001  # Note: correct port
```

Use the script to migrate:
```bash
# TODO: Create migration script
./scripts/migrate-env.sh
```

## Additional Resources

- [Configuration Template Documentation](../CONFIG_TEMPLATE.md)
- [Docker Compose Setup](../docker-compose.yaml)
- [Deployment Guide](../docs/deployment/deployment_guide.md)
