-- Drop triggers
DROP TRIGGER IF EXISTS update_payment_methods_updated_at ON payment_methods;
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_payment_methods_user_id;
DROP INDEX IF EXISTS idx_refunds_status;
DROP INDEX IF EXISTS idx_refunds_payment_id;
DROP INDEX IF EXISTS idx_transactions_payment_id;
DROP INDEX IF EXISTS idx_payments_gateway_payment_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_user_id;
DROP INDEX IF EXISTS idx_payments_order_id;

-- Drop tables
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS payments;
