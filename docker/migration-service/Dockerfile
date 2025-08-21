# Use Alpine Linux as base image
FROM alpine:latest

# Install required packages
RUN apk add --no-cache \
    postgresql-client \
    curl \
    bash

# Install golang-migrate CLI tool
RUN curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Create non-root user
RUN addgroup -g 1001 -S migration && \
    adduser -S migration -u 1001 -G migration

# Set working directory
WORKDIR /app

# Copy migration files
COPY migrations ./migrations

# Copy migration script
COPY docker/migration-service/migrate.sh ./migrate.sh

# Make script executable and change ownership
RUN chmod +x ./migrate.sh && \
    chown -R migration:migration /app

# Switch to non-root user
USER migration

# Set environment variables with defaults
ENV DATABASE_URL=""
ENV MIGRATION_PATH="./migrations"
ENV DB_USER=""
ENV DB_NAME=""
ENV DB_HOST="localhost"
ENV DB_PORT="5432"

# Run the migration script
CMD ["./migrate.sh"]