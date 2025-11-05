# 🛒 E-Commerce Microservices Platform# 🛒 E-commerce Microservices Platform



A production-ready, scalable e-commerce platform built with **Go microservices architecture**, featuring clean code, comprehensive testing, and complete documentation.A scalable e-commerce platform built with **Go microservices architecture**, designed for high performance and maintainability.



[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)## 🚀 Features

[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)- **Microservices Architecture**: 6 independent services with clear separation of concerns

- **API Gateway**: Centralized routing, authentication, and rate limiting

---- **Event-Driven**: Asynchronous communication using message queues

- **Database per Service**: Each service has its own database for data isolation

## 🚀 Features- **Containerized**: Docker-ready with docker-compose for easy deployment

- **gRPC Communication**: High-performance inter-service communication

- **Microservices Architecture** - 6 independent services with clear separation of concerns- **Clean Architecture**: Following Domain-Driven Design principles

- **API Gateway** - Centralized routing, authentication, and rate limiting

- **Event-Driven** - Asynchronous communication using RabbitMQ## 🏗️ Architecture

- **Database per Service** - PostgreSQL with isolated databases for each service

- **gRPC Communication** - High-performance inter-service communication```

- **RESTful APIs** - Clean REST API design with OpenAPI documentation┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐

- **Docker & Kubernetes** - Containerized deployment with orchestration support│   Web Client    │    │   Mobile App    │    │  Admin Portal   │

- **Monitoring & Tracing** - Prometheus, Grafana, and Jaeger integration└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘

- **Comprehensive Testing** - Unit, integration, and E2E tests          │                      │                      │

- **Clean Architecture** - Following DDD and SOLID principles          └──────────────────────┼──────────────────────┘

                                 │

---                    ┌────────────▼───────────────┐

                    │   API Gateway :8000        │

## 🏗️ System Architecture                    │   (Authentication,         │

                    │    Rate Limiting)          │

```                    └────────────┬───────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                                 │

│   Web Client    │    │   Mobile App    │    │  Admin Portal   │        ┌────────────────────────┼──────────────────────────┬────────────┐

└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘        │                        │                          │            │

          │                      │                      │   ┌────▼────┐  ┌────▼────┐  ┌──▼─────┐  ┌────▼─────┐  ┌──▼──────┐  ┌─▼────────┐

          └──────────────────────┼──────────────────────┘   │  User   │  │Product  │  │ Order  │  │Inventory │  │Payment  │  │Notification│

                                 │   │ :8001   │  │ :8002   │  │ :8003  │  │  :8005   │  │ :8006   │  │  :8004    │

                    ┌────────────▼───────────────┐   │ :9001   │  │ :9002   │  │ :9003  │  │  :9005   │  │ :9006   │  │  :9004    │

                    │   API Gateway :8000        │   └────┬────┘  └────┬────┘  └───┬────┘  └────┬─────┘  └────┬────┘  └────┬──────┘

                    │   (Auth, Rate Limiting)    │        │            │           │            │            │            │

                    └────────────┬───────────────┘   ┌────▼────┐  ┌───▼────┐  ┌───▼────┐  ┌───▼─────┐  ┌──▼─────┐  ┌───▼──────┐

                                 │   │users_db │  │products│  │orders  │  │inventory│  │payments│  │notifications│

        ┌────────────────────────┼──────────────────────────┬────────────┐   │         │  │_db     │  │_db     │  │_db      │  │_db     │  │_db         │

        │                        │                          │            │   └─────────┘  └────────┘  └────────┘  └─────────┘  └────────┘  └────────────┘

   ┌────▼────┐  ┌────▼────┐  ┌──▼─────┐  ┌────▼─────┐  ┌──▼──────┐  ┌─▼────────┐                                 │

   │  User   │  │Product  │  │ Order  │  │Inventory │  │Payment  │  │Notification│                          ┌──────▼──────┐

   │ :8001   │  │ :8002   │  │ :8003  │  │  :8005   │  │ :8006   │  │  :8004    │                          │  RabbitMQ   │

   │ :9001   │  │ :9002   │  │ :9003  │  │  :9005   │  │ :9006   │  │  :9004    │                          │   :5672     │

   └────┬────┘  └────┬────┘  └───┬────┘  └────┬─────┘  └────┬────┘  └────┬──────┘                          └─────────────┘

        │            │           │            │            │            │```

   ┌────▼────┐  ┌───▼────┐  ┌───▼────┐  ┌───▼─────┐  ┌──▼─────┐  ┌───▼──────┐

   │users_db │  │products│  │orders  │  │inventory│  │payments│  │notifications│**Note:** Each service has HTTP (80xx) and gRPC (90xx) ports for inter-service communication.

   │         │  │_db     │  │_db     │  │_db      │  │_db     │  │_db         │

   └─────────┘  └────────┘  └────────┘  └─────────┘  └────────┘  └────────────┘

                                 │## 🛠️ Tech Stack

                          ┌──────▼──────┐

                          │  RabbitMQ   │### Backend

                          │   :5672     │- **Language**: Go 1.24+

                          └─────────────┘- **Framework**: Gin/Echo

```- **Database**: PostgreSQL

