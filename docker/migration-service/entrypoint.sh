#!/bin/sh
#
# This entrypoint script simply runs `golang-migrate` on the
# pre-consolidated migrations directory created by the Dockerfile.
#

set -e

MIGRATE_BIN="/usr/local/bin/migrate"

echo "Applying consolidated database migrations from ./dist/migrations..."
# The `migrate` command now points to the consolidated directory
$MIGRATE_BIN -path ./dist/migrations -database "$DATABASE_URL" up

echo "All migrations applied successfully!"
