-- Add advanced performance indexes for Payment Service
-- These indexes optimize payment queries, transaction history, and financial reporting

-- ===== PAYMENTS TABLE INDEXES =====

-- Composite index for user payment history
CREATE INDEX IF NOT EXISTS idx_payments_user_created_desc ON payments(user_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Composite index for user payments by status
CREATE INDEX IF NOT EXISTS idx_payments_user_status ON payments(user_id, status) 
WHERE deleted_at IS NULL;

-- Composite index for order payment tracking
CREATE INDEX IF NOT EXISTS idx_payments_order_status ON payments(order_id, status) 
WHERE deleted_at IS NULL;

-- Index for payment amount (financial queries)
CREATE INDEX IF NOT EXISTS idx_payments_amount ON payments(amount) 
WHERE deleted_at IS NULL;

-- Index for payment method analytics
CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(method, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for currency-based queries
CREATE INDEX IF NOT EXISTS idx_payments_currency ON payments(currency, created_at DESC) 
WHERE deleted_at IS NULL;

-- Partial index for successful payments (hot data)
CREATE INDEX IF NOT EXISTS idx_payments_success ON payments(user_id, created_at DESC) 
WHERE status = 'succeeded' AND deleted_at IS NULL;

-- Partial index for failed payments (analysis)
CREATE INDEX IF NOT EXISTS idx_payments_failed ON payments(user_id, created_at DESC, failure_reason) 
WHERE status = 'failed' AND deleted_at IS NULL;

-- Partial index for pending payments (requires action)
CREATE INDEX IF NOT EXISTS idx_payments_pending ON payments(order_id, created_at DESC) 
WHERE status = 'pending' AND deleted_at IS NULL;

-- Index for daily payment reports
CREATE INDEX IF NOT EXISTS idx_payments_date_status ON payments(DATE(created_at), status, amount) 
WHERE deleted_at IS NULL;

-- Composite index for gateway reconciliation
CREATE INDEX IF NOT EXISTS idx_payments_gateway_customer ON payments(gateway_customer_id, gateway_payment_id) 
WHERE deleted_at IS NULL;

-- GIN index for metadata queries (JSON search)
CREATE INDEX IF NOT EXISTS idx_payments_metadata ON payments USING gin(metadata) 
WHERE deleted_at IS NULL;

-- ===== TRANSACTIONS TABLE INDEXES =====

-- Composite index for payment transactions
CREATE INDEX IF NOT EXISTS idx_transactions_payment_type ON transactions(payment_id, transaction_type) 
WHERE deleted_at IS NULL;

-- Index for transaction status tracking
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for transaction amount analysis
CREATE INDEX IF NOT EXISTS idx_transactions_amount ON transactions(amount, created_at DESC) 
WHERE deleted_at IS NULL;

-- Composite index for payment transaction history
CREATE INDEX IF NOT EXISTS idx_transactions_payment_created ON transactions(payment_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- GIN index for gateway response queries
CREATE INDEX IF NOT EXISTS idx_transactions_gateway_response ON transactions USING gin(gateway_response) 
WHERE deleted_at IS NULL;

-- ===== REFUNDS TABLE INDEXES =====

-- Composite index for payment refunds
CREATE INDEX IF NOT EXISTS idx_refunds_payment_created ON refunds(payment_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Composite index for refund status tracking
CREATE INDEX IF NOT EXISTS idx_refunds_status_created ON refunds(status, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for refund amount analysis
CREATE INDEX IF NOT EXISTS idx_refunds_amount ON refunds(amount) 
WHERE deleted_at IS NULL;

-- Index for gateway refund reconciliation
CREATE INDEX IF NOT EXISTS idx_refunds_gateway_id ON refunds(gateway_refund_id) 
WHERE deleted_at IS NULL;

-- Partial index for pending refunds
CREATE INDEX IF NOT EXISTS idx_refunds_pending ON refunds(payment_id, created_at DESC) 
WHERE status = 'pending' AND deleted_at IS NULL;

-- Index for daily refund reports
CREATE INDEX IF NOT EXISTS idx_refunds_date_status ON refunds(DATE(created_at), status, amount) 
WHERE deleted_at IS NULL;

-- ===== PAYMENT_METHODS TABLE INDEXES =====

-- Composite index for user payment methods
CREATE INDEX IF NOT EXISTS idx_payment_methods_user_type ON payment_methods(user_id, method_type) 
WHERE deleted_at IS NULL;

-- Index for default payment method lookup
CREATE INDEX IF NOT EXISTS idx_payment_methods_default ON payment_methods(user_id, is_default) 
WHERE is_default = true AND deleted_at IS NULL;

-- Index for gateway method reconciliation
CREATE INDEX IF NOT EXISTS idx_payment_methods_gateway ON payment_methods(gateway_method_id) 
WHERE deleted_at IS NULL;

-- Index for method brand analytics
CREATE INDEX IF NOT EXISTS idx_payment_methods_brand ON payment_methods(brand, method_type) 
WHERE deleted_at IS NULL;

-- ===== REPORTING & ANALYTICS INDEXES =====

-- Covering index for payment summaries (includes commonly selected columns)
CREATE INDEX IF NOT EXISTS idx_payments_summary ON payments(user_id, status, created_at DESC) 
INCLUDE (order_id, amount, currency, method) 
WHERE deleted_at IS NULL;

-- Index for revenue analytics by method
CREATE INDEX IF NOT EXISTS idx_payments_method_amount_date ON payments(method, DATE(created_at), amount) 
WHERE status = 'succeeded' AND deleted_at IS NULL;

-- Index for failed payment analysis
CREATE INDEX IF NOT EXISTS idx_payments_failed_analysis ON payments(method, failure_reason, created_at DESC) 
WHERE status = 'failed' AND deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON INDEX idx_payments_user_created_desc IS 'Optimizes user payment history queries';
COMMENT ON INDEX idx_payments_success IS 'Partial index for successful payments only';
COMMENT ON INDEX idx_payments_failed IS 'Partial index for failed payment analysis';
COMMENT ON INDEX idx_payments_pending IS 'Partial index for pending payment tracking';
COMMENT ON INDEX idx_payments_metadata IS 'Enables JSON search on payment metadata';
COMMENT ON INDEX idx_transactions_gateway_response IS 'Enables JSON search on gateway responses';
COMMENT ON INDEX idx_payments_summary IS 'Covering index for payment list queries';
COMMENT ON INDEX idx_payment_methods_default IS 'Fast lookup for default payment method';
COMMENT ON INDEX idx_refunds_pending IS 'Fast tracking of pending refunds';
