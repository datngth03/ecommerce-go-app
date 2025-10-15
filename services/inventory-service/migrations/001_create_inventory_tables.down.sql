-- Drop triggers
DROP TRIGGER IF EXISTS update_reservations_updated_at ON reservations;
DROP TRIGGER IF EXISTS update_stocks_updated_at ON stocks;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_reservations_expires_at;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_product_id;
DROP INDEX IF EXISTS idx_reservations_order_id;
DROP INDEX IF EXISTS idx_stock_movements_created_at;
DROP INDEX IF EXISTS idx_stock_movements_product_id;
DROP INDEX IF EXISTS idx_stocks_warehouse_id;
DROP INDEX IF EXISTS idx_stocks_product_id;

-- Drop tables
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS stocks;
