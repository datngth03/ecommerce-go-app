
-- migrations/product_service/V<timestamp>_create_categories_table.up.sql

-- Tạo bảng categories
CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE, -- Tên danh mục phải là duy nhất
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho cột name để tìm kiếm nhanh hơn
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories (name);

