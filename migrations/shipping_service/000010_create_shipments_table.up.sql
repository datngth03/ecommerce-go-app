-- migrations/shipping_service/V<timestamp>_create_shipments_table.up.sql

-- Tạo bảng shipments
CREATE TABLE IF NOT EXISTS shipments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL, -- ID đơn hàng liên quan
    user_id UUID NOT NULL, -- ID người dùng liên quan đến đơn hàng
    shipping_cost NUMERIC(10, 2) NOT NULL,
    tracking_number VARCHAR(255) UNIQUE, -- Mã theo dõi từ nhà vận chuyển (có thể null ban đầu)
    carrier VARCHAR(100) NOT NULL, -- Nhà vận chuyển (ví dụ: FedEx, UPS)
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- Trạng thái vận chuyển: pending, in_transit, delivered, failed
    shipping_address TEXT NOT NULL, -- Địa chỉ giao hàng (có thể denormalized từ Order Service)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tạo index cho order_id, user_id và tracking_number để tìm kiếm/lọc nhanh hơn
CREATE INDEX IF NOT EXISTS idx_shipments_order_id ON shipments (order_id);
CREATE INDEX IF NOT EXISTS idx_shipments_user_id ON shipments (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipments_tracking_number ON shipments (tracking_number) WHERE tracking_number IS NOT NULL;

-- Thêm ràng buộc khóa ngoại đến bảng orders (nếu orders nằm trong cùng DB)
-- Nếu Order Service có DB riêng, ràng buộc này sẽ không tồn tại ở đây
-- và việc kiểm tra tính toàn vẹn sẽ được thực hiện ở tầng application/service.
-- Giả định hiện tại: orders nằm trong cùng ecommerce_core_db
ALTER TABLE shipments
ADD CONSTRAINT fk_order_id
FOREIGN KEY (order_id) REFERENCES orders(id)
ON DELETE RESTRICT; -- Không cho phép xóa đơn hàng nếu có lô hàng liên quan
