# Database Schema Documentation

## Overview

This document describes the database schema for all microservices in the E-Commerce platform. Each service has its own isolated PostgreSQL database following the **Database-per-Service** pattern.

## Database Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    PostgreSQL Cluster                           │
├─────────────────────────────────────────────────────────────────┤
│  users_db  │  products_db  │  orders_db  │  payments_db  │     │
│  inventory_db  │  notifications_db                              │
└─────────────────────────────────────────────────────────────────┘
```

---

## 1. User Service Database (`users_db`)

### Tables

#### `users`
Stores user account information and credentials.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PRIMARY KEY | Auto-incrementing user ID |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email (login) |
| password_hash | VARCHAR(255) | NOT NULL | Bcrypt hashed password |
| name | VARCHAR(100) | NOT NULL | User full name |
| phone | VARCHAR(20) | | Phone number (optional) |
| is_active | BOOLEAN | DEFAULT TRUE | Account status |
| created_at | TIMESTAMP | DEFAULT NOW() | Account creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |

**Indexes:**
- `idx_users_email` on `email`
- `idx_users_created_at` on `created_at`

**Default Data:**
- Admin user: `admin@example.com` / `Admin123!`

---

## 2. Product Service Database (`products_db`)

### Tables

#### `categories`
Product categories for organization.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Category UUID |
| name | VARCHAR(100) | UNIQUE, NOT NULL | Category name |
| slug | VARCHAR(120) | UNIQUE, NOT NULL | URL-friendly slug |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |

**Indexes:**
- `idx_categories_slug` on `slug`

**Default Categories:**
- Electronics, Clothing, Books, Home & Garden, Sports

#### `products`
Product catalog with details and pricing.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Product UUID |
| name | VARCHAR(255) | NOT NULL | Product name |
| slug | VARCHAR(280) | NOT NULL | URL-friendly slug |
| description | TEXT | | Product description |
| price | DECIMAL(10,2) | NOT NULL, CHECK > 0 | Product price |
| category_id | UUID | FK → categories(id), NOT NULL | Category reference |
| image_url | VARCHAR(500) | | Product image URL |
| is_active | BOOLEAN | DEFAULT TRUE | Product availability |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |

**Indexes:**
- `idx_products_category_id` on `category_id`
- `idx_products_slug` on `slug`
- `idx_products_is_active` on `is_active`
- `idx_products_created_at` on `created_at`

**Triggers:**
- Auto-update `updated_at` on row modification

---

## 3. Order Service Database (`orders_db`)

### Tables

#### `carts`
Shopping carts for users.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Cart UUID |
| user_id | BIGINT | UNIQUE, NOT NULL | User reference |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Cart creation |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update |

**Indexes:**
- `idx_carts_user_id` on `user_id`

#### `cart_items`
Items in shopping carts.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Cart item UUID |
| cart_id | UUID | FK → carts(id), NOT NULL | Cart reference |
| product_id | VARCHAR(255) | NOT NULL | Product reference |
| product_name | VARCHAR(255) | NOT NULL | Cached product name |
| quantity | INTEGER | CHECK > 0, NOT NULL | Item quantity |
| price | DECIMAL(12,2) | CHECK >= 0, NOT NULL | Item price |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Item added time |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update |

**Unique Constraint:** `(cart_id, product_id)`

**Indexes:**
- `idx_cart_items_cart_id` on `cart_id`
- `idx_cart_items_product_id` on `product_id`

#### `orders`
Customer orders.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Order UUID |
| user_id | BIGINT | NOT NULL | User reference |
| status | VARCHAR(50) | DEFAULT 'pending', NOT NULL | Order status |
| total_amount | DECIMAL(12,2) | DEFAULT 0.00, NOT NULL | Total order amount |
| shipping_address | TEXT | NOT NULL | Delivery address |
| payment_method | VARCHAR(50) | NOT NULL | Payment method used |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Order creation |
| updated_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Last update |

**Status Values:** `pending`, `confirmed`, `processing`, `shipped`, `delivered`, `cancelled`

**Indexes:**
- `idx_orders_user_id` on `user_id`
- `idx_orders_status` on `status`
- `idx_orders_created_at` on `created_at DESC`

#### `order_items`
Items within each order.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Order item UUID |
| order_id | UUID | FK → orders(id) CASCADE, NOT NULL | Order reference |
| product_id | VARCHAR(255) | NOT NULL | Product reference |
| product_name | VARCHAR(255) | NOT NULL | Cached product name |
| quantity | INTEGER | CHECK > 0, NOT NULL | Item quantity |
| price | DECIMAL(12,2) | CHECK >= 0, NOT NULL | Item price at order time |
| subtotal | DECIMAL(12,2) | CHECK >= 0, NOT NULL | Item subtotal |
| created_at | TIMESTAMP | DEFAULT CURRENT_TIMESTAMP | Item creation |

**Indexes:**
- `idx_order_items_order_id` on `order_id`
- `idx_order_items_product_id` on `product_id`

---

## 4. Payment Service Database (`payments_db`)

### Tables

#### `payments`
Payment records for orders.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Payment UUID |
| order_id | VARCHAR(255) | NOT NULL | Order reference |
| user_id | VARCHAR(255) | NOT NULL | User reference |
| amount | DECIMAL(10,2) | CHECK > 0, NOT NULL | Payment amount |
| currency | VARCHAR(3) | DEFAULT 'USD', NOT NULL | Currency code |
| status | VARCHAR(50) | NOT NULL | Payment status |
| method | VARCHAR(50) | NOT NULL | Payment method |
| gateway_payment_id | VARCHAR(255) | | External payment ID (Stripe) |
| gateway_customer_id | VARCHAR(255) | | External customer ID |
| failure_reason | TEXT | | Failure reason if failed |
| metadata | JSONB | | Additional payment metadata |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Payment creation |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |
| deleted_at | TIMESTAMP WITH TIME ZONE | | Soft delete timestamp |

**Indexes:**
- `idx_payments_order_id` on `order_id`
- `idx_payments_user_id` on `user_id`
- `idx_payments_status` on `status`
- `idx_payments_gateway_payment_id` on `gateway_payment_id`

#### `transactions`
Transaction history and gateway responses.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Transaction UUID |
| payment_id | UUID | FK → payments(id) CASCADE, NOT NULL | Payment reference |
| transaction_type | VARCHAR(50) | NOT NULL | Transaction type |
| amount | DECIMAL(10,2) | NOT NULL | Transaction amount |
| status | VARCHAR(50) | NOT NULL | Transaction status |
| gateway_response | JSONB | | Raw gateway response |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Transaction time |
| deleted_at | TIMESTAMP WITH TIME ZONE | | Soft delete |

**Indexes:**
- `idx_transactions_payment_id` on `payment_id`

#### `refunds`
Refund records for payments.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Refund UUID |
| payment_id | UUID | FK → payments(id) CASCADE, NOT NULL | Payment reference |
| amount | DECIMAL(10,2) | CHECK > 0, NOT NULL | Refund amount |
| reason | TEXT | | Refund reason |
| status | VARCHAR(50) | NOT NULL | Refund status |
| gateway_refund_id | VARCHAR(255) | | External refund ID |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Refund creation |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |
| deleted_at | TIMESTAMP WITH TIME ZONE | | Soft delete |

**Indexes:**
- `idx_refunds_payment_id` on `payment_id`
- `idx_refunds_status` on `status`

#### `payment_methods`
Saved payment methods for users.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Payment method UUID |
| user_id | VARCHAR(255) | NOT NULL | User reference |
| method_type | VARCHAR(50) | NOT NULL | Method type (card, etc) |
| last4 | VARCHAR(4) | | Last 4 digits of card |
| brand | VARCHAR(50) | | Card brand (Visa, etc) |
| gateway_method_id | VARCHAR(255) | NOT NULL | External method ID |
| is_default | BOOLEAN | DEFAULT FALSE | Default payment method |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |
| deleted_at | TIMESTAMP WITH TIME ZONE | | Soft delete |

**Indexes:**
- `idx_payment_methods_user_id` on `user_id`

---

## 5. Inventory Service Database (`inventory_db`)

### Tables

#### `stocks`
Current stock levels per product.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Stock record UUID |
| product_id | VARCHAR(255) | UNIQUE, NOT NULL | Product reference |
| available | INTEGER | CHECK >= 0, DEFAULT 0, NOT NULL | Available quantity |
| reserved | INTEGER | CHECK >= 0, DEFAULT 0, NOT NULL | Reserved quantity |
| total | INTEGER | CHECK >= 0, DEFAULT 0, NOT NULL | Total quantity |
| warehouse_id | VARCHAR(255) | DEFAULT 'main', NOT NULL | Warehouse identifier |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |

**Constraints:**
- `total_equals_sum`: `total = available + reserved`

**Indexes:**
- `idx_stocks_product_id` on `product_id`
- `idx_stocks_warehouse_id` on `warehouse_id`

#### `stock_movements`
Audit trail of all stock changes.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Movement UUID |
| product_id | VARCHAR(255) | FK → stocks(product_id) CASCADE, NOT NULL | Product reference |
| movement_type | VARCHAR(50) | NOT NULL | Movement type |
| quantity | INTEGER | NOT NULL | Quantity changed |
| before_quantity | INTEGER | NOT NULL | Quantity before |
| after_quantity | INTEGER | NOT NULL | Quantity after |
| reference_type | VARCHAR(50) | NOT NULL | Reference type (order, etc) |
| reference_id | VARCHAR(255) | | Reference ID |
| reason | TEXT | | Change reason |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Movement time |

**Indexes:**
- `idx_stock_movements_product_id` on `product_id`
- `idx_stock_movements_created_at` on `created_at DESC`

#### `reservations`
Temporary stock holds for pending orders.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Reservation UUID |
| order_id | VARCHAR(255) | NOT NULL | Order reference |
| product_id | VARCHAR(255) | FK → stocks(product_id) CASCADE, NOT NULL | Product reference |
| quantity | INTEGER | CHECK > 0, NOT NULL | Reserved quantity |
| status | VARCHAR(50) | DEFAULT 'PENDING', NOT NULL | Reservation status |
| expires_at | TIMESTAMP WITH TIME ZONE | NOT NULL | Expiration time |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |

**Status Values:** `PENDING`, `CONFIRMED`, `EXPIRED`, `CANCELLED`

**Indexes:**
- `idx_reservations_order_id` on `order_id`
- `idx_reservations_product_id` on `product_id`
- `idx_reservations_status` on `status`
- `idx_reservations_expires_at` on `expires_at`

---

## 6. Notification Service Database (`notifications_db`)

### Tables

#### `notifications`
Notification queue and history.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Notification UUID |
| user_id | VARCHAR(255) | | User reference (optional) |
| type | VARCHAR(50) | NOT NULL | Notification type |
| channel | VARCHAR(50) | NOT NULL | Delivery channel |
| recipient | VARCHAR(255) | NOT NULL | Recipient address |
| subject | VARCHAR(500) | | Email subject |
| content | TEXT | NOT NULL | Notification content |
| status | VARCHAR(50) | NOT NULL | Delivery status |
| error_message | TEXT | | Error details if failed |
| template_id | UUID | | Template reference |
| metadata | JSONB | | Additional metadata |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| sent_at | TIMESTAMP WITH TIME ZONE | | Sent timestamp |
| deleted_at | TIMESTAMP WITH TIME ZONE | | Soft delete |

**Channel Values:** `email`, `sms`, `push`
**Status Values:** `pending`, `sent`, `failed`, `retrying`

**Indexes:**
- `idx_notifications_user_id` on `user_id`
- `idx_notifications_type` on `type`
- `idx_notifications_status` on `status`
- `idx_notifications_created_at` on `created_at DESC`

#### `templates`
Reusable notification templates.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PRIMARY KEY | Template UUID |
| name | VARCHAR(255) | UNIQUE, NOT NULL | Template name |
| type | VARCHAR(50) | NOT NULL | Template type |
| subject | VARCHAR(500) | | Email subject template |
| body | TEXT | NOT NULL | Template body |
| variables | JSONB | | Available variables |
| is_active | BOOLEAN | DEFAULT TRUE | Template status |
| created_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | DEFAULT CURRENT_TIMESTAMP | Last update |
| deleted_at | TIMESTAMP WITH TIME ZONE | | Soft delete |

**Indexes:**
- `idx_templates_name` on `name`
- `idx_templates_type` on `type`

---

## Database Relationships

### Cross-Service References

Since each service has an independent database, cross-service references are handled via IDs (not foreign keys):

```
users.id (BIGINT) → orders.user_id
                  → payments.user_id
                  → notifications.user_id

products.id (UUID) → order_items.product_id
                   → cart_items.product_id
                   → stocks.product_id

orders.id (UUID) → payments.order_id
                 → reservations.order_id
```

### Eventual Consistency

- Product name changes → Cached in `order_items`, `cart_items`
- Stock updates → Event-driven sync with Order Service
- Payment status → Event triggers Order/Notification services

---

## Database Administration

### Connection Info
```yaml
Host: localhost (docker: ecommerce-postgres)
Port: 5432
User: postgres
Password: postgres123
```

### Backup Strategy
```bash
# Backup all databases
pg_dumpall -h localhost -U postgres > backup_all.sql

# Backup specific database
pg_dump -h localhost -U postgres users_db > users_db_backup.sql
```

### Migrations
Each service manages its own migrations in:
```
services/{service}/migrations/*.sql
```

Run migrations via Docker:
```bash
docker exec -it ecommerce-postgres psql -U postgres -d {db_name} -f /migrations/001_*.sql
```

---

**Last Updated:** October 2025  
**Schema Version:** 1.0.0
