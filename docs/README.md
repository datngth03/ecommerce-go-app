# Documentation & Testing - Quick Reference

## 📋 Overview

This directory contains all deliverables for **Phase 3: Documentation & Testing** of the E-Commerce microservices platform.

---

## 📚 Documentation Index

### API Documentation
- **[API_REFERENCE.md](./api/API_REFERENCE.md)** - Complete REST API documentation
  - 30+ endpoints across 6 services
  - Request/response examples
  - Authentication guide
  - Error handling reference

- **[api/swagger.yaml](./api/swagger.yaml)** - OpenAPI 3.0 specification
  - Machine-readable API spec for client SDK generation
  - Validate: `npx @stoplight/spectral-cli lint docs/api/swagger.yaml`
  - Import into Swagger Editor or Postman

- **[api/postman/](./api/postman/)** - Postman collection & environment
  - `ecommerce.postman_collection.json` - Complete API test collection
  - `ecommerce-local.postman_environment.json` - Local environment variables
  - See [POSTMAN_GUIDE.md](./api/postman/POSTMAN_GUIDE.md) for usage

### Deployment
- **[deployment_guide.md](./deployment/deployment_guide.md)** - Comprehensive deployment documentation
  - Development setup (Docker Compose)
  - Production deployment (Kubernetes)
  - Database management
  - Monitoring & observability
  - Backup & recovery
  - Troubleshooting guide
  - Rollback procedures

### Testing
- **[api/postman/POSTMAN_GUIDE.md](./api/postman/POSTMAN_GUIDE.md)** - API testing with Postman
  - Collection: `ecommerce.postman_collection.json`
  - Environment setup & test flows
  - Automated tests & troubleshooting

- **[Integration Tests](../tests/integration/)** - Go integration test suite
  - Integration service testing
  - Run: `cd tests/integration && go test -v`

- **[End-to-End Tests](../tests/e2e/)** - Go end-to-end test suite
  - End-to-end service testing
  - Script: `.\tests\e2e\test-simple.ps1`

---

## 🚀 Quick Start

### For Developers

```bash
# 1. Set up development environment
cd ecommerce-go-app
docker-compose up -d

# 2. Run integration tests
cd tests/integration
go test -v

# 3. Test API with Postman
# Import: docs/api/postman/ecommerce.postman_collection.json
# Import: docs/api/postman/ecommerce-local.postman_environment.json
```

### For DevOps

```bash
# 1. Review deployment guide
cat docs/deployment/deployment_guide.md

# 2. Set up production environment
# Follow Kubernetes deployment section

# 3. Configure monitoring
# Follow Monitoring & Health Checks section
```

---

## 🔗 Related Documentation

### Architecture
- [System Design](./architecture/system_design.md)
- [Database Schema](./architecture/database_schema.md)

### Getting Started
- [Quick Start Guide](../QUICK_START.md)

---

## 💡 Key Features Documented

### Microservices Architecture
- 6 independent services
- gRPC inter-service communication
- API Gateway pattern
- Event-driven design with RabbitMQ

### Infrastructure
- PostgreSQL 15 (separate DB per service)
- Redis 7 (caching & sessions)
- RabbitMQ (async messaging)
- Prometheus + Grafana (monitoring)

### Security
- JWT authentication
- Password hashing (bcrypt)
- TLS encryption
- Rate limiting
- Secrets management

### Deployment
- Docker Compose (development)
- Kubernetes (production)
- CI/CD ready
- Zero-downtime deployments
- Automated backups

### Testing
- Unit tests
- Integration tests
- E2E tests
- Load tests (future)
- Contract tests (future)

---

## 🎯 Usage Examples

### Run Complete Test Suite

```bash
# E2E tests
cd tests/e2e
./test-simple.ps1

# Integration tests
cd tests/integration
go test -v ./...

```

### Deploy to Production

```bash
# Kubernetes
kubectl create namespace ecommerce-prod
kubectl apply -f infrastructure/k8s/

# Docker Compose
docker-compose -f docker-compose.prod.yaml up -d
```

---



**Last Updated**: October 21, 2025  
**Version**: 1.0.0  
**Maintained by**: DatGuru
