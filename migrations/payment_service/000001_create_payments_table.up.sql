-- migrations/payment_service/V<timestamp>_create_payments_table.up.sql

-- Tạo bảng payments

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    method VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    transaction_id VARCHAR(100),
    amount DECIMAL(12,2) NOT NULL,
    currency VARCHAR(3) NOT NULL, -- Ví dụ: USD, VND
    provider_response JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index để tìm kiếm thanh toán theo order_id và transaction_id
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments (order_id);
CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments (transaction_id);