- **Cache**: Redis

**Note:** Each service exposes HTTP (80xx) and gRPC (90xx) ports.- **Message Queue**: RabbitMQ

- **Communication**: gRPC + HTTP REST

---- **Container**: Docker + Docker Compose



## 📋 Microservices Overview### DevOps & Tools

- **API Documentation**: Swagger/OpenAPI

| Service | HTTP | gRPC | Responsibility | Database |- **Testing**: Testify, Integration Tests

|---------|------|------|----------------|----------|- **Monitoring**: Prometheus + Grafana

| **API Gateway** | 8000 | - | Request routing, auth, rate limiting | - |- **Logging**: Structured logging with Zap

| **User Service** | 8001 | 9001 | User management, authentication, JWT | `users_db` |- **Migration**: golang-migrate

| **Product Service** | 8002 | 9002 | Product catalog, categories | `products_db` |

| **Order Service** | 8003 | 9003 | Shopping cart, order processing | `orders_db` |## 📋 Services

| **Notification Service** | 8004 | 9004 | Email/SMS notifications | `notifications_db` |

| **Inventory Service** | 8005 | 9005 | Stock management, reservations | `inventory_db` || Service | Port | Description | Database |

| **Payment Service** | 8006 | 9006 | Payment processing (Stripe), refunds | `payments_db` ||---------|------|-------------|----------|

| API Gateway | 8000 | Entry point, routing, auth | - |

---| User Service | 8001 | User management, authentication | users_db |

| Product Service | 8002 | Product catalog, categories | products_db |

## 🛠️ Tech Stack| Order Service | 8003 | Order processing, shopping cart | orders_db |

| Notification Service | 8004 | Email, SMS notifications | notifications_db |

**Backend:**| Inventory Service | 8005 | Stock management | inventory_db |

