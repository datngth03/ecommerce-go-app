-- migrations/payment_service/V<timestamp>_create_payments_table.down.sql

-- Xóa ràng buộc khóa ngoại trước khi xóa bảng
ALTER TABLE payments
DROP CONSTRAINT IF EXISTS fk_order_id;

-- Xóa bảng payments
DROP TABLE IF EXISTS payments;
