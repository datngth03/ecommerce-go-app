# Stage 1: Build the `migrate` tool binary from source
FROM golang:1.23-alpine AS migrator-builder

WORKDIR /go/src/app

# Get `golang-migrate` and build the binary
RUN apk add --no-cache git && \
    git clone https://github.com/golang-migrate/migrate.git && \
    cd migrate/cmd/migrate && \
    go build -tags 'postgres' -ldflags="-s -w" -o /go/bin/migrate

# Stage 2: Create the final lean image for running migrations
FROM alpine:latest

# Install postgres client (required for some versions of migrate, and useful for debugging)
RUN apk add --no-cache postgresql-client

WORKDIR /app

# Copy the pre-built `migrate` tool
COPY --from=migrator-builder /go/bin/migrate /usr/local/bin/migrate

# Copy all the separate migration directories
COPY migrations/ ./migrations/

# Copy the migration consolidation script and make it executable
COPY docker/migration-service/merge-migrations.sh /usr/local/bin/merge-migrations.sh
RUN chmod +x /usr/local/bin/merge-migrations.sh

# This is the key step: run the script during the build process
# to create the final, ordered migrations directory
RUN /usr/local/bin/merge-migrations.sh

# Copy the final entrypoint script
COPY docker/migration-service/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# The entrypoint will now run on the consolidated migrations
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
