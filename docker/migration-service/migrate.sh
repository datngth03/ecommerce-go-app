#!/bin/bash

MIGRATE_BIN="/usr/local/bin/migrate"

echo "Waiting for database to be ready..."

# Extract host, port, user from DATABASE_URL
# ví dụ: postgres://inventory_service:inventory_password_123@inventory_db:5432/inventory_service_db?sslmode=disable
DB_HOST=$(echo $DATABASE_URL | sed -E 's/.*@([^:]+):.*/\1/')
DB_PORT=$(echo $DATABASE_URL | sed -E 's/.*:([0-9]+).*/\1/')
DB_USER=$(echo $DATABASE_URL | sed -E 's/.*:\/\/(.*):.*@.*/\1/')

# Wait for Postgres to be ready
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; do
    echo "Postgres is unavailable - sleeping"
    sleep 1
done

echo "Postgres is ready. Applying database migrations..."

# Run migrations using the migrate CLI tool
$MIGRATE_BIN -path $MIGRATION_PATH -database "$DATABASE_URL" up

if [ $? -eq 0 ]; then
    echo "All migrations applied successfully!"
else
    echo "Migration failed!"
    exit 1
fi