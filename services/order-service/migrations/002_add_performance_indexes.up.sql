-- Add advanced performance indexes for Order Service
-- These indexes optimize order queries, reporting, and analytics

-- ===== ORDERS TABLE INDEXES =====

-- Composite index for user orders by status (common query pattern)
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON orders(user_id, status);

-- Composite index for user orders sorted by date
CREATE INDEX IF NOT EXISTS idx_orders_user_created_desc ON orders(user_id, created_at DESC);

-- Composite index for user orders by status and date
CREATE INDEX IF NOT EXISTS idx_orders_user_status_created ON orders(user_id, status, created_at DESC);

-- Index for order total amount (for revenue queries)
CREATE INDEX IF NOT EXISTS idx_orders_total_amount ON orders(total_amount);

-- Index for orders by date range (analytics)
CREATE INDEX IF NOT EXISTS idx_orders_created_at_date ON orders(DATE(created_at));

-- Composite index for status and date (order management dashboard)
CREATE INDEX IF NOT EXISTS idx_orders_status_created ON orders(status, created_at DESC);

-- Partial index for pending orders (hot data)
CREATE INDEX IF NOT EXISTS idx_orders_pending ON orders(user_id, created_at DESC) 
WHERE status = 'pending';

-- Partial index for active orders (not delivered/cancelled)
CREATE INDEX IF NOT EXISTS idx_orders_active ON orders(user_id, status, created_at DESC) 
WHERE status NOT IN ('delivered', 'cancelled');

-- Index for payment method analytics
CREATE INDEX IF NOT EXISTS idx_orders_payment_method ON orders(payment_method, created_at DESC);

-- ===== ORDER_ITEMS TABLE INDEXES =====

-- Composite index for order items with product info
CREATE INDEX IF NOT EXISTS idx_order_items_order_product ON order_items(order_id, product_id);

-- Index for product sales analytics
CREATE INDEX IF NOT EXISTS idx_order_items_product_created ON order_items(product_id, created_at DESC);

-- Index for quantity analysis
CREATE INDEX IF NOT EXISTS idx_order_items_quantity ON order_items(quantity);

-- ===== CARTS TABLE INDEXES =====

-- Index for cart age (cleanup old carts)
CREATE INDEX IF NOT EXISTS idx_carts_updated_at ON carts(updated_at);

-- Composite index for active carts
CREATE INDEX IF NOT EXISTS idx_carts_user_updated ON carts(user_id, updated_at DESC);

-- ===== CART_ITEMS TABLE INDEXES =====

-- Composite index for cart items with quantities
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_quantity ON cart_items(cart_id, quantity);

-- Index for product popularity in carts
CREATE INDEX IF NOT EXISTS idx_cart_items_product_count ON cart_items(product_id, created_at DESC);

-- Index for cart item updates
CREATE INDEX IF NOT EXISTS idx_cart_items_updated_at ON cart_items(updated_at);

-- ===== REPORTING & ANALYTICS INDEXES =====

-- Composite index for daily revenue reports
CREATE INDEX IF NOT EXISTS idx_orders_date_status_amount ON orders(DATE(created_at), status, total_amount);

-- Index for order count by user (customer analytics)
CREATE INDEX IF NOT EXISTS idx_orders_user_count ON orders(user_id, id);

-- Covering index for order summary queries (includes commonly selected columns)
CREATE INDEX IF NOT EXISTS idx_orders_summary ON orders(user_id, status, created_at DESC) 
INCLUDE (total_amount, payment_method);

-- Add comments for documentation
COMMENT ON INDEX idx_orders_user_status IS 'Optimizes user order queries filtered by status';
COMMENT ON INDEX idx_orders_user_created_desc IS 'Optimizes user order history queries';
COMMENT ON INDEX idx_orders_status_created IS 'Optimizes order management dashboard';
COMMENT ON INDEX idx_orders_pending IS 'Partial index for fast pending order lookup';
COMMENT ON INDEX idx_orders_active IS 'Partial index for active orders (excludes completed)';
COMMENT ON INDEX idx_orders_date_status_amount IS 'Optimizes daily revenue reports';
COMMENT ON INDEX idx_order_items_product_created IS 'Optimizes product sales analytics';
COMMENT ON INDEX idx_cart_items_product_count IS 'Optimizes cart abandonment analysis';
COMMENT ON INDEX idx_orders_summary IS 'Covering index for order list queries (reduces heap access)';
