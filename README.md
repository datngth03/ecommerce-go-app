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



- Dưới đây là sơ đồ kiến trúc cho product-sevice:

```mermaid
graph TB
    %% External Layer
    Client[📱 Client App<br/>Web/Mobile]
    Gateway[🌐 API Gateway<br/>:8080]
    
    %% Docker Network
    subgraph DockerNetwork["🐳 Docker Network: ecommerce-network"]
        %% Product Service Container
        subgraph ProductContainer["📦 Product Service Container :8082"]
            PH[🔄 Product Handler<br/>HTTP Endpoints]
            PS[⚙️ Product Service<br/>Business Logic]
            PR[💾 Product Repository<br/>Database Access]
            PDB[(🗄️ Products DB<br/>PostgreSQL)]
            
            %% gRPC Clients in Product Service
            subgraph PClients["📡 gRPC Clients"]
                UC[👤 User Client<br/>→ user-service:9091]
                IC[📦 Inventory Client<br/>→ inventory-service:9092]
                RC[⭐ Review Client<br/>→ review-service:9093]
            end
        end
        
        %% User Service Container
        subgraph UserContainer["📦 User Service Container :8081"]
            UH[🔄 User Handler<br/>HTTP Endpoints]
            US[⚙️ User Service<br/>Business Logic]
            UR[💾 User Repository<br/>Database Access]
            UDB[(🗄️ Users DB<br/>PostgreSQL)]
            
            %% gRPC Server in User Service
            UGS[🌐 User gRPC Server<br/>:9091]
        end
        
        %% Inventory Service Container
        subgraph InventoryContainer["📦 Inventory Service Container :8083"]
            IH[🔄 Inventory Handler]
            IS[⚙️ Inventory Service]
            IR[💾 Inventory Repository]
            IDB[(🗄️ Inventory DB)]
            IGS[🌐 Inventory gRPC Server<br/>:9092]
        end
        
        %% Review Service Container
        subgraph ReviewContainer["📦 Review Service Container :8084"]
            RH[🔄 Review Handler]
            RS[⚙️ Review Service]
            RR[💾 Review Repository]
            RDB[(🗄️ Reviews DB)]
            RGS[🌐 Review gRPC Server<br/>:9093]
        end
    end
    
    %% HTTP Flow (External)
    Client -->|"1️⃣ HTTP GET<br/>/api/products/123"| Gateway
    Gateway -->|"2️⃣ HTTP Forward<br/>Route to Product Service"| PH
    
    %% Within Product Service
    PH -->|"3️⃣ Parse Request<br/>Extract product ID"| PS
    PS -->|"4️⃣ Get Product<br/>SELECT * FROM products WHERE id=123"| PR
    PR -->|"5️⃣ SQL Query"| PDB
    PDB -->|"6️⃣ Product Data"| PR
    PR -->|"7️⃣ Product Model"| PS
    
    %% gRPC Inter-Service Communication
    PS -->|"8️⃣ gRPC Call<br/>GetUser(seller_id)"| UC
    UC -.->|"9️⃣ gRPC Request<br/>user-service:9091"| UGS
    UGS -->|"🔟 Process Request"| US
    US -->|"1️⃣1️⃣ Query User Data"| UR
    UR -->|"1️⃣2️⃣ SQL Query"| UDB
    UDB -->|"1️⃣3️⃣ User Data"| UR
    UR -->|"1️⃣4️⃣ User Model"| US
    US -->|"1️⃣5️⃣ gRPC Response"| UGS
    UGS -.->|"1️⃣6️⃣ User Info"| UC
    UC -->|"1️⃣7️⃣ User Data"| PS
    
    %% More gRPC calls
    PS -->|"1️⃣8️⃣ gRPC Call<br/>GetStock(product_id)"| IC
    IC -.->|"1️⃣9️⃣ gRPC Request"| IGS
    IGS -->|"2️⃣0️⃣ Process"| IS
    IS -->|"2️⃣1️⃣ Query Stock"| IR
    IR -->|"2️⃣2️⃣ SQL Query"| IDB
    IDB -->|"2️⃣3️⃣ Stock Data"| IR
    IR -->|"2️⃣4️⃣ Stock Info"| IS
    IS -->|"2️⃣5️⃣ gRPC Response"| IGS
    IGS -.->|"2️⃣6️⃣ Stock Info"| IC
    IC -->|"2️⃣7️⃣ Stock Data"| PS
    
    PS -->|"2️⃣8️⃣ gRPC Call<br/>GetAvgRating(product_id)"| RC
    RC -.->|"2️⃣9️⃣ gRPC Request"| RGS
    RGS -->|"3️⃣0️⃣ Process"| RS
    RS -->|"3️⃣1️⃣ Query Reviews"| RR
    RR -->|"3️⃣2️⃣ SQL Query"| RDB
    RDB -->|"3️⃣3️⃣ Review Data"| RR
    RR -->|"3️⃣4️⃣ Rating Info"| RS
    RS -->|"3️⃣5️⃣ gRPC Response"| RGS
    RGS -.->|"3️⃣6️⃣ Rating Info"| RC
    RC -->|"3️⃣7️⃣ Rating Data"| PS
    
    %% Response Flow
    PS -->|"3️⃣8️⃣ Combine Data<br/>ProductDetails{Product, Seller, Stock, Rating}"| PH
    PH -->|"3️⃣9️⃣ HTTP JSON Response<br/>Status 200"| Gateway
    Gateway -->|"4️⃣0️⃣ HTTP Response<br/>Forward to Client"| Client
    
    %% Styling
    classDef client fill:#e1f5fe
    classDef gateway fill:#f3e5f5
    classDef handler fill:#e8f5e8
    classDef service fill:#fff3e0
    classDef repository fill:#fce4ec
    classDef database fill:#f1f8e9
    classDef grpcClient fill:#e3f2fd
    classDef grpcServer fill:#f9fbe7
    classDef container fill:#f5f5f5,stroke:#333,stroke-width:2px
    
    class Client client
    class Gateway gateway
    class PH,UH,IH,RH handler
    class PS,US,IS,RS service
    class PR,UR,IR,RR repository
    class PDB,UDB,IDB,RDB database
    class UC,IC,RC grpcClient
    class UGS,IGS,RGS grpcServer
```


## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.24+
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

### ⚡ Phase 2: One-Command Start (Recommended)
```powershell
# Start all services and run tests
.\scripts\quick-start-phase2.ps1 -RunTests

# Just start services (no tests)
.\scripts\quick-start-phase2.ps1

# Stop all services
.\scripts\quick-start-phase2.ps1 -StopAll
```

This will:
1. ✅ Check Docker is running
2. ✅ Start infrastructure (PostgreSQL, Redis, RabbitMQ)
3. ✅ Run database migrations
4. ✅ Start all 7 microservices
5. ✅ Verify health of all services
6. ✅ Run automated tests (with `-RunTests` flag)

**See:** `docs/PHASE2_SUMMARY.md` for complete Phase 2 documentation

---

### Manual Setup (Alternative)

### 1. Clone the repository
```bash
git clone https://github.com/datngth03/ecommerce-go-app.git
cd ecommerce-go-app
```

### 2. Environment Setup
```bash
cp .env.example .env
# Edit .env with your configurations
```

### 3. Start Infrastructure
```bash
# Start databases, message queue, and monitoring
docker-compose up -d postgres redis rabbitmq
```

### 4. Database Migration
```bash
# Run migrations for all services
make migrate-up
```

### 5. Start Services
```bash
# Option 1: Using Docker (Recommended)
docker-compose up -d

# Option 2: Local development
cd services/api-gateway && go run cmd/main.go
```

### 6. Verify Installation
```bash
# Check API Gateway health
curl http://localhost:8000/health

# Check individual services
curl http://localhost:8001/health  # User Service
curl http://localhost:8002/health  # Product Service
curl http://localhost:8003/health  # Order Service
curl http://localhost:8004/health  # Notification Service
curl http://localhost:8005/health  # Inventory Service
curl http://localhost:8006/health  # Payment Service
```

## 📚 API Documentation

### 🎯 Phase 2: Complete Testing Suite

**Quick Start:**
```powershell
# Run automated tests
.\tests\e2e\test-api.ps1

# Or use quick start with tests
.\scripts\quick-start-phase2.ps1 -RunTests
```

**Documentation:**
- **Complete Test Guide**: `docs/PHASE2_TESTING.md` - All test scenarios with curl examples
- **Phase 2 Summary**: `docs/PHASE2_SUMMARY.md` - Overview of testing infrastructure
- **Postman Collection**: `docs/api/postman/ecommerce-phase2.postman_collection.json` - 40+ endpoints

**Test Coverage:**
- ✅ User Service (Register, Login, Profile)
- ✅ Product Service (CRUD operations)
- ✅ Inventory Service (Stock management)
- ✅ Order Service (Cart, Orders)
- ✅ Payment Service (Process, Confirm, Refund)
- ✅ End-to-End E-Commerce Flow

---

### Swagger UI
- **API Gateway**: http://localhost:8000/swagger/ (Coming soon)
- **Individual Services**: http://localhost:800X/swagger/ (Coming soon)

### Postman Collection
Import `docs/api/postman/ecommerce-phase2.postman_collection.json` for testing all APIs.

**How to use:**
1. Import collection into Postman
2. Run "Register User" → Auto-saves user_id
3. Run "Login" → Auto-saves access_token
4. All subsequent requests use the token automatically
5. Follow the numbered folders (1. User Service → 5. Payment Service)

## 🧪 Testing

### ⚡ Automated Testing (Phase 2)
```powershell
# Run complete automated test suite (20+ tests)
.\tests\e2e\test-api.ps1

# Expected output:
# ✅ User Service (3 tests)
# ✅ Product Service (3 tests)
# ✅ Inventory Service (2 tests)
# ✅ Order Service (5 tests)
# ✅ Payment Service (6 tests)
# ✅ Inventory Verification (1 test)
# 
# 📊 Test Summary
# Total Tests: 20
# Passed: 20
# Failed: 0
# Pass Rate: 100%
```

**See:** `docs/PHASE2_TESTING.md` for detailed test scenarios

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

**See complete examples in:** `docs/PHASE2_TESTING.md`

```powershell
# User Registration
curl -X POST http://localhost:8000/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"user@example.com","password":"SecurePass123!","username":"user","full_name":"User Name"}'

# Login
curl -X POST http://localhost:8000/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"user@example.com","password":"SecurePass123!"}'

# Get Products (Public)
curl http://localhost:8000/api/v1/products?page=1&page_size=10

# Get Profile (Authenticated)
curl http://localhost:8000/api/v1/users/me `
  -H "Authorization: Bearer YOUR_TOKEN"

# Add to Cart
curl -X POST http://localhost:8000/api/v1/cart `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"product_id":1,"quantity":2,"price":99.99}'

# Create Order
curl -X POST http://localhost:8000/api/v1/orders `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"shipping_address":"123 Main St","payment_method":"stripe"}'

# Process Payment
curl -X POST http://localhost:8000/api/v1/payments `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"order_id":"1","amount":199.98,"method":"stripe","currency":"USD"}'
```

**For complete API reference with 40+ endpoints, see:**
- `docs/PHASE2_TESTING.md` - Complete curl examples
- `docs/api/postman/ecommerce-phase2.postman_collection.json` - Postman collection
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