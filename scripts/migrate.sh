#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Define a function to display usage instructions.
usage() {
    echo "Usage: $0 {up|down|status} <service_name>"
    echo ""
    echo "Commands:"
    echo "  up       - Apply all pending migrations."
    echo "  down     - Rollback the last applied migration."
    echo "  status   - Show the status of migrations."
    echo ""
    echo "Service Names:"
    echo "  all                 - Apply to all services"
    echo "  user-service"
    echo "  product-service"
    echo "  order-service"
    echo "  payment-service"
    echo "  inventory-service"
    echo "  notification-service"
    exit 1
}

# Check if the correct number of arguments are provided.
if [ "$#" -lt 2 ]; then
    usage
fi

COMMAND=$1
SERVICE_NAME=$2
DATABASE_NAME_PREFIX="ecommerce-" # Match the container name prefix

# Function to run migrations for a single service.
run_migration() {
    local service_path=$1
    local db_name=$2
    local db_port=$3

    echo "--- Running migration for $service_path (Database: $db_name) ---"
    
    if [ ! -d "$service_path/migrations" ]; then
        echo "Warning: No migrations directory found for $service_path. Skipping."
        return 0
    fi

    # Connect to the correct database instance using the container name and exposed port.
    # We use the host.docker.internal to connect from the host machine to the postgres container.
    # On Linux, you might need to use a different method like --add-host=host.docker.internal:host-gateway
    # Or simply connect to localhost if the port is exposed.
    # For simplicity and cross-platform compatibility, we'll use localhost since the port is exposed.
    
    export DATABASE_URL="postgres://postgres:postgres123@localhost:5432/$db_name?sslmode=disable"
    
    # Use the 'goose' tool to run migrations.
    # Assumes you have goose installed globally on your machine: go install github.com/pressly/goose/v3/cmd/goose@latest
    goose -dir "$service_path/migrations" postgres "$DATABASE_URL" "$COMMAND"
    echo "--- Finished migration for $service_path ---"
    echo ""
}

# Define database names for each service
case "$SERVICE_NAME" in
    "user-service")
        run_migration "services/user-service" "users_db" 5432
        ;;
    "product-service")
        run_migration "services/product-service" "products_db" 5432
        ;;
    "order-service")
        run_migration "services/order-service" "orders_db" 5432
        ;;
    "payment-service")
        run_migration "services/payment-service" "payments_db" 5432
        ;;
    "inventory-service")
        run_migration "services/inventory-service" "inventory_db" 5432
        ;;
    "notification-service")
        run_migration "services/notification-service" "notifications_db" 5432
        ;;
    "all")
        run_migration "services/user-service" "users_db" 5432
        run_migration "services/product-service" "products_db" 5432
        run_migration "services/order-service" "orders_db" 5432
        run_migration "services/payment-service" "payments_db" 5432
        run_migration "services/inventory-service" "inventory_db" 5432
        run_migration "services/notification-service" "notifications_db" 5432
        ;;
    *)
        echo "Error: Invalid service name '$SERVICE_NAME'"
        usage
        ;;
esac

echo "Migration script finished."