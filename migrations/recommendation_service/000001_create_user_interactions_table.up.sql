-- migrations/recommendation_service/V<timestamp>_create_user_interactions_table.up.sql

-- Tạo bảng user_interactions để lưu trữ tương tác của người dùng với sản phẩm
CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    product_id UUID NOT NULL,
    score DECIMAL(5,2) NOT NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index để tìm kiếm gợi ý cho người dùng cụ thể
CREATE INDEX IF NOT EXISTS idx_recommendations_user_id ON recommendations (user_id);