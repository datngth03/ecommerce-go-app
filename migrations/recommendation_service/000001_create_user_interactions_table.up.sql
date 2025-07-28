-- migrations/recommendation_service/V<timestamp>_create_user_interactions_table.up.sql

-- Tạo bảng user_interactions để lưu trữ tương tác của người dùng với sản phẩm
CREATE TABLE IF NOT EXISTS user_interactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- e.g., 'view', 'add_to_cart', 'purchase'
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo các index để tăng tốc độ truy vấn theo user_id và product_id
CREATE INDEX IF NOT EXISTS idx_user_interactions_user_id ON user_interactions (user_id);
CREATE INDEX IF NOT EXISTS idx_user_interactions_product_id ON user_interactions (product_id);
CREATE INDEX IF NOT EXISTS idx_user_interactions_event_type ON user_interactions (event_type);

-- Tạo index tổng hợp để lấy các tương tác gần đây của người dùng
CREATE INDEX IF NOT EXISTS idx_user_interactions_user_product_timestamp ON user_interactions (user_id, product_id, timestamp DESC);
