-- migrations/order_service/V<timestamp>_create_orders_table.down.sql

-- Xóa bảng order_items trước vì nó có khóa ngoại đến bảng orders
DROP TABLE IF EXISTS order_items;

-- Xóa bảng orders
DROP TABLE IF EXISTS orders;
