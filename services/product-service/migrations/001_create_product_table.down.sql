-- Migration: 001_initial_schema.down.sql
-- Description: Rollback initial schema for product service

-- Drop indexes first
DROP INDEX IF EXISTS idx_categories_slug;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_is_active;
DROP INDEX IF EXISTS idx_products_slug;
DROP INDEX IF EXISTS idx_products_category_id;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Drop trigger
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;

-- Drop extension (optional - only if no other tables use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
