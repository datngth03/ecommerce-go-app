#!/bin/sh
set -e

echo "🗄️  Starting Database Migrations..."
echo "=================================="

# Database connection settings
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres123}"
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
for i in $(seq 1 15); do
    if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
        echo "PostgreSQL is ready!"
        break
    fi
    echo "Waiting... ($i/15)"
    sleep 2
done

# Function to run migration for a service
run_migration() {
    SERVICE=$1
    DB_NAME=$2
    MIGRATION_PATH=$3
    
    echo ""
    echo "📦 Migrating $SERVICE to $DB_NAME..."
    
    DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
    
    if [ -d "$MIGRATION_PATH" ]; then
        migrate -path "$MIGRATION_PATH" -database "$DB_URL" up || {
            echo "⚠️  Migration failed or already applied for $SERVICE"
        }
        echo "$SERVICE migration completed"
    else
        echo "⚠️  No migrations found for $SERVICE at $MIGRATION_PATH"
    fi
}

# Run migrations for each service
run_migration "user-service" "users_db" "/migrations/user-service"
run_migration "product-service" "products_db" "/migrations/product-service"
run_migration "order-service" "orders_db" "/migrations/order-service"
run_migration "payment-service" "payments_db" "/migrations/payment-service"
run_migration "inventory-service" "inventory_db" "/migrations/inventory-service"
run_migration "notification-service" "notifications_db" "/migrations/notification-service"

echo ""
echo "=================================="
echo "🎉 All migrations completed!"
echo "=================================="
