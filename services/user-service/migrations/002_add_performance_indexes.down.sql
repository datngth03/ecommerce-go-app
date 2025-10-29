-- Rollback performance indexes for User Service

DROP INDEX IF EXISTS idx_users_name_trgm;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_active_created;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_created_at_desc;
