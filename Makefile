.PHONY: help build test run clean docker-build docker-push fmt lint check setup dev migrate-up migrate-down proto-gen proto-clean

# ====================================================================================
# VARIABLES
# ====================================================================================
# List of all microservices
SERVICES = user-service product-service order-service notification-service inventory-service payment-service

# API Gateway
GATEWAY = api-gateway

# Database configuration
DB_USER = postgres
DB_PASSWORD = postgres123
DB_HOST = localhost
DB_PORT = 5432

# Protobuf tools
PROTOC_GEN_GO_PATH = $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC_PATH = $(shell go env GOPATH)/bin/protoc-gen-go-grpc

# Docker registry (change to your registry)
DOCKER_REGISTRY = your-registry.com
DOCKER_TAG = latest

# ====================================================================================
# HELP
# ====================================================================================
help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ====================================================================================
# SETUP
# ====================================================================================
setup: ## Setup development environment and install Go tools
	@echo "Installing Go dependencies..."
	@go mod download
	@echo "Installing Protobuf plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Copying .env file..."
	@cp .env.example .env || true
	@echo "Setup complete. You can now run 'make dev' to start."

# ====================================================================================
# DEVELOPMENT
# ====================================================================================
dev: ## Start infrastructure (DB, Redis, RabbitMQ) and all services
	@echo "Starting infrastructure containers..."
	@docker-compose up -d postgres redis rabbitmq
	@echo "Waiting for infrastructure to be ready..."
	@sleep 10
	@make migrate-up
	@echo "Starting all services..."
	@docker-compose up -d

dev-logs: ## Show logs from all services
	@docker-compose logs -f

dev-services: ## Start only microservices (assumes infrastructure is running)
	@docker-compose up -d user-service product-service order-service notification-service inventory-service payment-service api-gateway

run-local: ## Run all services locally (without Docker)
	@echo "Starting services locally..."
	@for service in $(SERVICES); do \
		echo "Starting $$service..."; \
		(cd services/$$service && go run ./cmd/main.go) & \
	done
	@echo "Starting api-gateway..."
	@(cd services/api-gateway && go run ./cmd/main.go) &
	@echo "All services started. Press Ctrl+C to stop."
	@wait

build-all: ## Build all service binaries
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "-> Building $$service..."; \
		cd services/$$service && go build -o $$service.exe ./cmd; \
		cd ../..; \
	done
	@echo "-> Building api-gateway..."
	@cd services/api-gateway && go build -o api-gateway.exe ./cmd
	@echo "Build complete!"

stop: ## Stop all Docker containers
	@docker-compose down

stop-all: ## Stop and remove all containers, networks, and volumes
	@docker-compose down -v --remove-orphans

restart: stop dev ## Restart all services

clean: ## Clean everything (containers, volumes, binaries, generated files)
	@echo "Cleaning up..."
	@docker-compose down -v --remove-orphans
	@echo "Removing service binaries..."
	@find services -name "*.exe" -delete
	@echo "Removing coverage files..."
	@rm -f coverage.out coverage.html
	@echo "Clean complete."

