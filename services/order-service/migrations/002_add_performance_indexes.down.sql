-- Rollback performance indexes for Order Service

DROP INDEX IF EXISTS idx_orders_summary;
DROP INDEX IF EXISTS idx_orders_user_count;
DROP INDEX IF EXISTS idx_orders_date_status_amount;
DROP INDEX IF EXISTS idx_cart_items_updated_at;
DROP INDEX IF EXISTS idx_cart_items_product_count;
DROP INDEX IF EXISTS idx_cart_items_cart_quantity;
DROP INDEX IF EXISTS idx_carts_user_updated;
DROP INDEX IF EXISTS idx_carts_updated_at;
DROP INDEX IF EXISTS idx_order_items_quantity;
DROP INDEX IF EXISTS idx_order_items_product_created;
DROP INDEX IF EXISTS idx_order_items_order_product;
DROP INDEX IF EXISTS idx_orders_payment_method;
DROP INDEX IF EXISTS idx_orders_active;
DROP INDEX IF EXISTS idx_orders_pending;
DROP INDEX IF EXISTS idx_orders_status_created;
DROP INDEX IF EXISTS idx_orders_created_at_date;
DROP INDEX IF EXISTS idx_orders_total_amount;
DROP INDEX IF EXISTS idx_orders_user_status_created;
DROP INDEX IF EXISTS idx_orders_user_created_desc;
DROP INDEX IF EXISTS idx_orders_user_status;
