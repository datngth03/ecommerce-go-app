-- Tạo bảng addresses
CREATE TABLE IF NOT EXISTS addresses (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    name VARCHAR(100),
    address_line VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    zip_code VARCHAR(20),
    country VARCHAR(100) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Khóa ngoại liên kết đến bảng users
    CONSTRAINT fk_user_address FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Tạo index để tìm kiếm địa chỉ theo người dùng
CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses (user_id);