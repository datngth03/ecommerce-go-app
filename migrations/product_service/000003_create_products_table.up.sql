-- migrations/product_service/V<timestamp>_create_products_table.up.sql

-- Tạo bảng products
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL, -- Giá sản phẩm với 2 chữ số thập phân
    category_id UUID NOT NULL, -- ID danh mục, khóa ngoại đến bảng categories
    image_urls TEXT[], -- Mảng các URL hình ảnh
    stock_quantity INT DEFAULT 0, -- Số lượng tồn kho ban đầu
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Thêm ràng buộc khóa ngoại đến bảng categories
    CONSTRAINT fk_category
        FOREIGN KEY(category_id)
        REFERENCES categories(id)
        ON DELETE RESTRICT -- Không cho phép xóa danh mục nếu có sản phẩm liên quan
);

-- Tạo index cho cột name và category_id để tìm kiếm/lọc nhanh hơn
CREATE INDEX IF NOT EXISTS idx_products_name ON products (name);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products (category_id);
