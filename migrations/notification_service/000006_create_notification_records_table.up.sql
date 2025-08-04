-- migrations/notification_service/V<timestamp>_create_notification_records_table.up.sql

-- Tạo bảng notification_records
CREATE TABLE IF NOT EXISTS notification_records (
    id UUID PRIMARY KEY,
    user_id UUID, -- Optional: User associated with the notification
    type VARCHAR(50) NOT NULL, -- e.g., 'email', 'sms', 'push'
    recipient TEXT NOT NULL, -- e.g., email address, phone number, device token
    subject TEXT, -- For emails
    message TEXT NOT NULL, -- Content of the notification
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- e.g., 'pending', 'sent', 'failed'
    error_message TEXT, -- Error message if sending failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho user_id, type, và status để tìm kiếm/lọc nhanh hơn
CREATE INDEX IF NOT EXISTS idx_notification_records_user_id ON notification_records (user_id);
CREATE INDEX IF NOT EXISTS idx_notification_records_type_status ON notification_records (type, status);
