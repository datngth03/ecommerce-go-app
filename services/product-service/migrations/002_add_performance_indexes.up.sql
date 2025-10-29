-- Add advanced performance indexes for Product Service
-- These indexes optimize complex queries and improve search performance

-- ===== PRODUCTS TABLE INDEXES =====

-- Composite index for active products by category (most common query)
CREATE INDEX IF NOT EXISTS idx_products_category_active ON products(category_id, is_active) 
WHERE is_active = true;

-- Index for price range queries
CREATE INDEX IF NOT EXISTS idx_products_price ON products(price);

-- Composite index for price filtering within categories
CREATE INDEX IF NOT EXISTS idx_products_category_price ON products(category_id, price);

-- Index for recent products (created_at DESC for newest first)
CREATE INDEX IF NOT EXISTS idx_products_created_at_desc ON products(created_at DESC);

-- Composite index for active products sorted by date
CREATE INDEX IF NOT EXISTS idx_products_active_created_desc ON products(is_active, created_at DESC) 
WHERE is_active = true;

-- Full-text search index on product name and description
CREATE INDEX IF NOT EXISTS idx_products_name_search ON products USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_desc_search ON products USING gin(to_tsvector('english', coalesce(description, '')));

-- Trigram index for partial name matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_products_name_trgm ON products USING gin(name gin_trgm_ops);

-- ===== CATEGORIES TABLE INDEXES =====

-- Index on category name for search
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);

-- Index on created_at for temporal queries
CREATE INDEX IF NOT EXISTS idx_categories_created_at_desc ON categories(created_at DESC);

-- ===== COMPOSITE INDEXES FOR COMPLEX QUERIES =====

-- Products by category, active, and price range (e-commerce filtering)
CREATE INDEX IF NOT EXISTS idx_products_category_active_price ON products(category_id, is_active, price) 
WHERE is_active = true;

-- Products sorted by price within category (ascending for low-to-high)
CREATE INDEX IF NOT EXISTS idx_products_category_price_asc ON products(category_id, price ASC, created_at DESC) 
WHERE is_active = true;

-- Products sorted by price within category (descending for high-to-low)
CREATE INDEX IF NOT EXISTS idx_products_category_price_desc ON products(category_id, price DESC, created_at DESC) 
WHERE is_active = true;

-- Add comments for documentation
COMMENT ON INDEX idx_products_category_active IS 'Optimizes queries for active products by category';
COMMENT ON INDEX idx_products_price IS 'Optimizes price range filtering';
COMMENT ON INDEX idx_products_category_price IS 'Optimizes price filtering within categories';
COMMENT ON INDEX idx_products_created_at_desc IS 'Optimizes queries for newest products';
COMMENT ON INDEX idx_products_name_search IS 'Enables full-text search on product names';
COMMENT ON INDEX idx_products_desc_search IS 'Enables full-text search on descriptions';
COMMENT ON INDEX idx_products_name_trgm IS 'Enables partial/fuzzy name matching';
COMMENT ON INDEX idx_products_category_active_price IS 'Optimizes filtered product browsing';
