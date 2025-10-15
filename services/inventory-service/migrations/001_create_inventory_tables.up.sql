-- Create stocks table
CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id VARCHAR(255) UNIQUE NOT NULL,
    available INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    total INTEGER NOT NULL DEFAULT 0,
    warehouse_id VARCHAR(255) NOT NULL DEFAULT 'main',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT available_positive CHECK (available >= 0),
    CONSTRAINT reserved_positive CHECK (reserved >= 0),
    CONSTRAINT total_positive CHECK (total >= 0),
    CONSTRAINT total_equals_sum CHECK (total = available + reserved)
);

-- Create stock_movements table for audit trail
CREATE TABLE IF NOT EXISTS stock_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id VARCHAR(255) NOT NULL,
    movement_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    before_quantity INTEGER NOT NULL,
    after_quantity INTEGER NOT NULL,
    reference_type VARCHAR(50) NOT NULL,
    reference_id VARCHAR(255),
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (product_id) REFERENCES stocks(product_id) ON DELETE CASCADE
);

-- Create reservations table
CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (product_id) REFERENCES stocks(product_id) ON DELETE CASCADE,
    
    CONSTRAINT quantity_positive CHECK (quantity > 0)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_stocks_product_id ON stocks(product_id);
CREATE INDEX IF NOT EXISTS idx_stocks_warehouse_id ON stocks(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_product_id ON stock_movements(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_movements_created_at ON stock_movements(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reservations_order_id ON reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_reservations_product_id ON reservations(product_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires_at ON reservations(expires_at);

-- Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_stocks_updated_at BEFORE UPDATE ON stocks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_reservations_updated_at BEFORE UPDATE ON reservations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
