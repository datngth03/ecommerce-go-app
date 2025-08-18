#!/bin/sh
#
# This entrypoint script simply runs `golang-migrate` on the
# pre-consolidated migrations directory created by the Dockerfile.
#

set -e

MIGRATE_BIN="/usr/local/bin/migrate"

echo "Waiting for database to be ready..."

until pg_isready -U $DB_USER -d $DB_NAME -h core_db -p 5432; do
  echo "Database not ready yet, sleeping..."
  sleep 2
done

echo "Applying consolidated database migrations from ./dist/migrations..."
# The `migrate` command now points to the consolidated directory
$MIGRATE_BIN -path ./dist/migrations -database "$DATABASE_URL" up

echo "All migrations applied successfully!"


