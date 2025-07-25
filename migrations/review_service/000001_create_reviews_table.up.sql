-- migrations/review_service/V<timestamp>_create_reviews_table.up.sql

-- Tạo bảng reviews
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    user_id UUID NOT NULL, -- UUID của người dùng, không cần khóa ngoại nếu User Service là riêng biệt
    rating INT NOT NULL,    -- Đánh giá (ví dụ: 1-5 sao)
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP

    -- Optional: Thêm khóa ngoại nếu bạn muốn ràng buộc cứng với bảng products.
    -- Nếu product_service và review_service dùng chung DB:
    -- CONSTRAINT fk_product
    --     FOREIGN KEY(product_id)
    --     REFERENCES products(id)
    --     ON DELETE CASCADE -- Xóa đánh giá khi sản phẩm bị xóa
);

-- Tạo index để tìm kiếm/lọc nhanh hơn theo product_id và user_id
CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews (product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews (user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews (rating);
