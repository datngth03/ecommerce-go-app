-- Add advanced performance indexes for Inventory Service
-- These indexes optimize stock queries, movement tracking, and reservation management

-- ===== STOCKS TABLE INDEXES =====

-- Index for available stock queries (hot data)
CREATE INDEX IF NOT EXISTS idx_stocks_available ON stocks(available) 
WHERE available > 0;

-- Index for low stock alerts
CREATE INDEX IF NOT EXISTS idx_stocks_low_stock ON stocks(product_id, available) 
WHERE available < 10;

-- Composite index for warehouse inventory
CREATE INDEX IF NOT EXISTS idx_stocks_warehouse_available ON stocks(warehouse_id, available);

-- Index for reserved stock tracking
CREATE INDEX IF NOT EXISTS idx_stocks_reserved ON stocks(reserved) 
WHERE reserved > 0;

-- Composite index for product stock status
CREATE INDEX IF NOT EXISTS idx_stocks_product_status ON stocks(product_id, available, reserved, total);

-- Index for stock updates (temporal queries)
CREATE INDEX IF NOT EXISTS idx_stocks_updated_at ON stocks(updated_at DESC);

-- Partial index for out-of-stock items
CREATE INDEX IF NOT EXISTS idx_stocks_out_of_stock ON stocks(product_id, warehouse_id) 
WHERE available = 0;

-- ===== STOCK_MOVEMENTS TABLE INDEXES =====

-- Composite index for product movement history
CREATE INDEX IF NOT EXISTS idx_stock_movements_product_created ON stock_movements(product_id, created_at DESC);

-- Index for movement type analysis
CREATE INDEX IF NOT EXISTS idx_stock_movements_type ON stock_movements(movement_type, created_at DESC);

-- Composite index for reference tracking
CREATE INDEX IF NOT EXISTS idx_stock_movements_reference ON stock_movements(reference_type, reference_id);

-- Index for quantity changes (audit)
CREATE INDEX IF NOT EXISTS idx_stock_movements_quantity ON stock_movements(quantity, created_at DESC);

-- Composite index for product movement by type
CREATE INDEX IF NOT EXISTS idx_stock_movements_product_type ON stock_movements(product_id, movement_type, created_at DESC);

-- Index for temporal movement queries
CREATE INDEX IF NOT EXISTS idx_stock_movements_date ON stock_movements(DATE(created_at));

-- Index for movement amount analysis
CREATE INDEX IF NOT EXISTS idx_stock_movements_before_after ON stock_movements(before_quantity, after_quantity);

-- ===== RESERVATIONS TABLE INDEXES =====

-- Composite index for order reservations
CREATE INDEX IF NOT EXISTS idx_reservations_order_status ON reservations(order_id, status);

-- Composite index for product reservations
CREATE INDEX IF NOT EXISTS idx_reservations_product_status ON reservations(product_id, status);

-- Partial index for active reservations
CREATE INDEX IF NOT EXISTS idx_reservations_active ON reservations(product_id, quantity, expires_at) 
WHERE status = 'PENDING';

-- Index for expired reservations cleanup
CREATE INDEX IF NOT EXISTS idx_reservations_expired ON reservations(expires_at, status) 
WHERE status = 'PENDING' AND expires_at < NOW();

-- Composite index for reservation expiry tracking
CREATE INDEX IF NOT EXISTS idx_reservations_expires_status ON reservations(expires_at, status);

-- Index for reservation quantity analysis
CREATE INDEX IF NOT EXISTS idx_reservations_quantity ON reservations(quantity, status);

-- Composite index for product reservation history
CREATE INDEX IF NOT EXISTS idx_reservations_product_created ON reservations(product_id, created_at DESC);

-- Composite index for order reservation tracking
CREATE INDEX IF NOT EXISTS idx_reservations_order_created ON reservations(order_id, created_at DESC);

-- Partial index for confirmed reservations
CREATE INDEX IF NOT EXISTS idx_reservations_confirmed ON reservations(product_id, quantity) 
WHERE status = 'CONFIRMED';

-- Partial index for cancelled reservations
CREATE INDEX IF NOT EXISTS idx_reservations_cancelled ON reservations(product_id, created_at DESC) 
WHERE status = 'CANCELLED';

-- ===== COMPOSITE INDEXES FOR COMPLEX QUERIES =====

-- Index for warehouse stock availability
CREATE INDEX IF NOT EXISTS idx_stocks_warehouse_product_available ON stocks(warehouse_id, product_id, available);

-- Index for product reservation with stock check
CREATE INDEX IF NOT EXISTS idx_stocks_product_available_reserved ON stocks(product_id, available, reserved);

-- Index for movement tracking with references
CREATE INDEX IF NOT EXISTS idx_stock_movements_ref_type_id ON stock_movements(reference_type, reference_id, created_at DESC);

-- ===== REPORTING & ANALYTICS INDEXES =====

-- Index for daily stock reports
CREATE INDEX IF NOT EXISTS idx_stock_movements_daily_report ON stock_movements(DATE(created_at), movement_type, quantity);

-- Index for warehouse utilization reports
CREATE INDEX IF NOT EXISTS idx_stocks_warehouse_total ON stocks(warehouse_id, total);

-- Covering index for stock summary queries
CREATE INDEX IF NOT EXISTS idx_stocks_summary ON stocks(product_id, warehouse_id) 
INCLUDE (available, reserved, total, updated_at);

-- Index for reservation fulfillment rate
CREATE INDEX IF NOT EXISTS idx_reservations_fulfillment ON reservations(status, created_at DESC, quantity);

-- Add comments for documentation
COMMENT ON INDEX idx_stocks_available IS 'Optimizes available stock queries';
COMMENT ON INDEX idx_stocks_low_stock IS 'Fast lookup for low stock alerts';
COMMENT ON INDEX idx_stocks_out_of_stock IS 'Partial index for out-of-stock items';
COMMENT ON INDEX idx_reservations_active IS 'Partial index for active reservations';
COMMENT ON INDEX idx_reservations_expired IS 'Fast cleanup of expired reservations';
COMMENT ON INDEX idx_stock_movements_product_created IS 'Optimizes product movement history';
COMMENT ON INDEX idx_stocks_summary IS 'Covering index for stock queries (reduces heap access)';
COMMENT ON INDEX idx_reservations_confirmed IS 'Fast lookup for confirmed reservations';
