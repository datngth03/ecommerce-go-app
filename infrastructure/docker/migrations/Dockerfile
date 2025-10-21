# Dockerfile for running database migrations
FROM migrate/migrate:v4.17.0

# Copy all migration files
COPY services/user-service/migrations /migrations/user-service
COPY services/product-service/migrations /migrations/product-service
COPY services/order-service/migrations /migrations/order-service
COPY services/payment-service/migrations /migrations/payment-service
COPY services/inventory-service/migrations /migrations/inventory-service
COPY services/notification-service/migrations /migrations/notification-service

# Copy the migration script
COPY infrastructure/docker/migrations/run-migrations.sh /run-migrations.sh
RUN chmod +x /run-migrations.sh

ENTRYPOINT ["/run-migrations.sh"]
