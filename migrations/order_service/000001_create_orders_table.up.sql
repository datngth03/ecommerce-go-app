-- migrations/order_service/V<timestamp>_create_orders_table.up.sql

-- Tạo bảng orders
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(12,2) NOT NULL,
    subtotal_amount DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) DEFAULT 0,
    shipping_fee DECIMAL(12,2) DEFAULT 0,
    tax_amount DECIMAL(12,2) DEFAULT 0,
    shipping_address JSONB NOT NULL,
    billing_address JSONB,
    payment_method VARCHAR(50),
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    shipping_status VARCHAR(20) DEFAULT 'pending',
    placed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    canceled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Khóa ngoại đến bảng users
    -- Giả sử user_service và order_service có thể giao tiếp qua một message queue hoặc API,
    -- nhưng trong migration, chúng ta không tạo khóa ngoại trực tiếp giữa các service.
    -- Bạn có thể giữ cột này như một tham chiếu logic.
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
);

-- Index để tìm kiếm đơn hàng theo người dùng và số đơn hàng
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders (order_number);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

-- Bảng để lưu trữ các sản phẩm trong mỗi đơn hàng (Order Items)
-- Sử dụng JSONB để lưu trữ linh hoạt các chi tiết của OrderItem
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    variant_id UUID,
    variant_sku VARCHAR(100) NOT NULL,
    variant_attributes JSONB NOT NULL,
    specifications JSONB,
    quantity INT NOT NULL,
    original_price DECIMAL(12, 2) NOT NULL,
    discount DECIMAL(5,2),
    unit_price DECIMAL(12,2) NOT NULL,
    total_price DECIMAL(12,2) NOT NULL,

    -- Khóa ngoại đến bảng orders
    CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
    -- Lưu ý: product_id và variant_id là các tham chiếu logic đến product_service.
    -- Không tạo khóa ngoại trực tiếp giữa các service.
);

-- Index để tìm kiếm các mục trong một đơn hàng cụ thể
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
