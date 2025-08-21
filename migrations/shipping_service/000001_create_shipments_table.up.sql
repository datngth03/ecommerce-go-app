-- migrations/shipping_service/V<timestamp>_create_shipments_table.up.sql

-- Tạo bảng shipments
CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    provider_id UUID NOT NULL,
    tracking_number VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',
    shipping_fee DECIMAL(12,2) DEFAULT 0,
    shipping_address JSONB NOT NULL,
    estimated_delivery TIMESTAMP WITH TIME ZONE,
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index để tìm kiếm theo mã đơn hàng và mã vận đơn
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON shipments (order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_tracking_number ON shipments (tracking_number);



-- Tạo bảng shipping_providers
-- Bảng này lưu thông tin về các nhà cung cấp dịch vụ vận chuyển
CREATE TABLE IF NOT EXISTS shipping_providers (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    contact VARCHAR(100),
    website VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho các cột name và code để tìm kiếm nhanh hơn
CREATE INDEX IF NOT EXISTS idx_shipping_providers_name ON shipping_providers (name);
CREATE INDEX IF NOT EXISTS idx_shipping_providers_code ON shipping_providers (code);