- [Go 1.24+](https://go.dev/) - Primary language| Payment Service | 8006 | Payment processing, transactions | payments_db |

- [Gin](https://gin-gonic.com/) / [Echo](https://echo.labstack.com/) - HTTP frameworks

- [gRPC](https://grpc.io/) - Inter-service communication## 🚀 Quick Start

- [PostgreSQL 14+](https://www.postgresql.org/) - Primary database

- [Redis 6+](https://redis.io/) - Caching & sessions### Prerequisites

- [RabbitMQ](https://www.rabbitmq.com/) - Message queue- Go 1.24 or higher

- Docker & Docker Compose

**DevOps & Infrastructure:**- PostgreSQL 14+

- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/) - Containerization- Redis 6+

- [Kubernetes](https://kubernetes.io/) - Orchestration (production)

- [Prometheus](https://prometheus.io/) & [Grafana](https://grafana.com/) - Monitoring & metrics### ⚡ Phase 2: One-Command Start (Recommended)

- [Jaeger](https://www.jaegertracing.io/) - Distributed tracing```powershell

- [Nginx](https://nginx.org/) - Reverse proxy & load balancer# Start all services and run tests

.\scripts\quick-start-phase2.ps1 -RunTests

**Development:**

- [OpenAPI 3.0](https://swagger.io/specification/) - API specification# Just start services (no tests)

- [Postman](https://www.postman.com/) - API testing.\scripts\quick-start-phase2.ps1

- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations

- [Testify](https://github.com/stretchr/testify) - Testing framework# Stop all services

.\scripts\quick-start-phase2.ps1 -StopAll

---```



## 🚀 Quick StartThis will:

1. Check Docker is running

### Prerequisites2. Start infrastructure (PostgreSQL, Redis, RabbitMQ)

3. Run database migrations

- [Docker Desktop](https://www.docker.com/products/docker-desktop) installed and running4. Start all 7 microservices

- [Go 1.24+](https://go.dev/dl/) (for local development)5. Verify health of all services

- [Git](https://git-scm.com/) for version control6. Run automated tests (with `-RunTests` flag)



### Option 1: Docker Compose (Recommended)**See:** [Quick Start Guide](QUICK_START.md) for detailed setup instructions



```bash---

# Clone the repository

git clone https://github.com/datngth03/ecommerce-go-app.git### Manual Setup (Alternative)

cd ecommerce-go-app

### 1. Clone the repository

# Start all services```bash

docker-compose up -dgit clone https://github.com/datngth03/ecommerce-go-app.git

cd ecommerce-go-app

# Check status```

docker-compose ps

### 2. Environment Setup

# View logs```bash

docker-compose logs -fcp .env.example .env

# Edit .env with your configurations

# Stop all services```

docker-compose down

```### 3. Start Infrastructure

```bash

**Services will be available at:**# Start databases, message queue, and monitoring

- API Gateway: http://localhost:8000docker-compose up -d postgres redis rabbitmq

- Grafana: http://localhost:3000 (admin/admin123)```

- Prometheus: http://localhost:9090

- Jaeger UI: http://localhost:16686### 4. Database Migration

- RabbitMQ Management: http://localhost:15672 (admin/admin123)```bash

# Run migrations for all services

### Option 2: Quick Start Scriptmake migrate-up

```

```powershell

# Windows PowerShell### 5. Start Services

.\scripts\quick-start-phase2.ps1```bash

# Option 1: Using Docker (Recommended)

# With automated testsdocker-compose up -d

.\scripts\quick-start-phase2.ps1 -RunTests

# Option 2: Local development

# Stop all servicescd services/api-gateway && go run cmd/main.go

.\scripts\quick-start-phase2.ps1 -StopAll```

```

### 6. Verify Installation

### Option 3: Manual Local Development```bash

# Check API Gateway health

See [QUICK_START.md](QUICK_START.md) for detailed manual setup instructions.curl http://localhost:8000/health



---# Check individual services

curl http://localhost:8001/health  # User Service

## 📚 Documentationcurl http://localhost:8002/health  # Product Service

curl http://localhost:8003/health  # Order Service

### Getting Startedcurl http://localhost:8004/health  # Notification Service

- [Quick Start Guide](QUICK_START.md) - Comprehensive setup guide with troubleshootingcurl http://localhost:8005/health  # Inventory Service

- [Architecture Documentation](docs/architecture/system_design.md) - System design & patternscurl http://localhost:8006/health  # Payment Service

- [Database Schema](docs/architecture/database_schema.md) - Complete database documentation```



### API Documentation## 📚 API Documentation

- [API Reference](docs/API_REFERENCE.md) - Complete REST API documentation (40+ endpoints)

- [OpenAPI Specification](docs/api/swagger.yaml) - Machine-readable API spec### 🎯 Phase 2: Complete Testing Suite

- [Postman Collection](docs/api/postman/) - Ready-to-use API test collection

- [Postman Guide](docs/api/postman/POSTMAN_GUIDE.md) - API testing instructions**Quick Start:**

```powershell

### Deployment# Run automated tests

- [Deployment Guide](docs/deployment/deployment_guide.md) - Docker, Kubernetes, production deployment.\tests\e2e\test-api.ps1

- [Environment Configuration](QUICK_START.md#configuration) - Environment variables & secrets

# Or use quick start with tests

### Additional Resources.\scripts\quick-start-phase2.ps1 -RunTests

- [Documentation Index](docs/README.md) - Complete documentation overview```



---**Documentation:**

- **API Reference**: [API_REFERENCE.md](docs/API_REFERENCE.md) - Complete API documentation

## 🧪 Testing- **Postman Guide**: [POSTMAN_GUIDE.md](docs/api/postman/POSTMAN_GUIDE.md) - API testing guide

- **Postman Collection**: [ecommerce.postman_collection.json](docs/api/postman/ecommerce.postman_collection.json)

### Run All Tests

**Test Coverage:**

```bash- User Service (Register, Login, Profile)

# Integration tests- Product Service (CRUD operations)

cd tests/integration- Inventory Service (Stock management)

go test -v- Order Service (Cart, Orders)

- Payment Service (Process, Confirm, Refund)

# E2E tests (PowerShell)- End-to-End E-Commerce Flow

.\tests\e2e\test-simple.ps1

```---



### API Testing with Postman### Swagger UI

- **API Gateway**: http://localhost:8000/swagger/ (Coming soon)

1. Import collection: `docs/api/postman/ecommerce.postman_collection.json`- **Individual Services**: http://localhost:800X/swagger/ (Coming soon)

2. Import environment: `docs/api/postman/ecommerce-local.postman_environment.json`

3. Select "E-Commerce Local Environment"### Postman Collection

4. Run requests sequentially or use Collection RunnerImport `docs/api/postman/ecommerce-phase2.postman_collection.json` for testing all APIs.



### Quick API Test (curl)**How to use:**

1. Import collection into Postman

```bash2. Run "Register User" → Auto-saves user_id

# Health check3. Run "Login" → Auto-saves access_token

curl http://localhost:8000/health4. All subsequent requests use the token automatically

5. Follow the numbered folders (1. User Service → 5. Payment Service)

# Register user

curl -X POST http://localhost:8000/api/v1/auth/register \## 🧪 Testing

  -H "Content-Type: application/json" \

  -d '{### ⚡ Automated Testing (Phase 2)

    "email": "user@example.com",```powershell

    "password": "SecurePass123!",# Run complete automated test suite (20+ tests)

    "username": "testuser",.\tests\e2e\test-api.ps1

    "full_name": "Test User"

  }'# Expected output:

# User Service (3 tests)

# Login (get token)# Product Service (3 tests)

curl -X POST http://localhost:8000/api/v1/auth/login \# Inventory Service (2 tests)

  -H "Content-Type: application/json" \# Order Service (5 tests)

  -d '{# Payment Service (6 tests)

    "email": "user@example.com",# Inventory Verification (1 test)

    "password": "SecurePass123!"# 

  }'# 📊 Test Summary

# Total Tests: 20

# List products# Passed: 20

curl http://localhost:8000/api/v1/products# Failed: 0

```# Pass Rate: 100%

```

---

**See:** [Integration Tests](tests/integration/) for test suite details

## 🔍 Monitoring & Observability

### Unit Tests

### Metrics (Prometheus + Grafana)```bash

# Run tests for all services

- **Prometheus**: http://localhost:9090make test

  - Service health metrics

  - Business metrics (orders, revenue)# Run tests for specific service

  - Infrastructure metrics (CPU, memory)cd services/user-service && go test ./...

```

- **Grafana**: http://localhost:3000 (admin/admin123)

  - Pre-configured dashboards### Integration Tests

  - Real-time service monitoring```bash

  - Custom alerting rules# Start test environment

make test-env-up

### Distributed Tracing (Jaeger)

# Run integration tests

- **Jaeger UI**: http://localhost:16686make test-integration

  - Request tracing across services```

  - Performance bottleneck identification

  - Error analysis### Load Testing

```bash

### Logging# Using K6

k6 run tests/load/k6/load_test.js

- Structured JSON logging with correlation IDs```

- Centralized log aggregation

- Log levels: DEBUG, INFO, WARNING, ERROR## 📊 Monitoring



---### Prometheus Metrics

- **URL**: http://localhost:9090

## 🏢 Project Structure- **Metrics**: Request duration, error rates, database connections



```### Grafana Dashboards

ecommerce-go-app/- **URL**: http://localhost:3000

├── services/                      # Microservices- **Default Login**: admin/admin

│   ├── api-gateway/              # API Gateway service- **Dashboards**: Service metrics, business metrics

│   ├── user-service/             # User management

│   ├── product-service/          # Product catalog### Application Logs

│   ├── order-service/            # Order processing```bash

│   ├── payment-service/          # Payment handling# View logs for all services

│   ├── inventory-service/        # Stock managementdocker-compose logs -f

│   └── notification-service/     # Notifications

│# View logs for specific service

├── proto/                        # gRPC protocol definitionsdocker-compose logs -f user-service

│   ├── user_service/```

│   ├── product_service/

│   └── ...## 🔧 Development

│

├── shared/                       # Shared packages### Project Structure

│   └── pkg/```

│       ├── config/              # Configuration utilsecommerce-microservices/

│       ├── middleware/          # Shared middleware├── services/           # Microservices

│       └── errors/              # Error handling├── shared/            # Shared libraries

│├── infrastructure/    # Docker, K8s configs

├── infrastructure/               # Infrastructure configs├── docs/             # Documentation

│   ├── docker/                  # Dockerfiles & configs└── scripts/          # Build and deployment scripts

│   ├── k8s/                     # Kubernetes manifests```

│   └── monitoring/              # Monitoring configs

│### Adding a New Service

├── docs/                        # Documentation1. Create service directory in `services/`

│   ├── api/                     # API documentation2. Follow the established structure (cmd, internal, pkg)

│   ├── architecture/            # Architecture docs3. Add to docker-compose.yml

│   └── deployment/              # Deployment guides4. Update API Gateway routing

│5. Add monitoring and documentation

├── tests/                       # Test suites

│   ├── integration/             # Integration tests### Code Standards

│   └── e2e/                     # End-to-end tests- Follow Go conventions and best practices

│- Use dependency injection

├── scripts/                     # Automation scripts- Implement proper error handling

├── docker-compose.yaml          # Docker Compose config- Write comprehensive tests

├── Makefile                     # Build automation- Document APIs with Swagger

└── README.md                    # This file

```## 📱 Client Applications



---The backend provides RESTful APIs that can be consumed by:

- **Web Applications** (React, Vue.js, Angular)

## 🔧 Development- **Mobile Apps** (React Native, Flutter)

- **Desktop Applications** (Electron)

### Build All Services- **Third-party Integrations**



```bash### Example API Calls

# Using Make

make build-all**See complete examples in:** [API Reference](docs/API_REFERENCE.md) and [Postman Guide](docs/api/postman/POSTMAN_GUIDE.md)



# Or manually```powershell

cd services/user-service && go build -o user-service.exe cmd/main.go# User Registration

cd services/product-service && go build -o product-service.exe cmd/main.gocurl -X POST http://localhost:8000/api/v1/auth/register `

# ... repeat for other services  -H "Content-Type: application/json" `

```  -d '{"email":"user@example.com","password":"SecurePass123!","username":"user","full_name":"User Name"}'



### Run Individual Service Locally# Login

curl -X POST http://localhost:8000/api/v1/auth/login `

```bash  -H "Content-Type: application/json" `

cd services/user-service  -d '{"email":"user@example.com","password":"SecurePass123!"}'



# Set environment variables# Get Products (Public)

export DB_HOST=localhostcurl http://localhost:8000/api/v1/products?page=1&page_size=10

export DB_PASSWORD=postgres123

export HTTP_PORT=8001# Get Profile (Authenticated)

export GRPC_PORT=9001curl http://localhost:8000/api/v1/users/me `

  -H "Authorization: Bearer YOUR_TOKEN"

# Run service

go run cmd/main.go# Add to Cart

```curl -X POST http://localhost:8000/api/v1/cart `

  -H "Content-Type: application/json" `

### Generate gRPC Code  -H "Authorization: Bearer YOUR_TOKEN" `

  -d '{"product_id":1,"quantity":2,"price":99.99}'

```bash

# Regenerate all proto files# Create Order

./scripts/generate_protos.shcurl -X POST http://localhost:8000/api/v1/orders `

```  -H "Content-Type: application/json" `

  -H "Authorization: Bearer YOUR_TOKEN" `

### Database Migrations  -d '{"shipping_address":"123 Main St","payment_method":"stripe"}'



```bash# Process Payment

# Run migrationscurl -X POST http://localhost:8000/api/v1/payments `

make migrate-up  -H "Content-Type: application/json" `

  -H "Authorization: Bearer YOUR_TOKEN" `

# Rollback migrations  -d '{"order_id":"1","amount":199.98,"method":"stripe","currency":"USD"}'

make migrate-down```



# Reset databases (⚠️ deletes all data)**For complete API reference with 40+ endpoints, see:**

make db-reset- [API Reference](docs/API_REFERENCE.md) - Complete REST API documentation

```- [Postman Collection](docs/api/postman/ecommerce.postman_collection.json) - Ready-to-use API tests

- [Postman Guide](docs/api/postman/POSTMAN_GUIDE.md) - Testing instructions

---

## Quick API Examples

## 🚀 Deployment```bash

# Get Products

### Docker Productioncurl http://localhost:8000/api/v1/products



```bash# Create Order  

# Build production imagescurl -X POST http://localhost:8000/api/v1/orders \

docker-compose -f docker-compose.prod.yml build  -H "Authorization: Bearer YOUR_JWT_TOKEN" \

  -H "Content-Type: application/json" \

# Deploy to production  -d '{"items":[{"product_id":"<uuid>","quantity":2}]}'

docker-compose -f docker-compose.prod.yml up -d```

```

## 🚀 Deployment

### Kubernetes

### Docker Production

```bash```bash

# Apply all Kubernetes manifests# Build all services

kubectl apply -f infrastructure/k8s/make build-docker



# Check deployment status# Deploy to production

kubectl get pods -n ecommercedocker-compose -f docker-compose.prod.yml up -d

```

# View logs

kubectl logs -f <pod-name> -n ecommerce### Kubernetes

``````bash

# Apply Kubernetes manifests

### Environment Variables (Production)kubectl apply -f infrastructure/k8s/

```

Key environment variables for production deployment:

### Environment Variables

```bashKey environment variables for production:

# Database```bash

DB_HOST=your-postgres-host# Database

DB_PASSWORD=your-secure-passwordDB_HOST=your-postgres-host

DB_PASSWORD=your-secure-password

# JWT Authentication

JWT_SECRET=your-jwt-secret-key# JWT

JWT_SECRET=your-jwt-secret

# Payment Gateway

STRIPE_SECRET_KEY=your-stripe-key# Payment

STRIPE_WEBHOOK_SECRET=your-webhook-secretSTRIPE_SECRET_KEY=your-stripe-key



# Notifications# Notification

SMTP_HOST=smtp.example.comSMTP_PASSWORD=your-smtp-password

SMTP_PASSWORD=your-smtp-password```



# Monitoring## 🤝 Contributing

PROMETHEUS_ENABLED=true

JAEGER_ENABLED=true1. Fork the repository

```2. Create a feature branch (`git checkout -b feature/amazing-feature`)

3. Commit your changes (`git commit -m 'Add amazing feature'`)

---4. Push to the branch (`git push origin feature/amazing-feature`)

5. Open a Pull Request

## 🤝 Contributing

### Development Workflow

Contributions are welcome! Please follow these steps:- Follow Git Flow branching model

- Write tests for new features

1. Fork the repository- Update documentation

2. Create a feature branch (`git checkout -b feature/amazing-feature`)- Ensure all CI checks pass

3. Commit your changes (`git commit -m 'Add amazing feature'`)

4. Push to the branch (`git push origin feature/amazing-feature`)## 📜 License

5. Open a Pull Request

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Development Guidelines

## 👨‍💻 Author

- Follow [Go best practices](https://go.dev/doc/effective_go)

- Write unit tests for new features**Your Name**

- Update documentation for API changes- GitHub: [@your-username](https://github.com/datngth03)

- Run `go fmt` and `go vet` before committing- LinkedIn: [Your LinkedIn](https://linkedin.com/in/datngth9903)

- Ensure all tests pass before submitting PR- Email: datnt9903@gmail.com



---## 🙏 Acknowledgments



## 📝 License- Go community for excellent libraries and tools

- Microservices patterns from industry best practices

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.- Clean Architecture principles by Robert C. Martin



------



## 👨‍💻 Author⭐ **Star this repository if you find it helpful!**



**Dat Nguyen**## 📈 Project Status

- GitHub: [@datngth03](https://github.com/datngth03)

- LinkedIn: [datngth9903](https://linkedin.com/in/datngth9903)- **MVP**: Core e-commerce functionality

- Email: datnt9903@gmail.com- 🚧 **In Progress**: Advanced analytics, recommendation engine

- 📋 **Planned**: Multi-tenant support, advanced search

---

**Last Updated**: September 2025
## 🙏 Acknowledgments

- Go community for excellent libraries and tools
- Microservices patterns from industry best practices
- Clean Architecture principles by Robert C. Martin
- The amazing open-source community

---

## ⭐ Star This Repository

If you find this project helpful, please consider giving it a star! It helps others discover this project.

---

## 📈 Project Status

- **Core Features**: Complete and production-ready
- **Documentation**: Comprehensive API and deployment docs
- **Testing**: Unit, integration, and E2E tests implemented
- 🚧 **In Progress**: Advanced analytics, recommendation engine
- 📋 **Planned**: Multi-tenant support, advanced search (Elasticsearch)

---

**Last Updated**: October 2025  
**Version**: 2.0.0
