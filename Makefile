.PHONY: help build test run clean docker-build docker-push fmt lint check setup dev migrate-up migrate-down proto-gen proto-clean

# ====================================================================================
# VARIABLES
# ====================================================================================
# List of all microservices and the API Gateway
SERVICES = user-service product-service order-service payment-service inventory-service notification-service api-gateway

# Database configuration
DB_USER = postgres
DB_PASSWORD = postgres123
DB_HOST = localhost
DB_PORT = 5432

# Protobuf tools
PROTOC_GEN_GO_PATH = $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC_PATH = $(shell go env GOPATH)/bin/protoc-gen-go-grpc

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
	@docker-compose up -d
	@sleep 5
	@make migrate-up
	@echo "Running all services..."
	@for service in $(SERVICES); do \
		echo "Starting $$service..."; \
		go run ./services/$$service/cmd/main.go & \
	done
	@wait

run: ## Start all services (for manual testing after a build)
	@for service in $(SERVICES); do \
		echo "Starting $$service..."; \
		go run ./services/$$service/cmd/main.go & \
	done
	@wait

stop: ## Stop all services
	@docker-compose down

clean: ## Clean everything (containers, volumes, local binaries, and generated files)
	@docker-compose down -v --remove-orphans
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@make proto-clean

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
	@for dir in proto/*; do \
		if [ -d "$$dir" ]; then \
			service_name=$$(basename "$$dir"); \
			echo "-> Generating for $$service_name..."; \
			protoc --proto_path=proto \
				--go_out=./services/$$service_name-service/internal/rpc \
				--go_opt=paths=source_relative \
				--go-grpc_out=./services/$$service_name-service/internal/rpc \
				--go-grpc_opt=paths=source_relative \
				proto/$$service_name/*.proto; \
		fi \
	done
	@echo "Protobuf generation complete!"

proto-clean: ## Clean generated Protobuf files
	@echo "Cleaning up generated Protobuf files..."
	@find services -name "*.pb.go" -delete
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
			echo "-> Migrating $$service-service..."; \
			migrate -path services/$$service-service/migrations -database "$$db_url" up || true; \
		else \
			echo "-> No migrations for $$service-service. Skipping."; \
		fi \
	done

migrate-down: ## Rollback the last database migration
	@echo "Rolling back database migrations..."
	@for service in user product order payment inventory notification; do \
		db_name="$${service}s_db"; \
		db_url="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$${db_name}?sslmode=disable"; \
		if [ -d services/$$service-service/migrations ]; then \
			echo "-> Rolling back $$service-service..."; \
			migrate -path services/$$service-service/migrations -database "$$db_url" down 1 || true; \
		else \
			echo "-> No migrations for $$service-service. Skipping."; \
		fi \
	done

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
docker-build: ## Build Docker images
	@docker-compose build

docker-push: ## Push to registry
	@for service in $(SERVICES); do \
		docker tag ecommerce-$$service:latest your-registry.com/$$service:latest; \
		docker push your-registry.com/$$service:latest; \
	done