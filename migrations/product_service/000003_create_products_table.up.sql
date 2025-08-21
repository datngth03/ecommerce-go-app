CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    brand_id UUID,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    rating DECIMAL(3,2) DEFAULT 0,
    review_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    
);

-- Tạo index cho cột name và category_id để tìm kiếm/lọc nhanh hơn
CREATE INDEX IF NOT EXISTS idx_products_name ON products (name);
