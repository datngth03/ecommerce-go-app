-- migrations/payment_service/V<timestamp>_create_payments_table.up.sql

-- Tạo bảng payments
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL, -- ID đơn hàng liên quan
    user_id UUID NOT NULL, -- ID người dùng thực hiện thanh toán
    amount NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL, -- Ví dụ: USD, VND
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- Trạng thái: pending, completed, failed, refunded
    payment_method VARCHAR(50) NOT NULL, -- Ví dụ: credit_card, paypal, bank_transfer
    transaction_id VARCHAR(255) UNIQUE, -- ID giao dịch từ cổng thanh toán (có thể null ban đầu)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho order_id, user_id và transaction_id để tìm kiếm/lọc nhanh hơn
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments (transaction_id) WHERE transaction_id IS NOT NULL;

-- Thêm ràng buộc khóa ngoại đến bảng orders (nếu orders nằm trong cùng DB)
-- Nếu Order Service có DB riêng, ràng buộc này sẽ không tồn tại ở đây
-- và việc kiểm tra tính toàn vẹn sẽ được thực hiện ở tầng application/service.
-- Giả định hiện tại: orders nằm trong cùng ecommerce_core_db
ALTER TABLE payments
ADD CONSTRAINT fk_order_id
FOREIGN KEY (order_id) REFERENCES orders(id)
ON DELETE RESTRICT; -- Không cho phép xóa đơn hàng nếu có thanh toán liên quan
