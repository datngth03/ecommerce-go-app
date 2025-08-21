CREATE TABLE IF NOT EXISTS categories (
    id UUID PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    slug VARCHAR(150) NOT NULL UNIQUE,
    description TEXT,
    image VARCHAR(255),
    parent_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_parent_category FOREIGN KEY (parent_id) REFERENCES categories (id) ON DELETE CASCADE
);

-- Tạo index cho cột name để tìm kiếm nhanh hơn
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories (name);

