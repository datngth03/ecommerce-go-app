-- Add advanced performance indexes for Notification Service
-- These indexes optimize notification queries, delivery tracking, and template management

-- ===== NOTIFICATIONS TABLE INDEXES =====

-- Composite index for user notification history
CREATE INDEX IF NOT EXISTS idx_notifications_user_created_desc ON notifications(user_id, created_at DESC) 
WHERE deleted_at IS NULL;

-- Composite index for user notifications by status
CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status) 
WHERE deleted_at IS NULL;

-- Composite index for user notifications by type
CREATE INDEX IF NOT EXISTS idx_notifications_user_type ON notifications(user_id, type, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for notification channel analytics
CREATE INDEX IF NOT EXISTS idx_notifications_channel ON notifications(channel, status, created_at DESC) 
WHERE deleted_at IS NULL;

-- Composite index for channel and type filtering
CREATE INDEX IF NOT EXISTS idx_notifications_channel_type ON notifications(channel, type, status) 
WHERE deleted_at IS NULL;

-- Index for recipient lookup
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient, status) 
WHERE deleted_at IS NULL;

-- Partial index for pending notifications (requires processing)
CREATE INDEX IF NOT EXISTS idx_notifications_pending ON notifications(created_at ASC) 
WHERE status = 'PENDING' AND deleted_at IS NULL;

-- Partial index for failed notifications (retry queue)
CREATE INDEX IF NOT EXISTS idx_notifications_failed ON notifications(created_at DESC, error_message) 
WHERE status = 'FAILED' AND deleted_at IS NULL;

-- Index for sent notifications tracking
CREATE INDEX IF NOT EXISTS idx_notifications_sent ON notifications(sent_at DESC) 
WHERE status = 'SENT' AND deleted_at IS NULL;

-- Index for notification templates
CREATE INDEX IF NOT EXISTS idx_notifications_template_id ON notifications(template_id) 
WHERE deleted_at IS NULL;

-- GIN index for metadata queries (JSON search)
CREATE INDEX IF NOT EXISTS idx_notifications_metadata ON notifications USING gin(metadata) 
WHERE deleted_at IS NULL;

-- Index for notification type analytics
CREATE INDEX IF NOT EXISTS idx_notifications_type_status ON notifications(type, status, created_at DESC) 
WHERE deleted_at IS NULL;

-- ===== TEMPLATES TABLE INDEXES =====

-- Index for active templates
CREATE INDEX IF NOT EXISTS idx_templates_active ON templates(is_active, type) 
WHERE is_active = true AND deleted_at IS NULL;

-- Composite index for template type filtering
CREATE INDEX IF NOT EXISTS idx_templates_type_active ON templates(type, is_active) 
WHERE deleted_at IS NULL;

-- Index for template name search
CREATE INDEX IF NOT EXISTS idx_templates_name_active ON templates(name, is_active) 
WHERE deleted_at IS NULL;

-- GIN index for variables queries
CREATE INDEX IF NOT EXISTS idx_templates_variables ON templates USING gin(variables) 
WHERE deleted_at IS NULL;

-- Index for template updates
CREATE INDEX IF NOT EXISTS idx_templates_updated_at ON templates(updated_at DESC) 
WHERE deleted_at IS NULL;

-- ===== REPORTING & ANALYTICS INDEXES =====

-- Index for daily notification reports
CREATE INDEX IF NOT EXISTS idx_notifications_daily_report ON notifications(DATE(created_at), type, status) 
WHERE deleted_at IS NULL;

-- Index for channel performance metrics
CREATE INDEX IF NOT EXISTS idx_notifications_channel_metrics ON notifications(channel, status, DATE(created_at)) 
WHERE deleted_at IS NULL;

-- Index for delivery rate analysis
CREATE INDEX IF NOT EXISTS idx_notifications_delivery_rate ON notifications(status, sent_at, created_at) 
WHERE deleted_at IS NULL;

-- Covering index for notification list queries
CREATE INDEX IF NOT EXISTS idx_notifications_summary ON notifications(user_id, status, created_at DESC) 
INCLUDE (type, channel, subject, sent_at) 
WHERE deleted_at IS NULL;

-- Index for template usage analytics
CREATE INDEX IF NOT EXISTS idx_notifications_template_usage ON notifications(template_id, status, created_at DESC) 
WHERE deleted_at IS NULL;

-- Index for error tracking and analysis
CREATE INDEX IF NOT EXISTS idx_notifications_errors ON notifications(status, error_message, created_at DESC) 
WHERE status = 'FAILED' AND deleted_at IS NULL;

-- Index for notification retry logic
CREATE INDEX IF NOT EXISTS idx_notifications_retry ON notifications(id, status, created_at) 
WHERE status IN ('FAILED', 'PENDING') AND deleted_at IS NULL;

-- Index for bulk notification tracking
CREATE INDEX IF NOT EXISTS idx_notifications_bulk ON notifications(type, created_at, status) 
WHERE type LIKE '%BULK%' AND deleted_at IS NULL;

-- ===== TEMPORAL QUERIES INDEXES =====

-- Index for sent_at queries (delivery time analysis)
CREATE INDEX IF NOT EXISTS idx_notifications_sent_at_desc ON notifications(sent_at DESC) 
WHERE sent_at IS NOT NULL AND deleted_at IS NULL;

-- Index for processing time calculation
CREATE INDEX IF NOT EXISTS idx_notifications_processing_time ON notifications(created_at, sent_at) 
WHERE sent_at IS NOT NULL AND deleted_at IS NULL;

-- Index for undelivered notifications (sent_at is null but should be sent)
CREATE INDEX IF NOT EXISTS idx_notifications_undelivered ON notifications(created_at ASC) 
WHERE status = 'SENT' AND sent_at IS NULL AND deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON INDEX idx_notifications_user_created_desc IS 'Optimizes user notification history queries';
COMMENT ON INDEX idx_notifications_pending IS 'Fast lookup for pending notifications';
COMMENT ON INDEX idx_notifications_failed IS 'Fast lookup for failed notifications needing retry';
COMMENT ON INDEX idx_notifications_metadata IS 'Enables JSON search on notification metadata';
COMMENT ON INDEX idx_templates_variables IS 'Enables JSON search on template variables';
COMMENT ON INDEX idx_notifications_summary IS 'Covering index for notification list (reduces heap access)';
COMMENT ON INDEX idx_notifications_retry IS 'Optimizes notification retry queue';
COMMENT ON INDEX idx_notifications_processing_time IS 'Enables delivery time analytics';
