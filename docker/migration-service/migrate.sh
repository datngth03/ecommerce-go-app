#!/bin/bash

MIGRATE_BIN="/usr/local/bin/migrate"

echo "Waiting for database to be ready..."

# Extract database connection details from DATABASE_URL if needed
# Or use individual env vars

echo "Applying database migrations from $MIGRATION_PATH..."

# Run migrations using the migrate CLI tool
$MIGRATE_BIN -path $MIGRATION_PATH -database "$DATABASE_URL" up

if [ $? -eq 0 ]; then
    echo "All migrations applied successfully!"
else
    echo "Migration failed!"
    exit 1
fi