# ====================================================================================
# PROTOC GENERATION
# ====================================================================================
proto-gen: ## Generate Protobuf Go code for all services
	@echo "Checking for Protobuf tools..."
	@if [ ! -f $(PROTOC_GEN_GO_PATH) ] || [ ! -f $(PROTOC_GEN_GO_GRPC_PATH) ]; then \
		echo "Protobuf plugins not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@echo "Generating Protobuf code..."
	@for proto_dir in proto/*/; do \
		service_name=$$(basename "$$proto_dir"); \
		echo "-> Generating for $$service_name..."; \
		protoc --proto_path=proto \
			--go_out=proto/$$service_name \
			--go_opt=paths=source_relative \
			--go-grpc_out=proto/$$service_name \
			--go-grpc_opt=paths=source_relative \
			proto/$$service_name/*.proto; \
	done
	@echo "Protobuf generation complete!"

proto-clean: ## Clean generated Protobuf files
	@echo "Cleaning up generated Protobuf files..."
	@find proto -name "*.pb.go" -delete
	@find proto -name "*_grpc.pb.go" -delete
	@echo "Clean complete."

# ====================================================================================
# DATABASE
# ====================================================================================
migrate-up: ## Run all pending database migrations
	@echo "Running database migrations..."
	@for service in user product order payment inventory notification; do \
		db_name="$${service}s_db"; \
		db_url="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$${db_name}?sslmode=disable"; \
		if [ -d services/$$service-service/migrations ]; then \
			echo "-> Migrating $$service-service to $${db_name}..."; \
			migrate -path services/$$service-service/migrations -database "$$db_url" up || echo "Migration already applied or error occurred"; \
		else \
			echo "-> No migrations for $$service-service. Skipping."; \
		fi \
	done
	@echo "Migrations complete!"

migrate-down: ## Rollback the last database migration for all services
	@echo "Rolling back database migrations..."
	@for service in user product order payment inventory notification; do \
		db_name="$${service}s_db"; \
		db_url="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$${db_name}?sslmode=disable"; \
		if [ -d services/$$service-service/migrations ]; then \
			echo "-> Rolling back $$service-service from $${db_name}..."; \
			migrate -path services/$$service-service/migrations -database "$$db_url" down 1 || echo "Nothing to rollback"; \
		else \
			echo "-> No migrations for $$service-service. Skipping."; \
		fi \
	done
	@echo "Rollback complete!"

migrate-force: ## Force migration version (use: make migrate-force SERVICE=user VERSION=1)
	@if [ -z "$(SERVICE)" ] || [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-force SERVICE=user VERSION=1"; \
		exit 1; \
	fi
	@db_name="$(SERVICE)s_db"; \
	db_url="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$${db_name}?sslmode=disable"; \
	echo "Forcing migration version $(VERSION) for $(SERVICE)-service..."; \
	migrate -path services/$(SERVICE)-service/migrations -database "$$db_url" force $(VERSION)

db-reset: ## Drop and recreate all databases (WARNING: destroys all data!)
	@echo "WARNING: This will destroy all data! Press Ctrl+C to cancel, or wait 5 seconds..."
	@sleep 5
	@echo "Dropping and recreating databases..."
	@for service in user product order payment inventory notification; do \
		db_name="$${service}s_db"; \
		echo "-> Recreating $${db_name}..."; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $${db_name};"; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $${db_name};"; \
	done
	@echo "Databases recreated. Run 'make migrate-up' to apply migrations."

# ====================================================================================
# CODE QUALITY & TESTING
# ====================================================================================
test: ## Run tests
	@go test -v -race ./...

test-cover: ## Run tests with coverage
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

fmt: ## Format code
	@go fmt ./...

lint: ## Run linter
	@golangci-lint run

check: fmt lint test ## Run all checks

# ====================================================================================
# DOCKER & DEPLOYMENT
# ====================================================================================
docker-build: ## Build Docker images for all services
	@echo "Building Docker images..."
	@docker-compose build

docker-build-service: ## Build Docker image for specific service (use: make docker-build-service SERVICE=user-service)
	@if [ -z "$(SERVICE)" ]; then \
		echo "Usage: make docker-build-service SERVICE=user-service"; \
		exit 1; \
	fi
	@echo "Building Docker image for $(SERVICE)..."
	@docker-compose build $(SERVICE)

docker-push: ## Push Docker images to registry
	@echo "Pushing Docker images to $(DOCKER_REGISTRY)..."
	@for service in $(SERVICES) $(GATEWAY); do \
		echo "-> Tagging and pushing $$service..."; \
		docker tag ecommerce-$$service:latest $(DOCKER_REGISTRY)/ecommerce-$$service:$(DOCKER_TAG); \
		docker push $(DOCKER_REGISTRY)/ecommerce-$$service:$(DOCKER_TAG); \
	done
	@echo "Push complete!"

docker-pull: ## Pull Docker images from registry
	@echo "Pulling Docker images from $(DOCKER_REGISTRY)..."
	@for service in $(SERVICES) $(GATEWAY); do \
		echo "-> Pulling $$service..."; \
		docker pull $(DOCKER_REGISTRY)/ecommerce-$$service:$(DOCKER_TAG); \
	done

docker-logs: ## Show logs from all Docker containers
	@docker-compose logs -f

docker-ps: ## Show running containers
	@docker-compose ps