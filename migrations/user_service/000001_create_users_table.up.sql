-- migrations/user_service/V<timestamp>_create_users_table.up.sql

-- Tạo bảng users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY, -- Sử dụng UUID làm khóa chính
    email VARCHAR(255) NOT NULL UNIQUE, -- Email phải là duy nhất
    password VARCHAR(255) NOT NULL, -- Mật khẩu đã băm
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho cột email để tìm kiếm nhanh hơn
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
