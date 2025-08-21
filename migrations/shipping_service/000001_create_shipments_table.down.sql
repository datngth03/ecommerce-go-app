-- migrations/shipping_service/V<timestamp>_create_shipments_table.down.sql

-- Xóa bảng shipments
DROP TABLE IF EXISTS shipments;
DROP TABLE IF EXISTS shipping_providers;