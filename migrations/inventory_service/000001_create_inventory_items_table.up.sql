-- migrations/inventory_service/V<timestamp>_create_inventory_items_table.up.sql

-- Tạo bảng inventory_items
CREATE TABLE IF NOT EXISTS inventory_items (
    product_id UUID PRIMARY KEY, -- ID sản phẩm làm khóa chính
    quantity INT NOT NULL DEFAULT 0, -- Số lượng có sẵn
    reserved_quantity INT NOT NULL DEFAULT 0, -- Số lượng đã đặt trước
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho product_id (đã là PRIMARY KEY, nên không cần index riêng)
