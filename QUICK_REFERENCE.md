# 🚀 Quick Reference - Docker & Makefile Commands

## 📋 Most Common Commands

### Start Everything (First Time)
```bash
make setup          # Install dependencies
make dev            # Start all services with Docker
make dev-logs       # Watch logs
```

### Daily Development
```bash
make dev            # Start all services
make restart        # Restart everything
make stop           # Stop containers
make docker-logs    # View logs
```

### Build & Test
```bash
make build-all      # Build all .exe files
make test           # Run tests
make check          # Format + Lint + Test
```

### Database Operations
```bash
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make db-reset       # Reset all databases (⚠️ DELETES DATA!)
```

### Docker Operations
```bash
make docker-build              # Build all images
make docker-ps                 # Show containers
docker-compose up -d postgres  # Start only PostgreSQL
docker-compose logs -f <service>  # View service logs
```

---

## 🌐 Service URLs

### Main Gateway
- **API Gateway**: http://localhost:8000
- **Health**: http://localhost:8000/health

### Microservices (HTTP)
- **User**: http://localhost:8001
- **Product**: http://localhost:8002
- **Order**: http://localhost:8003
- **Notification**: http://localhost:8004
- **Inventory**: http://localhost:8005
- **Payment**: http://localhost:8006

### Microservices (gRPC)
- **User**: localhost:9001
- **Product**: localhost:9002
- **Order**: localhost:9003
- **Notification**: localhost:9004
- **Inventory**: localhost:9005
- **Payment**: localhost:9006

### Infrastructure
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **RabbitMQ Management**: http://localhost:15672 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin123)
- **Jaeger UI**: http://localhost:16686

---

## 🧪 Quick Tests

### Health Checks
```bash
# All services
curl http://localhost:8000/health  # API Gateway
curl http://localhost:8001/health  # User Service
curl http://localhost:8002/health  # Product Service
curl http://localhost:8003/health  # Order Service
curl http://localhost:8004/health  # Notification Service
curl http://localhost:8005/health  # Inventory Service
curl http://localhost:8006/health  # Payment Service
```

### Test API (via Gateway)
```bash
# Register User
curl -X POST http://localhost:8000/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test123!"
  }'

# Login
curl -X POST http://localhost:8000/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!"
  }'

# Get Products (with JWT token from login)
curl -X GET http://localhost:8000/api/v1/products \
  -H "Authorization: Bearer <your-jwt-token>"
```

---

## 🗄️ Database Access

### PostgreSQL via CLI
```bash
# Connect to PostgreSQL container
docker exec -it ecommerce-postgres psql -U postgres

# List databases
\l

# Connect to specific database
\c users_db

# List tables
\dt

# Query users
SELECT * FROM users;

# Exit
\q
```

### Redis via CLI
```bash
# Connect to Redis container
docker exec -it ecommerce-redis redis-cli

# Check keys
KEYS *

# Get value
GET user:123

# List all databases
INFO keyspace

# Select database
SELECT 0  # User service uses DB 0

# Exit
exit
```

---

## 🔧 Troubleshooting

### Service Won't Start
```bash
# View logs
docker-compose logs <service-name>

# Example
docker-compose logs user-service

# Restart specific service
docker-compose restart user-service
```

### Port Already in Use
```bash
# Windows: Find process using port
netstat -ano | findstr "8000"

# Kill process
taskkill /PID <PID> /F

# Or stop all Docker containers
make stop-all
```

### Database Connection Failed
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check databases exist
docker exec -it ecommerce-postgres psql -U postgres -c "\l"

# Recreate databases
make db-reset
make migrate-up
```

### Reset Everything
```bash
# Nuclear option: Reset everything
make stop-all
make clean
make dev
```

---

## 📊 Monitoring

### View Metrics
```bash
# Prometheus
open http://localhost:9090

# Grafana
open http://localhost:3000
# Login: admin/admin123

# Jaeger (Distributed Tracing)
open http://localhost:16686
```

### Container Stats
```bash
# Real-time stats
docker stats

# Specific service
docker stats ecommerce-user-service
```

---

## 🐛 Debug Mode

### Run Service Locally (Outside Docker)
```bash
# Set environment variables
export HTTP_PORT=8001
export GRPC_PORT=9001
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres123
export DB_NAME=users_db

# Run service
cd services/user-service
go run ./cmd/main.go
```

### Run with Debugger (VS Code)
1. Set breakpoints in code
2. Press F5 (Start Debugging)
3. Select "Go: Launch Package"
4. Service will start with debugger attached

---

## 📦 Production Deployment

### Build & Push Images
```bash
# Build
make docker-build

# Tag for production
docker tag ecommerce-user-service:latest your-registry.com/ecommerce-user-service:v1.0.0

# Push to registry
make docker-push
```

### Deploy to Kubernetes
```bash
# Apply manifests (if you have K8s configs)
kubectl apply -f infrastructure/k8s/

# Check status
kubectl get pods
kubectl get services
```

---

## 📝 Environment Variables

### Change Default Ports (docker-compose.yaml)
```yaml
user-service:
  environment:
    - HTTP_PORT=8001  # Change this
    - GRPC_PORT=9001  # Change this
```

### Change Database Credentials
```yaml
postgres:
  environment:
    - POSTGRES_PASSWORD=your-new-password  # Change this

user-service:
  environment:
    - DB_PASSWORD=your-new-password  # Must match
```

---

## 🎯 Common Workflows

### Add New Service
1. Create service directory: `services/new-service/`
2. Add to Makefile `SERVICES` variable
3. Add to `docker-compose.yaml`
4. Create migrations: `services/new-service/migrations/`
5. Build: `make build-all`
6. Start: `make dev`

### Update Proto Definitions
1. Edit proto files: `proto/service_name/service.proto`
2. Generate code: `make proto-gen`
3. Rebuild services: `make build-all`

### Run Integration Tests
```bash
# Start services
make dev

# Run tests
go test ./tests/integration/...

# Or use Makefile
make test
```

---

## 🔐 Security Notes

⚠️ **Before Production**:
- Change all default passwords
- Update JWT secret (> 32 chars)
- Enable SSL/TLS for databases
- Configure firewall rules
- Use secrets management (Vault, AWS Secrets Manager)
- Enable rate limiting
- Configure CORS properly
- Use API keys for external services

---

## 📚 More Info

- **Full Setup Guide**: `DOCKER_SETUP.md`
- **Changelog**: `CHANGELOG_DOCKER.md`
- **Architecture**: `docs/architecture/system_design.md`
- **API Docs**: `docs/api/swagger.yaml`

---

**Last Updated**: 2025-10-15
