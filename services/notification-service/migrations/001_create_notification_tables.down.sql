-- Drop trigger
DROP TRIGGER IF EXISTS update_templates_updated_at ON templates;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_templates_type;
DROP INDEX IF EXISTS idx_templates_name;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_type;
DROP INDEX IF EXISTS idx_notifications_user_id;

-- Drop tables
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS notifications;
