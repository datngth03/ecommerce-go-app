-- Rollback performance indexes for Product Service

DROP INDEX IF EXISTS idx_products_category_price_desc;
DROP INDEX IF EXISTS idx_products_category_price_asc;
DROP INDEX IF EXISTS idx_products_category_active_price;
DROP INDEX IF EXISTS idx_categories_created_at_desc;
DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_products_name_trgm;
DROP INDEX IF EXISTS idx_products_desc_search;
DROP INDEX IF EXISTS idx_products_name_search;
DROP INDEX IF EXISTS idx_products_active_created_desc;
DROP INDEX IF EXISTS idx_products_created_at_desc;
DROP INDEX IF EXISTS idx_products_category_price;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_category_active;
