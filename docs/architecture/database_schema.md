# E-commerce Database Schema

## Database per Service Pattern

```
users_db → products_db → orders_db → payments_db → inventory_db
```

## 1. Users Service (users_db)

### users
```sql
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,
    name       VARCHAR(100) NOT NULL,
    phone      VARCHAR(20),
    role       VARCHAR(20) DEFAULT 'customer',
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 2. Products Service (products_db)

### categories
```sql
CREATE TABLE categories (
    id   BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL
);
```

### products
```sql
CREATE TABLE products (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    price       DECIMAL(10,2) NOT NULL,
    category_id BIGINT REFERENCES categories(id),
    image_url   VARCHAR(255),
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMP DEFAULT NOW()
);
```

## 3. Orders Service (orders_db)

### orders
```sql
CREATE TABLE orders (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL,
    status       VARCHAR(20) DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL,
    created_at   TIMESTAMP DEFAULT NOW()
);
```

### order_items
```sql
CREATE TABLE order_items (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT REFERENCES orders(id),
    product_id BIGINT NOT NULL,
    quantity   INTEGER NOT NULL,
    price      DECIMAL(10,2) NOT NULL
);
```

## 4. Payments Service (payments_db)

### payments
```sql
CREATE TABLE payments (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT NOT NULL,
    amount     DECIMAL(10,2) NOT NULL,
    status     VARCHAR(20) DEFAULT 'pending',
    method     VARCHAR(50) NOT NULL,
    stripe_id  VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 5. Inventory Service (inventory_db)

### inventory
```sql
CREATE TABLE inventory (
    id         BIGSERIAL PRIMARY KEY,
    product_id BIGINT UNIQUE NOT NULL,
    quantity   INTEGER NOT NULL DEFAULT 0,
    reserved   INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## 6. Notifications Service

### notifications
```sql
CREATE TABLE notifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    type       VARCHAR(50) NOT NULL,    -- 'email', 'sms'
    subject    VARCHAR(255),
    content    TEXT NOT NULL,
    recipient  VARCHAR(255) NOT NULL,   -- email hoặc phone
    status     VARCHAR(20) DEFAULT 'pending', -- 'pending', 'sent', 'failed'
    sent_at    TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Cross-Service Communication

Services communicate via:
- **HTTP APIs**: Client requests
- **gRPC**: Service-to-service calls  
- **Events**: Order created, Payment processed

## Key Design Decisions

1. **Simple Schema**: Focus on core functionality
2. **Logical References**: Store IDs, no cross-DB foreign keys
3. **Essential Indexes**: Only on frequently queried columns
4. **Money as DECIMAL**: Avoid floating point issues
5. **Timestamps**: Track creation time

## Example Flow

```
1. User creates order → orders_db
2. Order service calls inventory → inventory_db (reserve stock)
3. Order service calls payment → payments_db (process payment)
4. Success → Update order status
```

This keeps it simple while maintaining microservices principles!