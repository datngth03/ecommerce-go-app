-- Add performance indexes for User Service
-- These indexes improve query performance on frequently accessed columns

-- Index on email for faster lookups (already exists, but adding comment)
-- CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index on created_at DESC for recent users queries
CREATE INDEX IF NOT EXISTS idx_users_created_at_desc ON users(created_at DESC);

-- Index on is_active for filtering active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Composite index for active users ordered by creation date
CREATE INDEX IF NOT EXISTS idx_users_active_created ON users(is_active, created_at DESC);

-- Index on phone for phone number lookups
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE phone IS NOT NULL;

-- Index on name for search queries (using trigram for partial match)
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_users_name_trgm ON users USING gin(name gin_trgm_ops);

-- Add comment for documentation
COMMENT ON INDEX idx_users_created_at_desc IS 'Optimizes queries ordering by creation date descending';
COMMENT ON INDEX idx_users_is_active IS 'Optimizes queries filtering by active status';
COMMENT ON INDEX idx_users_active_created IS 'Optimizes queries for active users sorted by date';
COMMENT ON INDEX idx_users_phone IS 'Optimizes phone number lookups (partial index for non-null)';
COMMENT ON INDEX idx_users_name_trgm IS 'Enables fast partial text search on user names';
