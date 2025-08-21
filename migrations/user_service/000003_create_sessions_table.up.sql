-- Tạo bảng sessions
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    refresh_token VARCHAR(512) NOT NULL,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
    
    -- Khóa ngoại liên kết đến bảng users
);

-- Tạo index để tìm kiếm session theo người dùng
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);