-- migrations/order_service/V<timestamp>_create_orders_table.up.sql

-- Tạo bảng orders
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL, -- ID người dùng tạo đơn hàng
    total_amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- Trạng thái đơn hàng: pending, paid, shipped, cancelled, etc.
    shipping_address TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho user_id và status để tìm kiếm/lọc nhanh hơn
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders (user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders (status);

-- Bảng để lưu trữ các sản phẩm trong mỗi đơn hàng (Order Items)
-- Sử dụng JSONB để lưu trữ linh hoạt các chi tiết của OrderItem
CREATE TABLE IF NOT EXISTS order_items (
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    quantity INT NOT NULL,
    PRIMARY KEY (order_id, product_id), -- Khóa chính kép
    CONSTRAINT fk_order
        FOREIGN KEY(order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE -- Xóa các item khi đơn hàng bị xóa
);

-- Tạo index cho order_id
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);