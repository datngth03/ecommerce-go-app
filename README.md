# 🛒 E-commerce Microservices Platform

A scalable e-commerce platform built with **Go microservices architecture**, designed for high performance and maintainability.

## 🚀 Features

- **Microservices Architecture**: 6 independent services with clear separation of concerns
- **API Gateway**: Centralized routing, authentication, and rate limiting
- **Event-Driven**: Asynchronous communication using message queues
- **Database per Service**: Each service has its own database for data isolation
- **Containerized**: Docker-ready with docker-compose for easy deployment
- **gRPC Communication**: High-performance inter-service communication
- **Clean Architecture**: Following Domain-Driven Design principles

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Mobile App    │    │  Admin Portal   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴──────────────┐
                    │      API Gateway           │
                    │   (Authentication,         │
                    │    Rate Limiting)          │
                    └─────────────┬──────────────┘
                                  │
        ┌─────────────────────────┼─────────────────────────┐
        │                         │                         │
   ┌────▼────┐  ┌────▼────┐  ┌───▼────┐  ┌────▼────┐  ┌───▼────┐
   │  User   │  │Product  │  │ Order  │  │Payment  │  │Inventory│
   │Service  │  │Service  │  │Service │  │Service  │  │Service  │
   └─────────┘  └─────────┘  └────────┘  └─────────┘  └────────┘
        │            │           │           │            │
   ┌────▼────┐  ┌───▼────┐  ┌───▼────┐  ┌───▼────┐   ┌──▼────┐
   │PostgreSQL│  │PostgreSQL│ │PostgreSQL│ │PostgreSQL│ │PostgreSQL│
   └─────────┘  └────────┘  └────────┘  └────────┘   └───────┘
```

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin/Echo
- **Database**: PostgreSQL
- **Cache**: Redis
- **Message Queue**: RabbitMQ
- **Communication**: gRPC + HTTP REST
- **Container**: Docker + Docker Compose

### DevOps & Tools
- **API Documentation**: Swagger/OpenAPI
- **Testing**: Testify, Integration Tests
- **Monitoring**: Prometheus + Grafana
- **Logging**: Structured logging with Zap
- **Migration**: golang-migrate

## 📋 Services

| Service | Port | Description | Database |
|---------|------|-------------|----------|
| API Gateway | 8080 | Entry point, routing, auth | - |
| User Service | 8081 | User management, authentication | users_db |
| Product Service | 8082 | Product catalog, categories | products_db |
| Order Service | 8083 | Order processing, shopping cart | orders_db |
| Payment Service | 8084 | Payment processing, transactions | payments_db |
| Inventory Service | 8085 | Stock management | inventory_db |
| Notification Service | 8086 | Email, SMS notifications | notifications_db |

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 6+

### 1. Clone the repository
```bash
git clone https://github.com/your-username/ecommerce-microservices.git
cd ecommerce-microservices
```

### 2. Environment Setup
```bash
cp .env.example .env
# Edit .env with your configurations
```

### 3. Start Infrastructure
```bash
# Start databases, message queue, and monitoring
docker-compose up -d postgres redis rabbitmq prometheus grafana
```

### 4. Database Migration
```bash
# Run migrations for all services
make migrate-up
```

### 5. Start Services
```bash
# Option 1: Using Docker (Recommended)
docker-compose up

# Option 2: Local development
make run-all
```

### 6. Verify Installation
```bash
# Check API Gateway health
curl http://localhost:8080/health

# Check individual services
curl http://localhost:8081/health  # User Service
curl http://localhost:8082/health  # Product Service
```

## 📚 API Documentation

### Swagger UI
- **API Gateway**: http://localhost:8080/swagger/
- **Individual Services**: http://localhost:808X/swagger/

### Postman Collection
Import `docs/api/postman/ecommerce.postman_collection.json` for testing APIs.

## 🧪 Testing

### Unit Tests
```bash
# Run tests for all services
make test

# Run tests for specific service
cd services/user-service && go test ./...
```

### Integration Tests
```bash
# Start test environment
make test-env-up

# Run integration tests
make test-integration
```

### Load Testing
```bash
# Using K6
k6 run tests/load/k6/load_test.js
```

## 📊 Monitoring

### Prometheus Metrics
- **URL**: http://localhost:9090
- **Metrics**: Request duration, error rates, database connections

### Grafana Dashboards
- **URL**: http://localhost:3000
- **Default Login**: admin/admin
- **Dashboards**: Service metrics, business metrics

### Application Logs
```bash
# View logs for all services
docker-compose logs -f

# View logs for specific service
docker-compose logs -f user-service
```

## 🔧 Development

### Project Structure
```
ecommerce-microservices/
├── services/           # Microservices
├── shared/            # Shared libraries
├── infrastructure/    # Docker, K8s configs
├── docs/             # Documentation
└── scripts/          # Build and deployment scripts
```

### Adding a New Service
1. Create service directory in `services/`
2. Follow the established structure (cmd, internal, pkg)
3. Add to docker-compose.yml
4. Update API Gateway routing
5. Add monitoring and documentation

### Code Standards
- Follow Go conventions and best practices
- Use dependency injection
- Implement proper error handling
- Write comprehensive tests
- Document APIs with Swagger

## 📱 Client Applications

The backend provides RESTful APIs that can be consumed by:
- **Web Applications** (React, Vue.js, Angular)
- **Mobile Apps** (React Native, Flutter)
- **Desktop Applications** (Electron)
- **Third-party Integrations**

### Example API Calls
```bash
# User Registration
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Get Products
curl -X GET http://localhost:8080/api/v1/products?page=1&limit=10

# Create Order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":1,"quantity":2}]}'
```

## 🚀 Deployment

### Docker Production
```bash
# Build all services
make build-docker

# Deploy to production
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes
```bash
# Apply Kubernetes manifests
kubectl apply -f infrastructure/k8s/
```

### Environment Variables
Key environment variables for production:
```bash
# Database
DB_HOST=your-postgres-host
DB_PASSWORD=your-secure-password

# JWT
JWT_SECRET=your-jwt-secret

# Payment
STRIPE_SECRET_KEY=your-stripe-key

# Notification
SMTP_PASSWORD=your-smtp-password
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Workflow
- Follow Git Flow branching model
- Write tests for new features
- Update documentation
- Ensure all CI checks pass

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👨‍💻 Author

**Your Name**
- GitHub: [@your-username](https://github.com/datngth03)
- LinkedIn: [Your LinkedIn](https://linkedin.com/in/datngth9903)
- Email: datnt9903@gmail.com

## 🙏 Acknowledgments

- Go community for excellent libraries and tools
- Microservices patterns from industry best practices
- Clean Architecture principles by Robert C. Martin

---

⭐ **Star this repository if you find it helpful!**

## 📈 Project Status

- ✅ **MVP**: Core e-commerce functionality
- 🚧 **In Progress**: Advanced analytics, recommendation engine
- 📋 **Planned**: Multi-tenant support, advanced search

**Last Updated**: September 2025