-- migrations/shipping_service/V<timestamp>_create_shipments_table.down.sql

-- Xóa ràng buộc khóa ngoại trước khi xóa bảng
ALTER TABLE shipments
DROP CONSTRAINT IF EXISTS fk_order_id;

-- Xóa bảng shipments
DROP TABLE IF EXISTS shipments;
